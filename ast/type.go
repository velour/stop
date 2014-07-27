package ast

import (
	"fmt"
	"math"
	"math/big"

	"github.com/velour/stop/token"
)

// A ConstKind is a constant kind.
type ConstKind int

// Constant kinds.
const (
	NilConst ConstKind = iota
	RuneConst
	IntConst
	FloatConst
	ComplexConst
	StringConst
	BoolConst
)

// An Untyped is a Type that representes an untyped constant.
type Untyped ConstKind

func (Untyped) Start() token.Location { panic("unimplemented") }
func (Untyped) End() token.Location   { panic("unimplemented") }
func (Untyped) Loc() token.Location   { panic("unimplemented") }
func (Untyped) Identical(t Type) bool { return false }
func (n Untyped) Underlying() Type    { return n }
func (n Untyped) Type() Type          { return n }

// IsAssignable returns whether an expression is assignable to a variable of a given type.
//	A value x is assignable to a variable of type T ("x is assignable to T") in any of these cases:
//	x's type is identical to T.
//	x's type V and T have identical underlying types and at least one of V or T is not a named type.
//	T is an interface type and x implements T.
//	x is a bidirectional channel value, T is a channel type, x's type V and T have identical element types, and at least one of V or T is not a named type.
//	x is the predeclared identifier nil and T is a pointer, function, slice, map, channel, or interface type.
//	x is an untyped constant representable by a value of type T.
func IsAssignable(x Expression, t Type) bool {
	xt := x.Type()
	_, xtIsNamed := xt.(*TypeName)
	_, tIsNamed := t.(*TypeName)
	_, xIsNil := x.(*NilLiteral)
	_, xIsUntyped := xt.(Untyped)
	xch, xtIsChan := xt.(*ChannelType)
	tch, tIsChan := t.(*ChannelType)

	switch {
	case xt.Identical(t):
		return true

	case xt.Underlying().Identical(t.Underlying()) && (!xtIsNamed || !tIsNamed):
		return true

	// BUG(eaburns): If t is an interface and x implements t: return true

	case xtIsChan && xch.Send && xch.Receive && tIsChan && xch.ElementType.Identical(tch.ElementType) && (!xtIsNamed || !tIsNamed):
		return true

	case xIsNil && Nilable(t):
		return true

	case xIsUntyped && IsRepresentable(x, t):
		return true
	}

	return false
}

// Nilable returns whether the type can be nil.
func Nilable(t Type) bool {
	switch t.(type) {
	case *Star:
		return true
	case *SliceType:
		return true
	case *MapType:
		return true
	case *ChannelType:
		return true
	case *InterfaceType:
		return true
	default:
		return false
	}
}

var bounds = map[predeclaredType]struct{ min, max *big.Int }{
	Int:     {big.NewInt(minInt), big.NewInt(maxInt)},
	Int8:    {big.NewInt(math.MinInt8), big.NewInt(math.MaxInt8)},
	Int16:   {big.NewInt(math.MinInt16), big.NewInt(math.MaxInt16)},
	Int32:   {big.NewInt(math.MinInt32), big.NewInt(math.MaxInt32)},
	Int64:   {big.NewInt(math.MinInt64), big.NewInt(math.MaxInt64)},
	Uint:    {big.NewInt(0), newUint(maxUint)},
	Uint8:   {big.NewInt(0), newUint(math.MaxUint8)},
	Uint16:  {big.NewInt(0), newUint(math.MaxUint16)},
	Uint32:  {big.NewInt(0), newUint(math.MaxUint32)},
	Uint64:  {big.NewInt(0), newUint(math.MaxUint64)},
	Uintptr: {big.NewInt(0), newUint(maxUintptr)},
}

func newUint(x uint64) *big.Int {
	var i big.Int
	i.SetUint64(x)
	return &i
}

// IsRepresentable returns whether a constant expression can be represented by a type.
func IsRepresentable(x Expression, t Type) bool {
	tn, ok := t.Underlying().(*TypeName)
	if !ok {
		return false
	}
	switch tn.Identifier.decl {
	case Bool:
		u, untyped := x.(Untyped)
		_, boolLit := x.(*BoolLiteral)
		return boolLit || (untyped && u == Untyped(BoolConst))

	case Complex64, Complex128:
		_, cmplxLit := x.(*ComplexLiteral)
		_, floatLit := x.(*FloatLiteral)
		_, intLit := x.(*IntegerLiteral)
		return cmplxLit || floatLit || intLit

	case Float32, Float64:
		_, floatLit := x.(*FloatLiteral)
		_, intLit := x.(*IntegerLiteral)
		return floatLit || intLit

	case String:
		_, strLit := x.(*StringLiteral)
		return strLit

	case Int, Int8, Int16, Int32, Int64, Uint, Uint8, Uint16, Uint32, Uint64, Uintptr:
		d := tn.Identifier.decl.(predeclaredType)
		if l, ok := x.(*IntegerLiteral); ok {
			b := bounds[d]
			return b.min.Cmp(l.Value) <= 0 && l.Value.Cmp(b.max) <= 0
		}
	}
	return false
}

// Identical returns whether the two types are identical.
// Two struct types are identical if they have the same sequence of fields, and if corresponding fields have the same names, and identical types, and identical tags. Two anonymous fields are considered to have the same name. Lower-case field names from different packages are always different.
func (t *StructType) Identical(other Type) bool {
	s, ok := other.(*StructType)
	if !ok || len(t.Fields) != len(s.Fields) {
		return false
	}
	for i := range t.Fields {
		switch m, n := &t.Fields[i], &s.Fields[i]; {
		case (m.Identifier == nil) != (n.Identifier == nil) || (m.Identifier != nil && m.Name != n.Name):
			return false
		case m.Identifier != nil && !m.Exported() && m.pkg != n.pkg:
			return false
		case !m.Type.Identical(n.Type):
			return false
		case (m.Tag == nil) != (n.Tag == nil) || (m.Tag != nil && m.Tag.Value != n.Tag.Value):
			return false
		}
	}
	return true
}

// Identical returns whether the two types are identical.
// Two interface types are identical if they have the same set of methods with the same names and identical function types. Lower-case method names from different packages are always different. The order of the methods is irrelevant.
//
// This method expects the methodSet field of InterfaceTypes to be populated by
// having called their Check methods.
func (t *InterfaceType) Identical(other Type) bool {
	s, ok := other.(*InterfaceType)
	if !ok || len(t.methodSet) != len(s.methodSet) {
		return false
	}
	for i, m := range t.methodSet {
		switch n := s.methodSet[i]; {
		case m.Name != n.Name:
			return false
		case !m.Exported() && m.pkg != n.pkg:
			return false
		case !m.identical(&n.Signature):
			return false
		}
	}
	return true
}

// Identical returns whether the two types are identical.
// Two function types are identical if they have the same number of parameters and result values, corresponding parameter and result types are identical, and either both functions are variadic or neither is. Parameter and result names are not required to match.
func (t *FunctionType) Identical(other Type) bool {
	s, ok := other.(*FunctionType)
	return ok && t.Signature.identical(&s.Signature)
}

func (t *Signature) identical(s *Signature) bool {
	n := len(t.Parameters)
	if n != len(s.Parameters) || len(t.Results) != len(s.Results) || (n > 0 && t.Parameters[n-1].DotDotDot != s.Parameters[n-1].DotDotDot) {
		return false
	}
	for i := range t.Parameters {
		p := &t.Parameters[i]
		q := &s.Parameters[i]
		if !p.Type.Identical(q.Type) {
			return false
		}
	}
	for i := range t.Results {
		p := &t.Results[i]
		q := &s.Results[i]
		if !p.Type.Identical(q.Type) {
			return false
		}
	}
	return true
}

// Identical returns whether the two types are identical.
// Two channel types are identical if they have identical value types and the same direction.
func (t *ChannelType) Identical(other Type) bool {
	s, ok := other.(*ChannelType)
	return ok && t.Receive == s.Receive && t.Send == s.Send && t.ElementType.Identical(s.ElementType)
}

// Identical returns whether the two types are identical.
// Two map types are identical if they have identical key and value types.
func (t *MapType) Identical(other Type) bool {
	s, ok := other.(*MapType)
	return ok && t.Key.Identical(s.Key) && t.Value.Identical(s.Value)
}

// Identical returns whether the two types are identical.
// Two array types are identical if they have identical element types and the same array length.
//
// This method requires the sizes on array types to be folded to IntegerLiterals.
func (t *ArrayType) Identical(other Type) bool {
	s, ok := other.(*ArrayType)
	if !ok {
		return false
	}
	tSz := t.Size.(*IntegerLiteral)
	sSz := s.Size.(*IntegerLiteral)
	return tSz.Value.Cmp(sSz.Value) == 0 && t.ElementType.Identical(s.ElementType)
}

// Identical returns whether the two types are identical.
// Two slice types are identical if they have identical element types.
func (t *SliceType) Identical(other Type) bool {
	s, ok := other.(*SliceType)
	return ok && t.ElementType.Identical(s.ElementType)
}

// Identical returns whether the two types are identical.
// Two pointer types are identical if they have identical base types.
//
// This method requires the Target of Star types to be Types. In other
// words, the Star nodes must not be a pointer dereference operator.
func (t *Star) Identical(other Type) bool {
	s, ok := other.(*Star)
	return ok && t.Target.(Type).Identical(s.Target.(Type))
}

// Identical returns whether the two types are identical.
// Two named types are identical if their type names originate in the same TypeSpec.
func (t *TypeName) Identical(other Type) bool {
	s, ok := other.(*TypeName)
	return ok && t.Identifier.decl == s.Identifier.decl
}

func (t *StructType) Underlying() Type    { return t }
func (t *InterfaceType) Underlying() Type { return t }
func (t *FunctionType) Underlying() Type  { return t }
func (t *ChannelType) Underlying() Type   { return t }
func (t *MapType) Underlying() Type       { return t }
func (t *ArrayType) Underlying() Type     { return t }
func (t *SliceType) Underlying() Type     { return t }
func (t *Star) Underlying() Type          { return t }

func (t *TypeName) Underlying() Type {
	switch d := t.Identifier.decl.(type) {
	case predeclaredType:
		return t
	case *TypeSpec:
		return d.Type.Underlying()
	default:
		panic(fmt.Sprintf("bad TypeName decl: %T", d))
	}
}

func (n *StructType) Type() Type    { return n }
func (n *InterfaceType) Type() Type { return n }
func (n *FunctionType) Type() Type  { return n }
func (n *ChannelType) Type() Type   { return n }
func (n *MapType) Type() Type       { return n }
func (n *ArrayType) Type() Type     { return n }
func (n *SliceType) Type() Type     { return n }

func (n *Star) Type() Type {
	panic("unimplemented")
}

func (n *TypeName) Type() Type { return n }

func (n *CompositeLiteral) Type() Type { return n.LiteralType }

func (n *Index) Type() Type {
	panic("unimplemented")
}

func (n *Slice) Type() Type {
	panic("unimplemented")
}

func (n *TypeAssertion) Type() Type {
	panic("unimplemented")
}

func (n *Selector) Type() Type {
	panic("unimplemented")
}

func (n *Call) Type() Type {
	panic("unimplemented")
}

func (n *BinaryOp) Type() Type {
	panic("unimplemented")
}

func (n *UnaryOp) Type() Type {
	panic("unimplemented")
}

func (n *Identifier) Type() Type {
	switch d := n.decl.(type) {
	case *VarSpec:
		return d.Type

	case *MethodDecl:
		return &FunctionType{Signature: d.Signature}

	case *FunctionDecl:
		return &FunctionType{Signature: d.Signature}

	case *predeclaredFunc:
		// BUG(eaburns): Figure out Identifier.Type for predeclared functions.
		panic("unimplemented")

	// predeclaredType and *TypeSpec are changed to TypeNames by Check.
	// predeclaredConst and *ConstSpec are changed to literals by Check.
	// *ImportDecls are simply invalid in all places that x.Type() will be called.
	default:
		panic(fmt.Sprintf("Type called on identifier with bad decl type: %T", d))
	}
}

func (n *IntegerLiteral) Type() Type { return n.typ }
func (n *FloatLiteral) Type() Type   { return n.typ }
func (n *ComplexLiteral) Type() Type { return n.typ }
func (n *RuneLiteral) Type() Type    { return n.typ }
func (n *StringLiteral) Type() Type  { return n.typ }
func (n *BoolLiteral) Type() Type    { return n.typ }
func (n *NilLiteral) Type() Type     { return n.typ }
