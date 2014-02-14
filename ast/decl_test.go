package ast

import (
	"testing"
)

func TestVarDecl(t *testing.T) {
	x, y, z := ident("x"), ident("y"), ident("z")
	tests := parserTests{
		{`var a int`, decls(vars(ms(a), typeName("", "int")))},
		{`var a int = 5`, decls(
			vars(ms(a), typeName("", "int"), intLit("5")),
		)},
		{`var a, b int = 5, 6`, decls(
			vars(ms(a, b), typeName("", "int"), intLit("5"), intLit("6")),
		)},
		{`var (
			a int = 7
			b = 3.14
		)`, decls(
			vars(ms(a), typeName("", "int"), intLit("7")),
			vars(ms(b), nil, floatLit("3.14")),
		)},
		{`var (
			a int = 7
			b, c, d = 12, 13, 14
			x, y, z float64 = 3.0, 4.0, 5.0
		)`, decls(
			vars(ms(a), typeName("", "int"), intLit("7")),
			vars(ms(b, c, d), nil,
				intLit("12"), intLit("13"), intLit("14")),
			vars(ms(x, y, z), typeName("", "float64"),
				floatLit("3.0"), floatLit("4.0"), floatLit("5.0")),
		)},

		// If there is no type then there must be an expr list.
		{`var b`, parseErr("expected")},
	}
	tests.runDeclarations(t)
}

func TestConstDecl(t *testing.T) {
	x, y, z := ident("x"), ident("y"), ident("z")
	tests := parserTests{
		{`const a`, decls(cnst(ms(a), nil))},
		{`const a int`, decls(cnst(ms(a), typeName("", "int")))},
		{`const a int = 5`, decls(
			cnst(ms(a), typeName("", "int"), intLit("5")),
		)},
		{`const a, b int = 5, 6`, decls(
			cnst(ms(a, b), typeName("", "int"), intLit("5"), intLit("6")),
		)},
		{`const (
			a int = 7
			b
		)`, decls(
			cnst(ms(a), typeName("", "int"), intLit("7")),
			cnst(ms(b), nil),
		)},
		{`const (
			a int = 7
			b, c, d
			x, y, z float64 = 3.0, 4.0, 5.0
		)`, decls(
			cnst(ms(a), typeName("", "int"), intLit("7")),
			cnst(ms(b, c, d), nil),
			cnst(ms(x, y, z), typeName("", "float64"),
				floatLit("3.0"), floatLit("4.0"), floatLit("5.0")),
		)},
	}
	tests.runDeclarations(t)
}

func TestTypeDecl(t *testing.T) {
	tests := parserTests{
		{`type a int`, decls(typ(a, typeName("", "int")))},
		{`type (
			a int
			b float64
		)`, decls(
			typ(a, typeName("", "int")),
			typ(b, typeName("", "float64")),
		)},
	}
	tests.runDeclarations(t)
}
