package schego

import (
	"bytes"
	"encoding/binary"
	"io"
	"math"
	"strconv"
)

// interface to write a null-terminated string to stdout
type VMConsole interface {
	Write(string)
}

// data structure to contain the stack for a single VM instance
type VMStack struct {
	byteStack     []byte
	len           uint64
	lenLastPushed uint64
}

func (s *VMStack) PushByte(newValue byte) {
	s.byteStack = append(s.byteStack, newValue)
	s.len += 1
	s.lenLastPushed = 1
}

func (s *VMStack) PopByte() byte {
	top := s.byteStack[s.Length()-1]
	s.byteStack = s.byteStack[:s.Length()-1]
	s.len -= 1
	return top
}

func (s *VMStack) PushInt(intBytes []byte) {
	for _, intByte := range intBytes {
		s.PushByte(intByte)
	}
	s.lenLastPushed = 8
}

func (s *VMStack) PopInt() int64 {
	intBuffer := make([]byte, 0)
	for i := 0; i < 8; i++ {
		// insert at the front to make up for pushing the bytes
		// onto the stack in reverse order
		intBuffer = append([]byte{s.PopByte()}, intBuffer...)
	}
	// explicitly declare num as int64 since the underlying data
	// is a 64-bit integer
	var num int64
	binary.Read(bytes.NewBuffer(intBuffer), binary.LittleEndian, &num)
	return num
}

func (s *VMStack) PushDouble(doubleBytes []byte) {
	for _, doubleByte := range doubleBytes {
		s.PushByte(doubleByte)
	}
	s.lenLastPushed = 8
}

func (s *VMStack) PopDouble() float64 {
	doubleBuffer := make([]byte, 0)
	for i := 0; i < 8; i++ {
		doubleBuffer = append([]byte{s.PopByte()}, doubleBuffer...)
	}
	var num float64
	binary.Read(bytes.NewBuffer(doubleBuffer), binary.LittleEndian, &num)
	return num
}

func (s *VMStack) PushString(runeBytes []byte) {
	for _, runeByte := range runeBytes {
		s.PushByte(runeByte)
	}
	// push the length of the string (in bytes) onto the stack,
	// so we know how many bytes to pop when attempting to utilize it
	stringLength := new(bytes.Buffer)
	bufferLength := uint64(len(runeBytes))
	binary.Write(stringLength, binary.LittleEndian, bufferLength)
	s.PushInt(stringLength.Bytes())
	s.lenLastPushed = bufferLength
}

func (s *VMStack) PopString() []byte {
	// grab the string length off the stack that we pushed earlier
	stringLength := s.PopInt()
	utfBytes := make([]byte, 0)
	for i := int64(0); i < stringLength; i++ {
		utfBytes = append([]byte{s.PopByte()}, utfBytes...)
	}
	return utfBytes
}

func (s *VMStack) PushEmptyCell() {
	// push the equivalent of an empty list cell onto the stack
	zeroBuf := make([]byte, 8)
	s.PushInt(zeroBuf)
	s.PushInt(zeroBuf)
	s.PushInt(zeroBuf)
	// set lenLastPushed for compatibility with dup
	s.lenLastPushed = 24
}

func (s *VMStack) PushCell(cell []byte) {
	nextCell := cell[16:]
	s.PushInt(nextCell)
	dataAddress := cell[8:16]
	s.PushInt(dataAddress)
	dataLength := cell[:8]
	s.PushInt(dataLength)
	s.lenLastPushed = 24
}

func (s *VMStack) PopCell() []uint64 {
	cell := make([]uint64, 0)
	for i := 0; i < 3; i++ {
		cell = append(cell, uint64(s.PopInt()))
	}
	return cell
}

func (s *VMStack) Dup() {
	lastValue := s.byteStack[uint64(s.Length())-s.lenLastPushed:]
	for _, valueByte := range lastValue {
		s.PushByte(valueByte)
	}
	s.lenLastPushed = uint64(len(lastValue))
}

func (s VMStack) Length() uint64 {
	return s.len
}

var initialHeapSize uint64 = 16384
var blockSize uint64 = 32
var maxOrder uint8 = 10

type VMHeap struct {
	heapSpace    []byte
	unusedBlocks map[uint8][]uint64
	blockMap     map[uint64]uint8
}

func NewVMHeap() *VMHeap {
	h := new(VMHeap)
	h.heapSpace = make([]byte, initialHeapSize)
	h.blockMap = make(map[uint64]uint8)
	h.unusedBlocks = make(map[uint8][]uint64)
	for i := uint8(0); i <= maxOrder; i++ {
		h.unusedBlocks[i] = make([]uint64, 0)
	}
	h.AllocateRootBlock(initialHeapSize)
	return h
}

func (h *VMHeap) Allocate(numBytes uint64) uint64 {
	order := h.OrderFor(numBytes)
	if h.NoFreeBlocksFor(order) {
		h.CreateBlock(order)
	}
	blockAddress := h.GetFreeBlock(order)
	// GetFreeBlock always returns the first/0th free block,
	// so remove that one
	h.RemoveBlockFromUnused(0, order)
	return blockAddress
}

func (h *VMHeap) Free(address uint64) {
	order := h.blockMap[address]
	// add the newly freed block back to the list of unused blocks
	// MergeWithBuddy will take care of removing it if need be due to merging
	h.unusedBlocks[order] = append(h.unusedBlocks[order], address)
	if h.HasBuddy(address, order) {
		h.MergeWithBuddy(address, order)
	}
}

func (h *VMHeap) Write(data *bytes.Buffer, address uint64) {
	// making sure that no data is accidentally overwritten is left
	// as an exercise to the caller
	for index, dataByte := range data.Bytes() {
		h.heapSpace[address+uint64(index)] = dataByte
	}
}

func (h *VMHeap) Read(numBytes uint64, address uint64) *bytes.Buffer {
	buffer := new(bytes.Buffer)
	for i := uint64(0); i < numBytes; i++ {
		buffer.WriteByte(h.heapSpace[address+uint64(i)])
	}
	return buffer
}

func (h *VMHeap) ReadString(address uint64) []byte {
	strBytes := make([]byte, 0)
	for i := uint64(0); h.heapSpace[address+i] != 0; i++ {
		strBytes = append(strBytes, h.heapSpace[address+i])
	}
	return strBytes
}

func (h *VMHeap) AllocateRootBlock(heapSize uint64) {
	order := h.OrderFor(heapSize)
	h.unusedBlocks[order] = append(h.unusedBlocks[order], 0)
	h.blockMap[0] = order
}

func (h *VMHeap) OrderFor(requestedBytes uint64) uint8 {
	if(requestedBytes <= blockSize) {
		return 0
	}
	// TODO: handle requests past maxOrder gracefully
	// this is ugly, but kinda fast, and less stupid than the old solution
	var order uint8
	order = uint8(math.Log2(float64(requestedBytes)) - math.Log2(float64(blockSize)))
	return order
}

func (h *VMHeap) NoFreeBlocksFor(order uint8) bool {
	return len(h.unusedBlocks[order]) == 0
}

func (h *VMHeap) CreateBlock(order uint8) {
	// find smallest order that we can pull from
	freeOrder := order + 1
	for {
		if h.NoFreeBlocksFor(freeOrder) {
			freeOrder += 1
		} else {
			break
		}
	}
	// repeatedly split blocks until we get one (technically, two) of the order we originally wanted
	for freeOrder > order {
		blockAddress := h.GetFreeBlock(freeOrder)
		h.SplitBlock(blockAddress, freeOrder)
		freeOrder -= 1
	}
}

func (h *VMHeap) GetFreeBlock(order uint8) uint64 {
	// return the address of the first free block of the given order
	return h.unusedBlocks[order][0]
}

func (h *VMHeap) SplitBlock(address uint64, order uint8) {
	// find and remove block from the unused list, since
	// we're about to split it
	targetIndex := 0
	for index, candidateAddress := range h.unusedBlocks[order] {
		if candidateAddress == address {
			targetIndex = index
			break
		}
	}
	h.RemoveBlockFromUnused(targetIndex, order)
	targetOrder := order - 1
	// calculate offset from the start of the original block
	// adding the second address to the list of unused blocks puts smaller blocks out
	// at the end of the heap
	secondAddress := address + uint64(math.Pow(2, float64(targetOrder))*float64(blockSize))
	h.unusedBlocks[targetOrder] = append(h.unusedBlocks[targetOrder], secondAddress)
	h.blockMap[secondAddress] = targetOrder
	h.unusedBlocks[targetOrder] = append(h.unusedBlocks[targetOrder], address)
	h.blockMap[address] = targetOrder
}

func (h *VMHeap) GetUnusedBlockIndex(address uint64, order uint8) int {
	for index, candidateAddress := range h.unusedBlocks[order] {
		if candidateAddress == address {
			return index
		}
	}
	return -1
}

func (h *VMHeap) RemoveBlockFromUnused(index int, order uint8) {
	h.unusedBlocks[order] = append(h.unusedBlocks[order][:index], h.unusedBlocks[order][index+1:]...)
}

func (h *VMHeap) HasBuddy(address uint64, order uint8) bool {
	buddyAddress := h.GetBuddyAddress(address, order)
	for _, candidateAddress := range h.unusedBlocks[order] {
		if candidateAddress == buddyAddress {
			return true
		}
	}
	return false
}

func (h *VMHeap) GetBuddyAddress(address uint64, order uint8) uint64 {
	// buddy address calculation taken from http://www.cs.uml.edu/~jsmith/OSReport/frames.html
	totalBlockSize := uint64(math.Pow(2, float64(order)) * float64(blockSize))
	buddyNumber := address / totalBlockSize
	var buddyAddress uint64
	if buddyNumber%2 == 0 {
		buddyAddress = address + totalBlockSize
	} else {
		buddyAddress = address - totalBlockSize
	}
	return buddyAddress
}

func (h *VMHeap) MergeWithBuddy(address uint64, order uint8) {
	buddyAddress := h.GetBuddyAddress(address, order)
	// figure out which address is lower and delete the other block
	// take the lower address for the new merged block
	var newAddress uint64
	if buddyAddress < address {
		newAddress = buddyAddress
		delete(h.blockMap, address)
	} else {
		newAddress = address
		delete(h.blockMap, buddyAddress)
	}
	buddyIndex := h.GetUnusedBlockIndex(buddyAddress, order)
	h.RemoveBlockFromUnused(buddyIndex, order)
	blockIndex := h.GetUnusedBlockIndex(address, order)
	h.RemoveBlockFromUnused(blockIndex, order)
	h.blockMap[newAddress] = order + 1
	h.unusedBlocks[order+1] = append(h.unusedBlocks[order+1], newAddress)
	// recurse if we still have potential merging left undone
	if h.HasBuddy(newAddress, order+1) {
		h.MergeWithBuddy(newAddress, order+1)
	}
}

type VMState struct {
	Stack        VMStack
	Heap         VMHeap
	Console      VMConsole
	mnemonicMap  map[string]uint64
	opcodes      []byte
	opcodeBuffer bytes.Reader
	finished     bool
	exitCode     int64
}

func (v *VMState) CanStep() bool {
	return v.opcodeBuffer.Len() != 0 && !v.finished
}

func (v *VMState) NextOpcode() byte {
	byte, err := v.opcodeBuffer.ReadByte()
	if err != nil {
		// 0 is an invalid opcode
		return 0
	}
	return byte
}

func (v *VMState) ReadBytes(length int) []byte {
	byteBuffer := make([]byte, length)
	_, err := v.opcodeBuffer.Read(byteBuffer)
	if err != nil {
		// best we can do for now
		// TODO: better error handling here
		return []byte{0}
	}
	return byteBuffer
}

func (v *VMState) jump() {
	addressBytes := v.ReadBytes(8)
	var address int64
	binary.Read(bytes.NewBuffer(addressBytes), binary.LittleEndian, &address)
	v.opcodeBuffer.Seek(address, io.SeekCurrent)
}

func (v *VMState) Step() {
	if v.CanStep() == false {
		// TODO: properly handle finished VM
		return
	}
	currentOpcode := v.NextOpcode()
	switch currentOpcode {
	case 0x03:
		// pushi
		// simply grab the next 8 bytes and push them
		intBytes := v.ReadBytes(8)
		v.Stack.PushInt(intBytes)
	case 0x04:
		// pushd
		doubleBytes := v.ReadBytes(8)
		v.Stack.PushDouble(doubleBytes)
	case 0x05:
		// pushs
		utfBytes := make([]byte, 0)
		for {
			firstByte := v.ReadBytes(1)[0]
			if firstByte&0x80 == 0 || firstByte == 0 {
				// ASCII or null byte
				utfBytes = append(utfBytes, firstByte)
				// null?
				if firstByte == 0 {
					break
				} else {
					continue
				}
			}
			// with UTF-8, the most significant bits tell us
			// how many bytes to read, so construct some simple bitmasks
			// to check
			// see: https://en.wikipedia.org/wiki/UTF-8#Description
			// essentially, check the upper four bits of the first byte,
			// with 1111 meaning read 4 bytes total, 1110 3 bytes, and
			// 1100, 2 bytes
			var codepointLength int
			upperFourBits := firstByte >> 4
			if upperFourBits == 0xC {
				codepointLength = 2
			} else if upperFourBits == 0xE {
				codepointLength = 3
			} else {
				// naively assume 4-byte codepoint length,
				// could be dangerous, should probably be replaced
				// at some point in the future
				codepointLength = 4
			}
			// we've already read one byte (firstByte)
			numBytes := codepointLength - 1
			extraBytes := v.ReadBytes(numBytes)
			utfRune := append([]byte{firstByte}, extraBytes...)
			// append the finished character
			utfBytes = append(utfBytes, utfRune...)
		}
		v.Stack.PushString(utfBytes)
	case 0x06:
		// cons
		v.Stack.PushEmptyCell()
	case 0x07:
		// dup
		v.Stack.Dup()
	case 0x0A:
		// hstorei
		mnemonic := string(v.ReadBytes(2))
		address := v.mnemonicMap[mnemonic]
		num := v.Stack.PopInt()
		intBuffer := bytes.NewBuffer(make([]byte, 0))
		binary.Write(intBuffer, binary.LittleEndian, &num)
		v.Heap.Write(intBuffer, address)
	case 0x0C:
		// hstores
		mnemonic := string(v.ReadBytes(2))
		address := v.mnemonicMap[mnemonic]
		strBytes := v.Stack.PopString()
		var strBuffer bytes.Buffer
		numBytes, _ := strBuffer.Write(strBytes)
		numBytes64 := uint64(numBytes)
		var allocatedBytes uint64
		binary.Read(v.Heap.Read(8, address), binary.LittleEndian, &allocatedBytes)
		if numBytes64 > allocatedBytes {
			v.Heap.Free(address)
			newAddress := v.Heap.Allocate(8 + numBytes64)
			v.mnemonicMap[mnemonic] = newAddress
			intBuffer := bytes.NewBuffer(make([]byte, 0))
			binary.Write(intBuffer, binary.LittleEndian, &numBytes64)
			v.Heap.Write(intBuffer, newAddress)
			// offset by 8 to avoid writing over the int we just wrote
			v.Heap.Write(&strBuffer, newAddress+8)
		} else {
			v.Heap.Write(&strBuffer, address+8)
		}
	case 0x0D:
		// hstorel
		mnemonic := string(v.ReadBytes(2))
		address := v.mnemonicMap[mnemonic]
		cell := v.Stack.PopCell()
		for i := uint64(0); i < 3; i++ {
			buffer := bytes.NewBuffer(make([]byte, 0))
			binary.Write(buffer, binary.LittleEndian, &cell[i])
			v.Heap.Write(buffer, address+8*i)
		}
	case 0x16:
		// hloadi
		mnemonic := string(v.ReadBytes(2))
		address := v.mnemonicMap[mnemonic]
		buffer := v.Heap.Read(8, address)
		v.Stack.PushInt(buffer.Bytes())
	case 0x18:
		// hloads
		mnemonic := string(v.ReadBytes(2))
		address := v.mnemonicMap[mnemonic]
		// offset by 8 to avoid reading intial int containing storage info
		buffer := v.Heap.ReadString(address + 8)
		v.Stack.PushString(buffer)
	case 0x19:
		// hloadl
		mnemonic := string(v.ReadBytes(2))
		address := v.mnemonicMap[mnemonic]
		buffer := v.Heap.Read(24, address)
		v.Stack.PushCell(buffer.Bytes())
	case 0x22:
		// hnewi
		mnemonic := string(v.ReadBytes(2))
		address := v.Heap.Allocate(8)
		v.mnemonicMap[mnemonic] = address
	case 0x24:
		// hnews
		mnemonic := string(v.ReadBytes(2))
		initialMemory := v.Stack.PopInt()
		// allocate space for an int storing how many bytes was allocated
		// for the string, in addition to the inital memory requested
		address := v.Heap.Allocate(8 + uint64(initialMemory))
		v.mnemonicMap[mnemonic] = address
		// record the amount of string-only memory requested in the heap
		// this is useful if/when we try to resize the string later
		intBuffer := bytes.NewBuffer(make([]byte, 0))
		binary.Write(intBuffer, binary.LittleEndian, &initialMemory)
		v.Heap.Write(intBuffer, address)
	case 0x25:
		// hnewl
		mnemonic := string(v.ReadBytes(2))
		address := v.Heap.Allocate(24)
		v.mnemonicMap[mnemonic] = address
	case 0x2C:
		// jmp
		v.jump()
	case 0x2D:
		// jne
		cmpResult := v.Stack.PopByte()
		if cmpResult != 0 {
			v.jump()
		} else {
			// skip the jump address
			v.opcodeBuffer.Seek(8, io.SeekCurrent)
		}
	case 0x36:
		// addi
		y := v.Stack.PopInt()
		x := v.Stack.PopInt()
		newInt := x + y
		intBuffer := bytes.NewBuffer(make([]byte, 0))
		binary.Write(intBuffer, binary.LittleEndian, &newInt)
		v.Stack.PushInt(intBuffer.Bytes())
	case 0x40:
		// cmpi
		y := v.Stack.PopInt()
		x := v.Stack.PopInt()
		if x == y {
			v.Stack.PushByte(0)
		} else if x > y {
			v.Stack.PushByte(1)
		} else {
			v.Stack.PushByte(2)
		}
	case 0x41:
		// cmpd
		y := v.Stack.PopDouble()
		x := v.Stack.PopDouble()
		if x == y {
			v.Stack.PushByte(0)
		} else if x > y {
			v.Stack.PushByte(1)
		} else {
			v.Stack.PushByte(2)
		}
	case 0x43:
		// syscall
		syscall := v.ReadBytes(1)[0]
		switch syscall {
		case 0x03:
			// print integer
			intNum := v.Stack.PopInt()
			intString := strconv.FormatInt(intNum, 10)
			v.Console.Write(intString)
		case 0x04:
			// print double
			doubleNum := v.Stack.PopDouble()
			doubleString := strconv.FormatFloat(doubleNum, 'f', -1, 64)
			v.Console.Write(doubleString)
		case 0x05:
			// print string
			utfBytes := v.Stack.PopString()
			utfString := string(utfBytes)
			v.Console.Write(utfString)
		case 0x06:
			// exit
			exitCode := v.Stack.PopInt()
			v.exitCode = exitCode
			v.finished = true
		}
	case 0x44:
		// hsmem
		mnemonic := string(v.ReadBytes(2))
		sourceMnemonic := string(v.ReadBytes(2))
		v.mnemonicMap[mnemonic] = v.mnemonicMap[sourceMnemonic]
	case 0x46:
		// cmpl
		firstAddress := v.Stack.PopCell()[1]
		secondAddress := v.Stack.PopCell()[1]
		if firstAddress == secondAddress {
			v.Stack.PushByte(0)
		} else if firstAddress > secondAddress {
			v.Stack.PushByte(1)
		} else {
			v.Stack.PushByte(2)
		}
	case 0x47:
		// hcar
		cell := v.Stack.PopCell()
		numBytes := cell[0]
		dataAddress := cell[1]
		cellData := v.Heap.Read(numBytes, dataAddress).Bytes()
		for i := uint64(0); i < numBytes; i++ {
			v.Stack.PushByte(cellData[numBytes-i-1])
		}
		// hack to get VMStack.lenLastPushed to play nicely with hcar
		// should probably be replaced with a dedicated method on
		// VMStack in the future (PushCellData?)
		v.Stack.lenLastPushed = numBytes
	case 0x49:
		// hcdr
		headCell := v.Stack.PopCell()
		cdrAddress := headCell[2]
		v.Stack.PushCell(v.Heap.Read(24, cdrAddress).Bytes())
	case 0x4B:
		// hscar
		data := make([]byte, 0)
		dataLength := v.Stack.lenLastPushed
		for i := uint64(0); i < dataLength; i++ {
			data = append(data, v.Stack.PopByte())
		}
		cell := v.Stack.PopCell()
		allocatedBytes := cell[0]
		dataAddress := cell[1]
		if dataLength > allocatedBytes {
			v.Heap.Free(dataAddress)
			dataAddress = v.Heap.Allocate(dataLength)
			cell[1] = dataAddress
			cell[0] = dataLength
		}
		v.Heap.Write(bytes.NewBuffer(data), dataAddress)
		cellBuffer := bytes.NewBuffer(make([]byte, 0))
		binary.Write(cellBuffer, binary.LittleEndian, cell)
		v.Stack.PushCell(cellBuffer.Bytes())
	case 0x4D:
		// hscdr
		mnemonic := string(v.ReadBytes(2))
		address := v.mnemonicMap[mnemonic]
		cell := v.Stack.PopCell()
		cell[2] = address
		cellBuffer := bytes.NewBuffer(make([]byte, 0))
		binary.Write(cellBuffer, binary.LittleEndian, cell)
		v.Stack.PushCell(cellBuffer.Bytes())
	}
}

func (v *VMState) ExitCode() int64 {
	return v.exitCode
}

func NewVM(opcodes []byte, console VMConsole) *VMState {
	vm := new(VMState)
	vm.opcodes = opcodes
	vm.opcodeBuffer = *bytes.NewReader(vm.opcodes)
	vm.Console = console
	vm.Heap = *NewVMHeap()
	vm.mnemonicMap = make(map[string]uint64)
	return vm
}

func RunVM(opcodes []byte, console VMConsole) int64 {
	vm := NewVM(opcodes, console)
	for vm.CanStep() {
		vm.Step()
	}
	return vm.ExitCode()
}
