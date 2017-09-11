package schego

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math"
	"testing"
)

// convenience functions to grab the value of the underlying byte buffer of a token
func getProperValue(input *Token) string {
	if input.Type == TokenIntLiteral {
		convertedNum := getIntValue(input)
		return fmt.Sprint(convertedNum)
	} else if input.Type == TokenFloatLiteral {
		convertedNum := getFloatValue(input)
		return fmt.Sprint(convertedNum)
	} else {
		return getStringValue(input)
	}
}
func getStringValue(input *Token) string {
	return input.Value.String()
}
func getIntValue(input *Token) int64 {
	num, _ := binary.Varint(input.Value.Bytes())
	return num
}
func getFloatValue(input *Token) float64 {
	bits := binary.LittleEndian.Uint64(input.Value.Bytes())
	return math.Float64frombits(bits)
}

func NewTokenNum(tokenType TokenType, tokenString string) *Token {
	token := NewTokenString(tokenType, tokenString)
	token.Value = *bufferStringToNum(tokenType, token.Value)
	return token
}

func equalTokens(first *Token, second *Token) bool {
	if first.Type != second.Type {
		return false
	} else if getProperValue(first) == getProperValue(second) {
		return true
	} else {
		return false
	}
}

func checkTokens(tokens []*Token, expectedTokens []*Token, t *testing.T) {
	for index, token := range tokens {
		expected := expectedTokens[index]
		if equalTokens(token, expected) == false {
			t.Error("Incorrect token, got",
				getProperValue(token),
				"expected",
				getProperValue(expected))
		}
	}
}

// test lexing a single s-expression
func TestLexSingleExp(t *testing.T) {
	tokens := LexExp("(abc def ghi)")
	expectedTokens := []*Token{
		NewTokenString(TokenLParen, "("),
		NewTokenString(TokenIdent, "abc"),
		NewTokenString(TokenIdent, "def"),
		NewTokenString(TokenIdent, "ghi"),
		NewTokenString(TokenRParen, ")")}
	checkTokens(tokens, expectedTokens, t)
}

// test lexing nested s-expressions
func TestLexNestedExp(t *testing.T) {
	tokens := LexExp("(abc (def ghi (jkl)))")
	expectedTokens := []*Token{
		NewTokenString(TokenLParen, "("),
		NewTokenString(TokenIdent, "abc"),
		NewTokenString(TokenLParen, "("),
		NewTokenString(TokenIdent, "def"),
		NewTokenString(TokenIdent, "ghi"),
		NewTokenString(TokenLParen, "("),
		NewTokenString(TokenIdent, "jkl"),
		NewTokenString(TokenRParen, ")"),
		NewTokenString(TokenRParen, ")"),
		NewTokenString(TokenRParen, ")")}
	checkTokens(tokens, expectedTokens, t)
}

// test to make sure whitespace is properly ignored
func TestIgnoreExtraWhitespace(t *testing.T) {
	tokens := LexExp("( ab   cd efg)")
	expectedTokens := []*Token{
		NewTokenString(TokenLParen, "("),
		NewTokenString(TokenIdent, "ab"),
		NewTokenString(TokenIdent, "cd"),
		NewTokenString(TokenIdent, "efg"),
		NewTokenString(TokenRParen, ")")}
	checkTokens(tokens, expectedTokens, t)
}

// test corner case where the is no closing rparen and only EOL/EOF
func TestLexIdentCorner(t *testing.T) {
	tokens := LexExp("(abc")
	expectedTokens := []*Token{
		NewTokenString(TokenLParen, "("),
		NewTokenString(TokenIdent, "abc")}
	checkTokens(tokens, expectedTokens, t)
}

// test number literals
func TestLexNumberLiterals(t *testing.T) {
	tokens := LexExp("(123 abc def 456.789 .012 345)")
	expectedTokens := []*Token{
		NewTokenString(TokenLParen, "("),
		NewTokenNum(TokenIntLiteral, "123"),
		NewTokenString(TokenIdent, "abc"),
		NewTokenString(TokenIdent, "def"),
		NewTokenNum(TokenFloatLiteral, "456.789"),
		NewTokenNum(TokenFloatLiteral, ".012"),
		NewTokenNum(TokenIntLiteral, "345"),
		NewTokenString(TokenRParen, ")")}
	checkTokens(tokens, expectedTokens, t)
}

// test special characters
func TestIdentSpecial(t *testing.T) {
	tokens := LexExp("ab.c . d|ef? |gh +i|")
	expectedTokens := []*Token{
		NewTokenString(TokenIdent, "ab.c"),
		NewTokenString(TokenDot, "."),
		NewTokenString(TokenIdent, "d|ef?"),
		NewTokenString(TokenIdent, "|gh +i|")}
	checkTokens(tokens, expectedTokens, t)
}

// test operators
func TestOps(t *testing.T) {
	tokens := LexExp("(>= 150 (* (+ 10 3.2) 5))")
	expectedTokens := []*Token{
		NewTokenString(TokenLParen, "("),
		NewTokenString(TokenOp, ">="),
		NewTokenNum(TokenIntLiteral, "150"),
		NewTokenString(TokenLParen, "("),
		NewTokenString(TokenOp, "*"),
		NewTokenString(TokenLParen, "("),
		NewTokenString(TokenOp, "+"),
		NewTokenNum(TokenIntLiteral, "10"),
		NewTokenNum(TokenFloatLiteral, "3.2"),
		NewTokenString(TokenRParen, ")"),
		NewTokenNum(TokenIntLiteral, "5"),
		NewTokenString(TokenRParen, ")"),
		NewTokenString(TokenRParen, ")")}
	checkTokens(tokens, expectedTokens, t)
}

// test newline
func TestNewline(t *testing.T) {
	tokens := LexExp("(ab\ncd\nef)")
	expectedTokens := []*Token{
		NewTokenString(TokenLParen, "("),
		NewTokenString(TokenIdent, "ab"),
		NewTokenString(TokenIdent, "cd"),
		NewTokenString(TokenIdent, "ef"),
		NewTokenString(TokenRParen, ")")}
	checkTokens(tokens, expectedTokens, t)
}

// test to make sure single-constant expressions are lexed/converted correctly
func TestSingleFloat(t *testing.T) {
	tokens := LexExp("3.14")
	expectedTokens := []*Token{NewTokenNum(TokenFloatLiteral, "3.14")}
	checkTokens(tokens, expectedTokens, t)
}

// test string literals
func TestString(t *testing.T) {
	tokens := LexExp("\"la li lu le lo\"")
	expectedTokens := []*Token{NewTokenString(TokenStringLiteral, "la li lu le lo")}
	checkTokens(tokens, expectedTokens, t)
}

// test bool literals
func TestBool(t *testing.T) {
	tokens := LexExp("#t #f bla")
	expectedTokens := []*Token{
		NewTokenRaw(TokenBoolLiteral, *bytes.NewBuffer([]byte{1})),
		NewTokenRaw(TokenBoolLiteral, *bytes.NewBuffer([]byte{0})),
		NewTokenString(TokenIdent, "bla")}
	checkTokens(tokens, expectedTokens, t)
}
