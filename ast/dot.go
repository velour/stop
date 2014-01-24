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

func (n *ChannelType) dot(cur int, out io.Writer) int {
	typ := cur + 1
	name := "chan"
	if !n.Send {
		name = "<-chan"
	} else if !n.Receive {
		name = "chan<-"
	}
	node(out, cur, name)
	arc(out, cur, typ)
	return n.Type.dot(typ, out)
}

func (n *MapType) dot(cur int, out io.Writer) int {
	key := cur + 1
	node(out, cur, "MapType")
	arcl(out, cur, key, "Key")
	typ := n.Key.dot(key, out)
	arcl(out, cur, typ, "Type")
	return n.Type.dot(typ, out)
}

func (n *ArrayType) dot(cur int, out io.Writer) int {
	size := cur + 1
	node(out, cur, "ArrayType")
	arcl(out, cur, size, "Size")
	typ := n.Size.dot(size, out)
	arcl(out, cur, typ, "Type")
	return n.Type.dot(typ, out)
}

func (n *SliceType) dot(cur int, out io.Writer) int {
	next := cur + 1
	node(out, cur, "SliceType")
	arc(out, cur, next)
	return n.Type.dot(next, out)
}

func (n *PointerType) dot(cur int, out io.Writer) int {
	next := cur + 1
	node(out, cur, "PointerType")
	arc(out, cur, next)
	return n.Type.dot(next, out)
}

func (n *TypeName) dot(cur int, out io.Writer) int {
	next := cur + 1
	node(out, cur, "TypeName")
	if n.Package != "" {
		arcl(out, cur, next, "Package")
		node(out, next, n.Package)
		next++
	}
	arcl(out, cur, next, "Name")
	node(out, next, n.Name)
	next++
	return next
}

func (n *Call) dot(cur int, out io.Writer) int {
	fun := cur + 1
	arg := n.Function.dot(fun, out)
	arc(out, cur, fun)
	for i, e := range n.Arguments {
		if n.DotDotDot && i == len(n.Arguments)-1 {
			arcl(out, cur, arg, "...")
		} else {
			arc(out, cur, arg)
		}
		arg = e.dot(arg, out)
	}
	node(out, cur, "Call")
	return arg
}

func (n *Index) dot(cur int, out io.Writer) int {
	expr := cur + 1
	index := n.Expression.dot(expr, out)
	node(out, cur, "Index")
	arc(out, cur, expr)
	arc(out, cur, index)
	return n.Index.dot(index, out)
}

func (n *Slice) dot(cur int, out io.Writer) int {
	expr := cur + 1
	node(out, cur, "Slice")
	next := n.Expression.dot(expr, out)
	arc(out, cur, expr)
	if n.Low != nil {
		arcl(out, cur, next, "Low")
		next = n.Low.dot(next, out)
	}
	if n.High != nil {
		arcl(out, cur, next, "High")
		next = n.High.dot(next, out)
	}
	if n.Max != nil {
		arcl(out, cur, next, "Max")
		next = n.Max.dot(next, out)
	}
	return next
}

func (n *TypeAssertion) dot(cur int, out io.Writer) int {
	expr := cur + 1
	sel := n.Expression.dot(expr, out)
	node(out, cur, "TypeAssertion")
	arc(out, cur, expr)
	arc(out, cur, sel)
	return n.Type.dot(sel, out)
}

func (n *Selector) dot(cur int, out io.Writer) int {
	expr := cur + 1
	sel := n.Expression.dot(expr, out)
	node(out, cur, "Selector")
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

// Arc writes an arc between src and dst to out.
func arcl(out io.Writer, src, dst int, label string) {
	_, err := fmt.Fprintf(out, "\tn%d -> n%d [label=\"%s\"]\n", src, dst, label)
	if err != nil {
		panic(err)
	}
}
