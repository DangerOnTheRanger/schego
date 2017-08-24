package schego

import (
	"bytes"
	"encoding/binary"
	"math"
	"strconv"
	"strings"
	"unicode"
)

type TokenType int

const (
	TokenNone TokenType = iota
	TokenRParen
	TokenLParen
	TokenIdent
	TokenIntLiteral
	TokenFloatLiteral
	TokenDot
	TokenOp
	TokenChar
)

type Token struct {
	Type  TokenType
	Value bytes.Buffer
}

// NewTokenString is a convenience function that returns a token with
// a given string value.
func NewTokenString(tokenType TokenType, tokenString string) *Token {
	tokenValue := bytes.NewBufferString(tokenString)
	token := Token{tokenType, *tokenValue}
	return &token
}

// NewTokenRaw creates a new token from a raw Buffer.
func NewTokenRaw(tokenType TokenType, tokenBuffer bytes.Buffer) *Token {
	tokenSlice := make([]byte, tokenBuffer.Len(), tokenBuffer.Len())
	// copy to avoid the slice (and thus buffer data) getting overriden by
	// future code
	copy(tokenSlice, tokenBuffer.Bytes())
	tokenValue := bytes.NewBuffer(tokenSlice)
	token := Token{tokenType, *tokenValue}
	return &token
}

// bufferStringToNum takes an input buffer and converts it from a string of
// character bytes to a float/int
func bufferStringToNum(tokenType TokenType, inputBuffer bytes.Buffer) *bytes.Buffer {
	bufferString := inputBuffer.String()
	var byteBuffer [binary.MaxVarintLen64]byte
	if tokenType == TokenFloatLiteral {
		num, _ := strconv.ParseFloat(bufferString, 64)
		binary.LittleEndian.PutUint64(byteBuffer[:], math.Float64bits(num))
	} else {
		num, _ := strconv.ParseInt(bufferString, 10, 64)
		binary.PutVarint(byteBuffer[:], num)
	}
	returnBuffer := bytes.NewBuffer(byteBuffer[:])
	return returnBuffer
}

// flushAccumulator empties the contents of the given Buffer into a new Token
// and resets it and the accumulator token type. A convenience function for LexExp.
func flushAccumulator(
	accumulatorType *TokenType,
	accumulatorBuffer *bytes.Buffer,
	tokenBuffer *[]*Token) {
	if *accumulatorType == TokenFloatLiteral || *accumulatorType == TokenIntLiteral {
		convertedBuffer := bufferStringToNum(*accumulatorType, *accumulatorBuffer)
		*tokenBuffer = append(*tokenBuffer, NewTokenRaw(*accumulatorType, *convertedBuffer))
	} else {
		*tokenBuffer = append(*tokenBuffer, NewTokenRaw(*accumulatorType, *accumulatorBuffer))
	}
	accumulatorBuffer.Reset()
	*accumulatorType = TokenNone
}

// peek peeks at the next rune in the given input string.
func peek(input string, currentIndex int) rune {
	// at the end of the string?
	if len(input)-1 == currentIndex {
		return '\000'
	}
	return rune(input[currentIndex+1])
}

// LexExp lexes an input string into Token objects. There are no possible user-facing
// errors from this process.
func LexExp(input string) []*Token {
	var tokens []*Token
	// accumulation variables for multi-character tokens such as idents and literals
	accumulating := false
	var accumulatingType TokenType
	var accumulatorBuffer bytes.Buffer
	// characters that can be used in an ident asides from ., which has meaning outside
	// idents
	specialInitials := "!$%&*/:<=>?^_~"
	// flag as to whether or not the | character has taken effect
	// anything enclosed within | | is a valid ident in R7RS
	identOverride := false
	// operator characters
	operatorChars := "+-/*<=>"
	for index, glyphRune := range input {
		glyph := string(glyphRune)
		if identOverride == true {
			accumulatorBuffer.WriteString(glyph)
			if glyph == "|" {
				flushAccumulator(&accumulatingType, &accumulatorBuffer, &tokens)
				accumulating = false
			}
		} else if unicode.IsSpace(glyphRune) {
			// flush the accumulator if we were trying to accumulate beforehand
			// no multi-char token accepts a space
			if accumulating == true {
				flushAccumulator(&accumulatingType, &accumulatorBuffer, &tokens)
				accumulating = false
			}
			// flush the accumulator for newlines, as well
		} else if glyph == "\n" {
			flushAccumulator(&accumulatingType, &accumulatorBuffer, &tokens)
			accumulating = false
			// lparen
		} else if glyph == "(" {
			if accumulating == true {
				flushAccumulator(&accumulatingType, &accumulatorBuffer, &tokens)
				accumulating = false
			}
			tokens = append(tokens, NewTokenString(TokenLParen, glyph))
			// rparen
		} else if glyph == ")" {
			if accumulating == true {
				flushAccumulator(&accumulatingType, &accumulatorBuffer, &tokens)
				accumulating = false
			}
			tokens = append(tokens, NewTokenString(TokenRParen, glyph))
			// identify any operators
			// normally they'll be a single character, but >= and <= aren't
		} else if strings.ContainsAny(glyph, operatorChars) && (accumulatingType == TokenOp || accumulatingType == TokenNone) {
			// handle >= and <= correctly
			if (glyph == ">" || glyph == "<") && (peek(input, index) == '=') {
				accumulating = true
				accumulatingType = TokenOp
				accumulatorBuffer.WriteString(glyph)
			} else {
				// did we already accumulate > or < and are now on =?
				if accumulating == true && glyph == "=" {
					accumulatorBuffer.WriteString(glyph)
					flushAccumulator(&accumulatingType, &accumulatorBuffer, &tokens)
					accumulating = false
				} else {
					// simplest case if we found a single-character op, just inject it directly
					tokens = append(tokens, NewTokenString(TokenOp, glyph))
				}
			}
			// idents delimited with | can contain pretty much any character
		} else if glyph == "|" {
			if accumulating == true && accumulatingType != TokenIdent {
				flushAccumulator(&accumulatingType, &accumulatorBuffer, &tokens)
			} else if accumulating == false {
				identOverride = true
			}
			accumulating = true
			accumulatorBuffer.WriteString(glyph)
			accumulatingType = TokenIdent
		} else if glyph == "." {
			// . is a valid character in an ident - add it to the accumulator
			// if we were building an ident
			if accumulating == true && accumulatingType == TokenIdent {
				accumulatorBuffer.WriteString(glyph)
				// we can't start an ident with . - are we building a floating point literal?
			} else if chr := peek(input, index); !unicode.IsSpace(chr) && unicode.IsNumber(chr) {
				accumulating = true
				accumulatingType = TokenFloatLiteral
				accumulatorBuffer.WriteString(glyph)
				// there's situations where a standalone . is valid
			} else {
				tokens = append(tokens, NewTokenString(TokenDot, glyph))
			}
			// ident
		} else if unicode.IsLetter(glyphRune) {
			// were we building a number literal beforehand?
			if accumulating == true && accumulatingType != TokenIdent {
				flushAccumulator(&accumulatingType, &accumulatorBuffer, &tokens)
			}
			accumulating = true
			accumulatingType = TokenIdent
			accumulatorBuffer.WriteString(glyph)
			// were we building an ident and are now trying to add a special initial?
		} else if strings.ContainsAny(glyph, specialInitials) {
			if accumulating == true && accumulatingType == TokenIdent {
				accumulatorBuffer.WriteString(glyph)
			} else {
				tokens = append(tokens, NewTokenString(TokenChar, glyph))
			}
			// number literal
		} else if unicode.IsNumber(glyphRune) {
			if accumulating == true && accumulatingType == TokenIdent {
				flushAccumulator(&accumulatingType, &accumulatorBuffer, &tokens)
			}
			accumulating = true
			// only declare that we are accumulating an int if we didn't see a . already
			if accumulatingType != TokenFloatLiteral {
				accumulatingType = TokenIntLiteral
			}
			accumulatorBuffer.WriteString(glyph)
			// we're not sure what this character is, let the parser deal with it
		} else {
			tokens = append(tokens, NewTokenString(TokenChar, glyph))
		}
	}
	// corner case if the input string while we're still accumulating
	// should never happen in proper Scheme, but still...
	if accumulating == true {
		tokens = append(tokens, NewTokenRaw(accumulatingType, accumulatorBuffer))
		accumulatorBuffer.Reset()
		accumulating = false
	}
	return tokens
}
