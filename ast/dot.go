// TODO(eaburns): Should print type information for nodes if available.

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

func (n *Identifier) dot(cur int, out io.Writer) int {
	id := n.Name
	if n.Package != "" {
		id = n.Package + "." + n.Name
	}
	node(out, cur, id)
	return cur + 1
}

func (n *IntegerLiteral) dot(cur int, out io.Writer) int {
	node(out, cur, n.StringValue)
	return cur + 1
}

func (n *FloatLiteral) dot(cur int, out io.Writer) int {
	node(out, cur, n.StringValue)
	return cur + 1
}

func (n *ImaginaryLiteral) dot(cur int, out io.Writer) int {
	node(out, cur, n.StringValue)
	return cur + 1
}

func (n *RuneLiteral) dot(cur int, out io.Writer) int {
	node(out, cur, n.StringValue)
	return cur + 1
}

func (n *StringLiteral) dot(cur int, out io.Writer) int {
	node(out, cur, n.Value)
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
