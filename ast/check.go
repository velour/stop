package ast

import (
	"fmt"
	"math/big"
)

// Check performs type checking and semantic analysis on the AST,
// returning any errors that are encountered.
func Check(files []*File) error {
	var errs errors

	_, err := pkgDecls(files)
	errs.Add(err)

	// First check TypeSpecs and ConstSpecs. These must be checked before
	// MethodDecl, FunctionDecl, and VarSpecs, because the aformentioned
	// declarations need constants within types to be folded.
	for _, f := range files {
		for _, d := range f.Declarations {
			switch d := d.(type) {
			case *TypeSpec:
				// BUG(eaburns): Check TypeSpecs.
				panic("unimplemented")
			case *ConstSpec:
				err := d.Check(f.syms)
				errs.Add(err)
			}
		}
	}

	// BUG(eaburns): Check MethodDecl, FunctionDecl, and VarSpecs.

	return errs.ErrorOrNil()
}

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

func (n *ConstSpec) Check(syms *symtab) error {
	switch n.state {
	case checking:
		panic("impossible, not recursive")
	case checkedError:
		return errors{}
	case checkedOK:
		return nil
	}
	if n.syms != nil {
		syms = n.syms
	}

	var errs errors
	if n.Type != nil {
		t, err := n.Type.Check(syms, -1)
		n.Type = t.(Type)
		errs.Add(err)
	}
	if len(n.Identifiers) != len(n.Values) {
		errs.Add(&AssignCountMismatch{n})
	}
	for _, v := range n.views {
		_, err := v.Check(syms, n.Iota)
		errs.Add(err)
	}
	if len(errs) > 0 {
		n.state = checkedError
	} else {
		n.state = checkedOK
	}
	return errs.ErrorOrNil()
}

// ConstSpec.Check must have been called on the ConstSpec of this view in
// order to check its Type field before it's used here.
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
	if n.syms != nil {
		syms = n.syms
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
