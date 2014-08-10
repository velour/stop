package ast

import (
	"fmt"
	"math/big"

	"github.com/velour/stop/token"
)

// Check performs type checking and semantic analysis on the AST,
// returning any errors that are encountered.
func Check(files []*File) error {
	var errs errors

	if _, err := pkgDecls(files); err != nil {
		errs = append(errs, err)
	}

	// First check TypeSpecs and ConstSpecs. These must be checked before
	// MethodDecl, FunctionDecl, and VarSpecs, because the aformentioned
	// declarations need constants within types to be folded.
	for _, f := range files {
		for _, d := range f.Declarations {
			switch d := d.(type) {
			case *TypeSpec:
				if err := d.Check(); err != nil {
					errs = append(errs, err)
				}
			case *ConstSpec:
				if err := d.Check(); err != nil {
					errs = append(errs, err)
				}
			}
		}
	}

	// BUG(eaburns): Check MethodDecl, FunctionDecl, and VarSpecs.

	return errs.ErrorOrNil()
}

// Check checks the TypeSpec, returning any errors.
//
// BUG(eaburns): Need to check for incorrect, recursively-defined types.
func (n *TypeSpec) Check() error {
	t, err := n.Type.Check(n.syms, -1)
	if err == nil {
		n.Type = t.(Type)
	}
	return err
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

func (n *TypeName) Check(syms *symtab, _ int) (Expression, error) {
	n.decl = syms.Find(n.Name)
	if n.decl == nil {
		return nil, Undeclared{&n.Identifier}
	}
	if n.Package == nil {
		return n, nil
	}
	n.Package.decl = syms.Find(n.Name)
	if n.Package.decl == nil {
		return nil, Undeclared{n.Package}
	}
	return n, nil
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

func (n *UnaryOp) Check(syms *symtab, iota int) (Expression, error) {
	var err error
	n.Operand, err = n.Operand.Check(syms, iota)
	if err != nil {
		return nil, err
	}
	n.typ = n.Operand.Type()

	switch n.Op {
	case token.Plus:
		if !IsInteger(n.typ) && !IsComplex(n.typ) {
			return nil, InvalidOperation{n, n.Op, n.Operand}
		}
		switch l := n.Operand.(type) {
		case *IntegerLiteral, *FloatLiteral, *ComplexLiteral:
			return l, nil
		}

	case token.Minus:
		if !IsInteger(n.typ) && !IsComplex(n.typ) {
			return nil, InvalidOperation{n, n.Op, n.Operand}
		}
		switch l := n.Operand.(type) {
		case *IntegerLiteral:
			l.Value.Neg(l.Value)
			return valueOKOrError(l)

		case *FloatLiteral:
			l.Value.Neg(l.Value)
			return valueOKOrError(l)

		case *ComplexLiteral:
			l.Real.Neg(l.Real)
			l.Imaginary.Neg(l.Imaginary)
			return valueOKOrError(l)
		}

	case token.Bang:
		if !IsBool(n.typ) {
			return nil, InvalidOperation{n, n.Op, n.Operand}
		}
		if l, ok := n.Operand.(*BoolLiteral); ok {
			l.Value = !l.Value
			return l, nil
		}

	case token.Carrot:
		if !IsInteger(n.typ) {
			return nil, InvalidOperation{n, n.Op, n.Operand}
		}
		if l, ok := n.Operand.(*IntegerLiteral); ok {
			l.Value.Not(l.Value)
			return valueOKOrError(l)
		}

	case token.And:
		n.typ = &Star{Target: n.typ}
		// BUG(eaburns): Verify that n.Operand is addressable.
		panic("unimplemented")

	case token.LessMinus:
		ch, ok := n.typ.(*ChannelType)
		if !ok || !ch.Receive {
			return nil, InvalidOperation{n, n.Op, n.Operand}
		}
		n.typ = ch.Element

	case token.Star:
		// Instead, this would be a Star node.
		panic("UnaryOp node cannot be Star")
	default:
		panic("bad unary op: " + n.Op.String())
	}

	if iota >= 0 {
		return nil, NotConstant{n.Operand}
	}
	return n, nil
}

// ValueOKOrError returns the literal expression if its value is representable
// by its type, otherwise it returns an error.
func valueOKOrError(l Expression) (Expression, error) {
	if !IsRepresentable(l, l.Type()) {
		return nil, BadConstAssign{l, l.Type()}
	}
	return l, nil
}

func (n *ConstSpec) Check() error {
	switch n.state {
	case checking:
		panic("impossible, not recursive")
	case checkedError:
		return errors{}
	case checkedOK:
		return nil
	}

	var errs errors
	if n.Type != nil {
		if t, err := n.Type.Check(n.syms, -1); err != nil {
			errs = append(errs, err)
			n.Type = nil
		} else {
			n.Type = t.(Type)
		}
	}
	if len(n.Identifiers) != len(n.Values) {
		errs = append(errs, AssignCountMismatch{n})
	}
	if len(errs) > 0 {
		n.state = checkedError
	} else {
		n.state = checkedOK
	}
	for _, v := range n.views {
		if _, err := v.Check(); err != nil {
			errs = append(errs, err)
		}
	}
	return errs.ErrorOrNil()
}

func (n *constSpecView) Check() (v Expression, err error) {
	defer func() {
		if err != nil {
			n.state = checkedError
			return
		}
		n.Value = v
		n.state = checkedOK
	}()
	switch n.state {
	case checking:
		return nil, ConstantLoop{n}
	case checkedError:
		return nil, errors{}
	case checkedOK:
		return n.Value, nil
	}

	if n.Index >= len(n.Values) {
		// This constant has no expression, but the error will be
		// reported when checking the ConstSpec instead of here.
		return nil, errors{}
	}
	v = n.Values[n.Index]

	var errs errors
	if err := n.ConstSpec.Check(); err != nil {
		errs = append(errs, err)
	}
	// If the type is specified in the ConstSpec, then all views get that type.
	// Otherwise, this will assign nil, and each view gets the type from its
	// bound expression.
	n.Type = n.ConstSpec.Type

	n.state = checking
	v, err = v.Check(n.syms, n.Iota)
	switch {
	case err != nil:
		return nil, append(errs, err)
	case !constOperand(v):
		return nil, append(errs, NotConstant{v})
	}

	switch _, vIsUntyped := v.Type().(Untyped); {
	case n.Type != nil && vIsUntyped && !IsRepresentable(v, n.Type):
		return nil, append(errs, BadConstAssign{v, n.Type})
	case n.Type != nil && !IsAssignable(v, n.Type):
		return nil, append(errs, BadAssign{v, n.Type})
	case n.Type == nil:
		n.Type = v.Type()
	}
	if n.Type != nil {
		// All Literals, which this must be after folding, have a SetType method.
		v.(interface {
			SetType(Type)
		}).SetType(n.Type)
	}
	return v, nil
}

// ConstOperand returns true if the expression is a constant operand.
// A constant operand is the result of constant folding on a truly
// constant expression.
func constOperand(e Expression) bool {
	switch e.(type) {
	case *IntegerLiteral, *FloatLiteral, *ComplexLiteral, *StringLiteral, *BoolLiteral:
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
			return (&IntegerLiteral{Value: v, span: n.span}).Check(syms, iota)
		case "true":
			return (&BoolLiteral{Value: true, span: n.span}).Check(syms, iota)
		case "false":
			return (&BoolLiteral{Value: false, span: n.span}).Check(syms, iota)
		case "nil":
			return (&NilLiteral{span: n.span}).Check(syms, iota)
		default:
			panic("unknown predeclared constant: " + n.Name)
		}

	case *constSpecView:
		return d.Check()

	case *predeclaredType:
		return (&TypeName{Identifier: *n}).Check(syms, iota)
	case *TypeSpec:
		return (&TypeName{Identifier: *n}).Check(syms, iota)

	default:
		panic(fmt.Sprintf("unimplemented identifier type: %T", d))
	}
}

func (n *IntegerLiteral) Check(*symtab, int) (Expression, error) {
	if n.Rune {
		n.typ = Untyped(RuneConst)
	} else {
		n.typ = Untyped(IntegerConst)
	}
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
