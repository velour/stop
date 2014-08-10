package ast

import (
	"math"
	"math/big"
	"strconv"
	"strings"

	"github.com/velour/stop/token"
)

func (e Untyped) Source() string { panic("Untyped.Source called") }

func (e *StructType) Source() string { return "struct{…}" }

func (e *InterfaceType) Source() string { return "interface{…}" }

func (e *FunctionType) Source() string {
	return "func" + e.Signature.Source()
}

func (e *Signature) Source() string {
	s := "("
	for i, p := range e.Parameters {
		if i != 0 {
			s += ", "
		}
		if p.Identifier != nil {
			s += p.Identifier.Source() + " "
		}
		if p.DotDotDot {
			s += "..."
		}
		s += p.Type.Source()
	}
	s += ")"

	nr := len(e.Results)
	if nr == 0 {
		return s
	}
	if nr > 1 || e.Results[0].Identifier != nil {
		s += "("
	}
	for i, r := range e.Results {
		if i != 0 {
			s += ", "
		}
		if r.Identifier != nil {
			s += r.Identifier.Source() + " "
		}
		s += r.Type.Source()
	}
	if nr > 1 || e.Results[0].Identifier != nil {
		s += ")"
	}
	return s
}

func (e *ChannelType) Source() string {
	switch {
	case e.Send && e.Receive:
		return "chan " + e.Element.Source()
	case e.Send:
		return "chan<- " + e.Element.Source()
	case e.Receive:
		return "<-chan " + e.Element.Source()
	}
	panic("non-send, non-receive channel")
}

func (e *MapType) Source() string {
	return "map[" + e.Key.Source() + "]" + e.Value.Source()
}

func (e *ArrayType) Source() string {
	if e.Size == nil {
		return "[...]" + e.Element.Source()
	}
	return "[" + e.Size.Source() + "]" + e.Element.Source()
}

func (e *SliceType) Source() string { return "[]" + e.Element.Source() }

func (e *Star) Source() string { return "*" + e.Target.Source() }

func (e *TypeName) Source() string {
	if e.Package != nil {
		return e.Package.Source() + "." + e.Identifier.Source()
	}
	return e.Identifier.Source()
}

func (e *FunctionLiteral) Source() string {
	return e.FunctionType.Source() + "{…}"
}

func (e *CompositeLiteral) Source() string {
	s := e.LiteralType.Source()
	if _, ok := e.LiteralType.(*Star); ok {
		s = "(" + s + ")"
	}
	s += "{"
	for i, e := range e.Elements {
		if i != 0 {
			s += ", "
		}
		if e.Key != nil {
			s += e.Key.Source() + ": "
		}
		s += e.Value.Source()
	}
	return s + "}"
}

func (e *Index) Source() string {
	return e.Expression.Source() + "[" + e.Index.Source() + "]"
}

func (e *Slice) Source() string {
	s := e.Expression.Source() + "["
	if e.Low != nil {
		s += e.Low.Source()
	}
	s += ":"
	if e.High != nil {
		s += e.High.Source()
	}
	if e.Max != nil {
		s += ":" + e.Max.Source()
	}
	return s + "]"
}

func (e *TypeAssertion) Source() string {
	return e.Expression.Source() + ".(" + e.AssertedType.Source() + ")"
}

func (e *Selector) Source() string {
	return e.Parent.Source() + "." + e.Identifier.Source()
}

func (e *Call) Source() string {
	s := e.Function.Source()
	if _, ok := e.Function.(*Star); ok {
		s = "(" + s + ")"
	}
	s += "("
	for i, a := range e.Arguments {
		if i != 0 {
			s += ", "
		}
		s += a.Source()
	}
	if e.DotDotDot {
		s += "..."
	}
	return s + ")"
}

func (e *BinaryOp) Source() string {
	l := e.Left.Source()
	if needParens(e.Op, e.Left) {
		l = "(" + l + ")"
	}
	r := e.Right.Source()
	if needParens(e.Op, e.Right) {
		r = "(" + r + ")"
	}
	return l + " " + e.Op.String() + " " + r
}

func needParens(op token.Token, c Expression) bool {
	b, ok := c.(*BinaryOp)
	return ok && precedence[b.Op] < precedence[op]
}

func (e *UnaryOp) Source() string {
	return e.Op.String() + e.Operand.Source()
}

func (e *Identifier) Source() string { return e.Name }

func (e *IntegerLiteral) Source() string {
	if i := e.Value.Int64(); e.Rune && i <= math.MaxInt32 && i >= math.MinInt32 {
		return strconv.QuoteRune(rune(i))
	}
	return e.Value.String()
}

func (e *FloatLiteral) Source() string { return floatString(e.Value) }

func (e *ComplexLiteral) Source() string {
	var r string
	var zero big.Rat
	if e.Real.Cmp(&zero) != 0 {
		r = floatString(e.Real)
		if e.Imaginary.Cmp(&zero) >= 0 {
			r += "+"
		}
	}
	return r + floatString(e.Imaginary) + "i"
}

func floatString(r *big.Rat) string {
	const prec = 50
	s := r.FloatString(prec)
	if strings.ContainsRune(s, '.') {
		s = strings.TrimRight(s, "0")
		if s[len(s)-1] == '.' {
			s += "0"
		}
	}
	return s
}

func (e *StringLiteral) Source() string { return strconv.Quote(e.Value) }

func (e *BoolLiteral) Source() string {
	if e.Value {
		return "true"
	}
	return "false"
}

func (e *NilLiteral) Source() string { return "nil" }
