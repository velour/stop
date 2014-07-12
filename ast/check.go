package ast

import (
	"fmt"
	"math/big"
)

func (n *StructType) Check(*symtab, int) (Expression, error) {
	panic("unimplemented")
}

func (n *InterfaceType) Check(*symtab, int) (Expression, error) {
	panic("unimplemented")
}

func (n *Signature) Check(*symtab, int) (Expression, error) {
	panic("unimplemented")
}

func (n *ChannelType) Check(*symtab, int) (Expression, error) {
	panic("unimplemented")
}

func (n *MapType) Check(*symtab, int) (Expression, error) {
	panic("unimplemented")
}

func (n *ArrayType) Check(*symtab, int) (Expression, error) {
	panic("unimplemented")
}

func (n *SliceType) Check(*symtab, int) (Expression, error) {
	panic("unimplemented")
}

func (n *Star) Check(*symtab, int) (Expression, error) {
	panic("unimplemented")
}

func (n *TypeName) Check(*symtab, int) (Expression, error) {
	panic("unimplemented")
}

func (n *CompositeLiteral) Check(*symtab, int) (Expression, error) {
	panic("unimplemented")
}

func (n *Index) Check(*symtab, int) (Expression, error) {
	panic("unimplemented")
}

func (n *Slice) Check(*symtab, int) (Expression, error) {
	panic("unimplemented")
}

func (n *TypeAssertion) Check(*symtab, int) (Expression, error) {
	panic("unimplemented")
}

func (n *Selector) Check(*symtab, int) (Expression, error) {
	panic("unimplemented")
}

func (n *Call) Check(*symtab, int) (Expression, error) {
	panic("unimplemented")
}

func (n *BinaryOp) Check(*symtab, int) (Expression, error) {
	panic("unimplemented")
}

func (n *UnaryOp) Check(*symtab, int) (Expression, error) {
	panic("unimplemented")
}

func (n *constSpecView) Check(syms *symtab, iota int) (v Expression, err error) {
	defer func() {
		if err != nil {
			n.state = checkedError
			return
		}
		n.Values[n.Index] = v
		n.state = checkedOK
	}()
	switch n.state {
	case checking:
		return nil, &ConstantLoop{n}
	case checkedError:
		return nil, errors{}
	case checkedOK:
		return n.Value(), nil
	}

	v = n.Value()
	if v == nil {
		// This constant has no expression, but the error will be
		// reported when checking the ConstSpec instead of here.
		return nil, errors{}
	}

	// If the type is specified in the ConstSpec, then all views get that type.
	// Otherwise, this will assign nil, and each view gets the type from its
	// bound expression.
	n.Type = n.ConstSpec.Type

	// If we have a file-scoped symbol table, this must be a package-level
	// constant declaration. We need to use the file-scoped symbol table
	// instead of the one given to us, since the given one may come from
	// a declaration in a separate file that this declaration cannot access.
	if n.syms != nil {
		syms = n.syms
	}

	n.state = checking
	v, err = v.Check(syms, iota)
	switch {
	case err != nil:
		return nil, err
	case !constOperand(v):
		return nil, NotConstant{v}
	}

	switch _, vIsUntyped := v.Type().(Untyped); {
	case n.Type != nil && vIsUntyped && !IsRepresentable(v, n.Type):
		return nil, &BadConstAssign{v, n.Type}
	case n.Type != nil && IsAssignable(v, n.Type):
		return nil, &BadAssign{v, n.Type}
	case n.Type == nil:
		n.Type = v.Type()
	}
	return v, nil
}

// ConstOperand returns true if the expression is a constant operand.
// A constant operand is the result of constant folding on a truly
// constant expression.
func constOperand(e Expression) bool {
	switch e.(type) {
	case *IntegerLiteral, *FloatLiteral, *ComplexLiteral, *RuneLiteral,
		*StringLiteral, *BoolLiteral, *NilLiteral:
		return true
	}
	return false
}

func (n *Identifier) Check(syms *symtab, iota int) (Expression, error) {
	n.decl = syms.Find(n.Name)
	if n.decl == nil {
		return nil, Undeclared{n}
	}
	switch d := n.decl.(type) {
	case *predeclaredConst:
		switch n.Name {
		case "iota":
			v := big.NewInt(int64(iota))
			return &IntegerLiteral{Value: v, span: n.span}, nil
		case "true":
			return &BoolLiteral{Value: true, span: n.span}, nil
		case "false":
			return &BoolLiteral{Value: false, span: n.span}, nil
		case "nil":
			return &NilLiteral{span: n.span}, nil
		default:
			panic("unknown predeclared constant: " + n.Name)
		}

	case *constSpecView:
		return d.Check(syms, iota)

	case *predeclaredType:
		return &TypeName{Identifier: *n}, nil
	case *TypeSpec:
		return &TypeName{Identifier: *n}, nil

	default:
		panic(fmt.Sprintf("unimplemented identifier type: %T", d))
	}
}

func (n *IntegerLiteral) Check(*symtab, int) (Expression, error) {
	n.typ = Untyped(IntegerConst)
	return n, nil
}

func (n *FloatLiteral) Check(*symtab, int) (Expression, error) {
	n.typ = Untyped(FloatConst)
	return n, nil
}

func (n *ComplexLiteral) Check(*symtab, int) (Expression, error) {
	n.typ = Untyped(ComplexConst)
	return n, nil
}

func (n *RuneLiteral) Check(*symtab, int) (Expression, error) {
	n.typ = Untyped(RuneConst)
	return n, nil
}

func (n *StringLiteral) Check(*symtab, int) (Expression, error) {
	n.typ = Untyped(StringConst)
	return n, nil
}

func (n *BoolLiteral) Check(*symtab, int) (Expression, error) {
	n.typ = Untyped(BoolConst)
	return n, nil
}

func (n *NilLiteral) Check(*symtab, int) (Expression, error) {
	n.typ = Untyped(NilConst)
	return n, nil
}

func (n Untyped) Check(*symtab, int) (Expression, error) { return n, nil }
