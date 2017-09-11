package schego

import (
	"bytes"
	"encoding/binary"
	"errors"
	"math"
	"strconv"
)

type AstNodeType int

const (
	ProgramNode AstNodeType = iota
	AddNode
	SubNode
	MulNode
	DivNode
	GtNode
	LtNode
	GteNode
	LteNode
	EqNode
	IntNode
	FloatNode
	StringNode
)

// base interface for functions needing to accept any kind of AST node
type AstNode interface {
	GetSubNodes() []AstNode
	AddSubNode(AstNode)
	GetType() AstNodeType
	DebugString() string
}

// base struct that all AST node implementations build off of
type SExp struct {
	subNodes []AstNode
}

func (s *SExp) GetSubNodes() []AstNode {
	return s.subNodes
}
func (s *SExp) AddSubNode(node AstNode) {
	s.subNodes = append(s.subNodes, node)
}

type Program struct {
	SExp
}

func NewProgram(nodes ...AstNode) *Program {
	program := new(Program)
	for _, node := range nodes {
		program.AddSubNode(node)
	}
	return program
}
func (p Program) GetType() AstNodeType {
	return ProgramNode
}

type AddExp struct {
	SExp
}

func NewAddExp(lhs AstNode, rhs AstNode) *AddExp {
	node := new(AddExp)
	node.AddSubNode(lhs)
	node.AddSubNode(rhs)
	return node
}
func (a AddExp) GetType() AstNodeType {
	return AddNode
}
func (a AddExp) DebugString() string {
	return "AddExp(" + a.subNodes[0].DebugString() + ", " + a.subNodes[1].DebugString() + ")"
}

type SubExp struct {
	SExp
}

func NewSubExp(lhs AstNode, rhs AstNode) *SubExp {
	node := new(SubExp)
	node.AddSubNode(lhs)
	node.AddSubNode(rhs)
	return node
}
func (s SubExp) GetType() AstNodeType {
	return SubNode
}
func (s SubExp) DebugString() string {
	return "SubExp(" + s.subNodes[0].DebugString() + ", " + s.subNodes[1].DebugString() + ")"
}

type MulExp struct {
	SExp
}

func NewMulExp(lhs AstNode, rhs AstNode) *MulExp {
	node := new(MulExp)
	node.AddSubNode(lhs)
	node.AddSubNode(rhs)
	return node
}
func (m MulExp) GetType() AstNodeType {
	return MulNode
}
func (m MulExp) DebugString() string {
	return "MulExp(" + m.subNodes[0].DebugString() + ", " + m.subNodes[1].DebugString() + ")"
}

type DivExp struct {
	SExp
}

func NewDivExp(lhs AstNode, rhs AstNode) *DivExp {
	node := new(DivExp)
	node.AddSubNode(lhs)
	node.AddSubNode(rhs)
	return node
}
func (d DivExp) GetType() AstNodeType {
	return DivNode
}
func (d DivExp) DebugString() string {
	return "DivExp(" + d.subNodes[0].DebugString() + ", " + d.subNodes[1].DebugString() + ")"
}

type LtExp struct {
	SExp
}

func NewLtExp(lhs AstNode, rhs AstNode) *LtExp {
	node := new(LtExp)
	node.AddSubNode(lhs)
	node.AddSubNode(rhs)
	return node
}
func (l LtExp) GetType() AstNodeType {
	return LtNode
}
func (l LtExp) DebugString() string {
	return "LtExp(" + l.subNodes[0].DebugString() + ", " + l.subNodes[1].DebugString() + ")"
}

type LteExp struct {
	SExp
}

func NewLteExp(lhs AstNode, rhs AstNode) *LteExp {
	node := new(LteExp)
	node.AddSubNode(lhs)
	node.AddSubNode(rhs)
	return node
}
func (l LteExp) GetType() AstNodeType {
	return LteNode
}
func (l LteExp) DebugString() string {
	return "LteExp(" + l.subNodes[0].DebugString() + ", " + l.subNodes[1].DebugString() + ")"
}

type GtExp struct {
	SExp
}

func NewGtExp(lhs AstNode, rhs AstNode) *GtExp {
	node := new(GtExp)
	node.AddSubNode(lhs)
	node.AddSubNode(rhs)
	return node
}
func (g GtExp) GetType() AstNodeType {
	return GtNode
}
func (g GtExp) DebugString() string {
	return "GtExp(" + g.subNodes[0].DebugString() + ", " + g.subNodes[1].DebugString() + ")"
}

type GteExp struct {
	SExp
}

func NewGteExp(lhs AstNode, rhs AstNode) *GteExp {
	node := new(GteExp)
	node.AddSubNode(lhs)
	node.AddSubNode(rhs)
	return node
}
func (g GteExp) GetType() AstNodeType {
	return LteNode
}
func (g GteExp) DebugString() string {
	return "GteExp(" + g.subNodes[0].DebugString() + ", " + g.subNodes[1].DebugString() + ")"
}

type EqExp struct {
	SExp
}

func NewEqExp(lhs AstNode, rhs AstNode) *EqExp {
	node := new(EqExp)
	node.AddSubNode(lhs)
	node.AddSubNode(rhs)
	return node
}
func (e EqExp) GetType() AstNodeType {
	return EqNode
}
func (e EqExp) DebugString() string {
	return "EqExp(" + e.subNodes[0].DebugString() + ", " + e.subNodes[1].DebugString() + ")"
}

type IntLiteral struct {
	SExp
	Value int64
}

func NewIntLiteral(value int64) *IntLiteral {
	node := new(IntLiteral)
	node.Value = value
	return node
}
func (i IntLiteral) GetType() AstNodeType {
	return IntNode
}
func (i IntLiteral) DebugString() string {
	return strconv.FormatInt(i.Value, 10)
}

type FloatLiteral struct {
	SExp
	Value float64
}

func NewFloatLiteral(value float64) *FloatLiteral {
	node := new(FloatLiteral)
	node.Value = value
	return node
}
func (f FloatLiteral) GetType() AstNodeType {
	return FloatNode
}
func (f FloatLiteral) DebugString() string {
	return strconv.FormatFloat(f.Value, 'g', -1, 64)
}

type StringLiteral struct {
	SExp
	Value string
}

func NewStringLiteral(value string) *StringLiteral {
	node := new(StringLiteral)
	node.Value = value
	return node
}
func (s StringLiteral) GetType() AstNodeType {
	return StringNode
}
func (s StringLiteral) DebugString() string {
	return "\"" + s.Value + "\""
}

// ParseTokens takes tokens and returns an AST (Abstract Syntax Tree) representation
func ParseTokens(tokens []*Token) *Program {
	program := NewProgram()
	currentIndex := 0
	for len(tokens)-1 >= currentIndex {
		node, _ := parseExpression(tokens, &currentIndex)
		program.AddSubNode(node)
	}
	return program
}

// accept checks to see if the current token matches a given token type, and advances if so
func accept(tokens []*Token, expectedType TokenType, currentIndex *int) bool {
	if tokens[*currentIndex].Type == expectedType {
		*currentIndex++
		return true
	}
	return false
}

// grabAccepted returns the token just before current, useful for grabbing the value of an accepted token
func grabAccepted(tokens []*Token, currentIndex *int) *Token {
	return tokens[*currentIndex-1]
}

// expect returns an error if the current token doesn't match the given type
func expect(tokens []*Token, expectedType TokenType, currentIndex *int) error {
	if len(tokens)-1 < *currentIndex {
		return errors.New("Unexpected EOF")
	} else if tokens[*currentIndex].Type != expectedType {
		return errors.New("Unexpected token")
	}
	return nil
}

func parseExpression(tokens []*Token, currentIndex *int) (AstNode, error) {
	// try literals first
	if accept(tokens, TokenIntLiteral, currentIndex) {
		literal := grabAccepted(tokens, currentIndex)
		return NewIntLiteral(bufferToInt(literal.Value)), nil
	} else if accept(tokens, TokenFloatLiteral, currentIndex) {
		literal := grabAccepted(tokens, currentIndex)
		return NewFloatLiteral(bufferToFloat(literal.Value)), nil
	} else if accept(tokens, TokenStringLiteral, currentIndex) {
		literal := grabAccepted(tokens, currentIndex)
		return NewStringLiteral(literal.Value.String()), nil
	}
	// not a literal, attempt to parse an expression
	lparenError := expect(tokens, TokenLParen, currentIndex)
	if lparenError != nil {
		return nil, lparenError
	}
	// jump past the lparen
	*currentIndex++
	if accept(tokens, TokenOp, currentIndex) {
		// grab the operator token so we can find out which one it is
		opToken := grabAccepted(tokens, currentIndex)
		// parse the left-hand and right hand sides recursively
		// this also takes care of handling nested expressions
		lhs, lhsError := parseExpression(tokens, currentIndex)
		if lhsError != nil {
			return nil, lhsError
		}
		rhs, rhsError := parseExpression(tokens, currentIndex)
		if rhsError != nil {
			return nil, rhsError
		}

		var expNode AstNode
		switch opToken.Value.String() {
		case "+":
			expNode = NewAddExp(lhs, rhs)
		case "-":
			expNode = NewSubExp(lhs, rhs)
		case "*":
			expNode = NewMulExp(lhs, rhs)
		case "/":
			expNode = NewDivExp(lhs, rhs)
		case "<":
			expNode = NewLtExp(lhs, rhs)
		case "<=":
			expNode = NewLteExp(lhs, rhs)
		case ">":
			expNode = NewGtExp(lhs, rhs)
		case ">=":
			expNode = NewGteExp(lhs, rhs)
		case "=":
			expNode = NewEqExp(lhs, rhs)
		}

		// make sure the expression has a closing rparen
		expError := closeExp(tokens, currentIndex)
		if expError != nil {
			return nil, expError
		}
		return expNode, nil
	}
	// no matches?
	return nil, errors.New("Unexpected token")
}

// convenience function to ensure an expression is properly closed
func closeExp(tokens []*Token, currentIndex *int) error {
	rparenError := expect(tokens, TokenRParen, currentIndex)
	if rparenError != nil {
		return rparenError
	}
	*currentIndex += 1
	return nil
}

func bufferToInt(buffer bytes.Buffer) int64 {
	num, _ := binary.Varint(buffer.Bytes())
	return num
}
func bufferToFloat(buffer bytes.Buffer) float64 {
	bits := binary.LittleEndian.Uint64(buffer.Bytes())
	return math.Float64frombits(bits)
}
