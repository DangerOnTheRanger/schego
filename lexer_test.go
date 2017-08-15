package schego

import (
	"testing"
)

// convenience function to grab the value of the underlying byte buffer of a token
func getStringValue(input *Token) string {
	return input.Value.String()
}

func compareTokens(first *Token, second *Token) bool {
	if first.Type != second.Type {
		return false
	} else if getStringValue(first) != getStringValue(second) {
		return false
	} else {
		return true
	}
}

func checkTokens(tokens []*Token, expectedTokens []*Token, t *testing.T) {
	for index, token := range tokens {
		expected := expectedTokens[index]
		if compareTokens(token, expected) == false {
			t.Error("Incorrect token, got",
				getStringValue(token), "expected",
				getStringValue(expected))
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
	tokens := LexExp("(123 abc def 456.789 .012)")
	expectedTokens := []*Token{
		NewTokenString(TokenLParen, "("),
		NewTokenString(TokenIntLiteral, "123"),
		NewTokenString(TokenIdent, "abc"),
		NewTokenString(TokenIdent, "def"),
		NewTokenString(TokenFloatLiteral, "456.789"),
		NewTokenString(TokenFloatLiteral, ".012"),
		NewTokenString(TokenRParen, ")")}
	checkTokens(tokens, expectedTokens, t)
}

// test special characters
func TestIdentSpecial(t *testing.T) {
	tokens := LexExp("ab.c . def?")
	expectedTokens := []*Token{
		NewTokenString(TokenIdent, "ab.c"),
		NewTokenString(TokenDot, "."),
		NewTokenString(TokenIdent, "def?")}
	checkTokens(tokens, expectedTokens, t)
}
