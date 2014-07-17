package ast

import (
	"fmt"
)

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
	return tSz.Eq(sSz) && t.ElementType.Identical(s.ElementType)
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
	case *predeclaredType:
		return t
	case *TypeSpec:
		return d.Type.Underlying()
	default:
		panic(fmt.Sprintf("bad TypeName decl: %T", d))
	}
}
