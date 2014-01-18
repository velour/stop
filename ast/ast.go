package ast

import (
	"io"
	"math/big"

	"bitbucket.org/eaburns/stop/token"
)

// The Node interface is implemented by all nodes.
type Node interface {
	// Start returns the start location of the first token that induced
	// this node.
	Start() token.Location

	// End returns the end location of the final token that induced
	// this node.
	End() token.Location

	// Print writes a human-readable representation of the node
	// and the subtree beneath it to out.  If an error occurs then
	// it is panicked.
	print(level int, out io.Writer)

	// Dot writes the node and the subtree beneath it to a writer
	// in the dot language of graphviz.  If an error occurs then it
	// is panicked.
	dot(cur int, out io.Writer) int
}

// The Expression interface is implemented by all nodes that are
// also expressions.
type Expression interface {
	Node
	// Loc returns a location that is indicative of this expression.
	// For example, the location of the operator of a binary
	// expression may be used.  Loc is used for error reporting.
	Loc() token.Location
}

// BinaryOp is an expression node representing a binary operator.
type BinaryOp struct {
	Op          token.Token
	opLoc       token.Location
	Left, Right Expression
}

func (b *BinaryOp) Start() token.Location { return b.Left.Start() }
func (b *BinaryOp) Loc() token.Location   { return b.opLoc }
func (b *BinaryOp) End() token.Location   { return b.Right.End() }

// UnaryOp is an expression node representing a unary operator.
type UnaryOp struct {
	Op      token.Token
	opLoc   token.Location
	Operand Expression
}

func (u *UnaryOp) Start() token.Location { return u.opLoc }
func (u *UnaryOp) Loc() token.Location   { return u.opLoc }
func (u *UnaryOp) End() token.Location   { return u.Operand.End() }

// Identifier is an expression node representing an identifier.
type Identifier struct {
	Package string
	Name    string
	span
}

// IntegerLiteral is an expression node representing a decimal,
// octal, or hex
// integer literal.
type IntegerLiteral struct {
	Value *big.Int
	span
}

// FloatLiteral is an expression node representing a floating point literal.
type FloatLiteral struct {
	Value *big.Rat
	span
}

// ImaginaryLiteral is an expression node representing an imaginary
// literal, the imaginary component of a complex number.
type ImaginaryLiteral struct {
	Value *big.Rat
	span
}

// RuneLiteral is an expression node representing a rune literal.
type RuneLiteral struct {
	Value rune
	span
}

// StringLiteral is an expression node representing an interpreted or
// raw string literal.
type StringLiteral struct {
	Value string
	span
}
