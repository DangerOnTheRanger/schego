package schego

import (
	"testing"
)

func checkProgram(program *Program, expectedProgram *Program, t *testing.T) {
	for i, subNode := range program.GetSubNodes() {
		expectedSubNode := expectedProgram.subNodes[i]
		if expectedSubNode.GetType() != subNode.GetType() {
			t.Error("Parser type error, expected:\n" + expectedSubNode.DebugString() + "\nGot:\n" + subNode.DebugString())
		} else if expectedSubNode.DebugString() != subNode.DebugString() {
			t.Error("Parser value error, expected:\n" + expectedSubNode.DebugString() + "\nGot:\n" + subNode.DebugString())
		}
	}
}

func TestParseSingleExp(t *testing.T) {
	tokens := LexExp("(+ 5 3)")
	program := ParseTokens(tokens)
	expectedProgram := NewProgram(NewAddExp(NewIntLiteral(5), NewIntLiteral(3)))
	checkProgram(program, expectedProgram, t)
}

func TestNestedExp(t *testing.T) {
	tokens := LexExp("(+ (+ 8 (+ 5 6)) 52)")
	program := ParseTokens(tokens)
	expectedProgram := NewProgram(NewAddExp(NewAddExp(NewIntLiteral(8), NewAddExp(NewIntLiteral(5), NewIntLiteral(6))), NewIntLiteral(52)))
	checkProgram(program, expectedProgram, t)
}

func TestMultipleExp(t *testing.T) {
	tokens := LexExp("(+ 3 4)\n(+ 5 6)")
	program := ParseTokens(tokens)
	expectedProgram := NewProgram(NewAddExp(NewIntLiteral(3), NewIntLiteral(4)), NewAddExp(NewIntLiteral(5), NewIntLiteral(6)))
	checkProgram(program, expectedProgram, t)
}
