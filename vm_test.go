package schego

import (
	"strings"
	"testing"
)

type DummyConsole struct {
	consoleOutput string
}

func (d *DummyConsole) Write(line string) {
	// trim null
	d.consoleOutput = strings.TrimRight(line, "\x00")
}

func TestHelloWorld(t *testing.T) {
	opcodes := []byte{
		0x05, // pushs
		0x48, // H
		0x65, // e
		0x6C, // l
		0x6C, // l
		0x6F, // o
		0x2C, // ,
		0x20, // space
		0x57, // W
		0x6F, // o
		0x72, // r
		0x6C, // l
		0x64, // d
		0x21, // !
		0x0A, // \n
		0x00, // null
		0x43, // syscall
		0x05, // print string
		0x03, // pushi
		0x00,
		0x00,
		0x00,
		0x00,
		0x00,
		0x00,
		0x00,
		0x00, // 0
		0x43, // syscall
		0x06, // exit
	}
	console := DummyConsole{}
	retcode := RunVM(opcodes, &console)
	if retcode != 0 {
		t.Error("Expected return code of 0, got:\n", retcode)
	}
	if console.consoleOutput != "Hello, World!\n" {
		t.Error("Incorrect output, got: ", console.consoleOutput)
	}
}

func TestHelloUnicode(t *testing.T) {
	opcodes := []byte{
		0x05, // pushs
		0xE3,
		0x81,
		0x93, // こ
		0xE3,
		0x82,
		0x93, // ん
		0xE3,
		0x81,
		0xAB, // に
		0xE3,
		0x81,
		0xA1, // ち
		0xE3,
		0x81,
		0xAF, // は
		0xE4,
		0xB8,
		0x96, // 世
		0xE7,
		0x95,
		0x8C, // 界
		0x21, // !
		0x0A, // \n
		0x00, // null
		0x43, // syscall
		0x05, // print string
		0x03, // pushi
		0x00,
		0x00,
		0x00,
		0x00,
		0x00,
		0x00,
		0x00,
		0x00, // 0
		0x43, // syscall
		0x06, // exit
	}
	console := DummyConsole{}
	RunVM(opcodes, &console)
	if console.consoleOutput != "こんにちは世界!\n" {
		t.Error("Incorrect output, got: ", console.consoleOutput)
	}
}

func TestDouble(t *testing.T) {
	opcodes := []byte{
		0x04, // pushd
		0x18,
		0x2D,
		0x44,
		0x54,
		0xFB,
		0x21,
		0x09,
		0x40, // pi (3.141592653589793)
		0x43, // syscall
		0x04, // print double
		0x03, // pushi
		0x00,
		0x00,
		0x00,
		0x00,
		0x00,
		0x00,
		0x00,
		0x00, // 0
		0x43, // syscall
		0x06, // exit
	}
	console := DummyConsole{}
	RunVM(opcodes, &console)
	if console.consoleOutput != "3.141592653589793" {
		t.Error("Incorrect output, got: ", console.consoleOutput)
	}
}

func TestUnconditionalJump(t *testing.T) {
	opcodes := []byte{
		0x03, // pushi
		0x04,
		0x00,
		0x00,
		0x00,
		0x00,
		0x00,
		0x00,
		0x00, // 4
		0x2C, // jmp
		0x09,
		0x00,
		0x00,
		0x00,
		0x00,
		0x00,
		0x00,
		0x00, // 9 - skip the pushi below
		0x03, // pushi
		0xDE,
		0xAD,
		0xBE,
		0xEF,
		0xDE,
		0xAD,
		0xBE,
		0xEF,
		0x43, // syscall
		0x03, // print integer (the 4 from earlier)
		0x03, // pushi
		0x00,
		0x00,
		0x00,
		0x00,
		0x00,
		0x00,
		0x00,
		0x00, // 0
		0x43, // syscall
		0x06, // exit
	}
	console := DummyConsole{}
	RunVM(opcodes, &console)
	if console.consoleOutput != "4" {
		t.Error("Incorrect output, got: ", console.consoleOutput)
	}
}
