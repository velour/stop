package ast

import (
	"fmt"
	"io"
	"strconv"
)

// Dot writes an abstract syntax tree to an io.Writer using the dot
// language of graphviz.
func Dot(out io.Writer, n Node) (err error) {
	defer recoverError(&err)
	if _, err = io.WriteString(out, "digraph {\n"); err != nil {
		return
	}
	n.dot(0, out)
	_, err = io.WriteString(out, "}")
	return
}

func (n *Call) dot(cur int, out io.Writer) int {
	fun := cur + 1
	arg := n.Function.dot(fun, out)
	arc(out, cur, fun)
	for _, e := range n.Arguments {
		arc(out, cur, arg)
		arg = e.dot(arg, out)
	}
	if n.DotDotDot {
		node(out, cur, "Call...")
	} else {
		node(out, cur, "Call")
	}
	return arg
}

func (n *TypeAssertion) dot(cur int, out io.Writer) int {
	expr := cur + 1
	sel := n.Expression.dot(expr, out)
	node(out, cur, "type assertion")
	arc(out, cur, expr)
	arc(out, cur, sel)
	return n.Type.dot(sel, out)
}

func (n *Selector) dot(cur int, out io.Writer) int {
	expr := cur + 1
	sel := n.Expression.dot(expr, out)
	node(out, cur, "selector")
	arc(out, cur, expr)
	arc(out, cur, sel)
	return n.Selection.dot(sel, out)
}

func (n *BinaryOp) dot(cur int, out io.Writer) int {
	left := cur + 1
	right := n.Left.dot(left, out)
	node(out, cur, n.Op.String())
	arc(out, cur, left)
	arc(out, cur, right)
	return n.Right.dot(right, out)
}

func (n *UnaryOp) dot(cur int, out io.Writer) int {
	kid := cur + 1
	node(out, cur, n.Op.String())
	arc(out, cur, kid)
	return n.Operand.dot(kid, out)
}

func (n *IntegerLiteral) dot(cur int, out io.Writer) int {
	node(out, cur, n.Value.String())
	return cur + 1
}

func (n *FloatLiteral) dot(cur int, out io.Writer) int {
	node(out, cur, n.printString())
	return cur + 1
}

func (n *ImaginaryLiteral) dot(cur int, out io.Writer) int {
	node(out, cur, n.printString())
	return cur + 1
}

func (n *RuneLiteral) dot(cur int, out io.Writer) int {
	node(out, cur, strconv.QuoteRune(n.Value))
	return cur + 1
}

func (n *StringLiteral) dot(cur int, out io.Writer) int {
	node(out, cur, n.Value)
	return cur + 1
}

func (n *Identifier) dot(cur int, out io.Writer) int {
	node(out, cur, n.Name)
	return cur + 1
}

// Node writes a node named cur with the given label to out.
func node(out io.Writer, cur int, label string) {
	_, err := fmt.Fprintf(out, "\tn%d [label=%s]\n", cur, strconv.QuoteToASCII(label))
	if err != nil {
		panic(err)
	}
}

// Arc writes an arc between src and dst to out.
func arc(out io.Writer, src, dst int) {
	_, err := fmt.Fprintf(out, "\tn%d -> n%d\n", src, dst)
	if err != nil {
		panic(err)
	}
}
