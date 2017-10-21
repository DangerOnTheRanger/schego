package schego

import (
	"bytes"
	"encoding/binary"
	"io"
	"strconv"
)

// interface to write a null-terminated string to stdout
type VMConsole interface {
	Write(string)
}

// data structure to contain the stack for a single VM instance
type VMStack struct {
	byteStack []byte
	len       int
}

func (s *VMStack) PushByte(newValue byte) {
	s.byteStack = append(s.byteStack, newValue)
	s.len += 1
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
}

func (s *VMStack) PopInt() int64 {
	intBuffer := make([]byte, 8)
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
}

func (s *VMStack) PopDouble() float64 {
	doubleBuffer := make([]byte, 8)
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
	bufferLength := int64(len(runeBytes))
	binary.Write(stringLength, binary.LittleEndian, bufferLength)
	s.PushInt(stringLength.Bytes())
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

func (s VMStack) Length() int {
	return s.len
}

type VMState struct {
	Stack        VMStack
	Console      VMConsole
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
	case 0x2C:
		// jmp
		addressBytes := v.ReadBytes(8)
		var address int64
		binary.Read(bytes.NewBuffer(addressBytes), binary.LittleEndian, &address)
		v.opcodeBuffer.Seek(address, io.SeekCurrent)
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
	return vm
}

func RunVM(opcodes []byte, console VMConsole) int64 {
	vm := NewVM(opcodes, console)
	for vm.CanStep() {
		vm.Step()
	}
	return vm.ExitCode()
}
