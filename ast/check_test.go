package ast

import (
	"reflect"
	"regexp"
	"sort"
	"testing"

	"github.com/eaburns/eq"
	"github.com/eaburns/pp"
	"github.com/velour/stop/token"
)

var (
	intType   = typ("int")
	byteType  = typ("byte")
	uint8Type = typ("uint8")
	runeType  = typ("rune")
	int32Type = typ("int32")
	t1Ident   = typ("T1")
	t1Diff    = typ("T1")
)

func init() {
	intType.Identifier.decl = univScope.Decls["int"]
	byteType.Identifier.decl = univScope.Decls["byte"]
	uint8Type.Identifier.decl = univScope.Decls["uint8"]
	runeType.Identifier.decl = univScope.Decls["rune"]
	int32Type.Identifier.decl = univScope.Decls["int32"]
	t0.Identifier.decl = &TypeSpec{Identifier: *id("T0"), Type: intType}
	t1.Identifier.decl = &TypeSpec{Identifier: *id("T1"), Type: t0}
	// t1Ident is just like t1 and from the same declaration as t1.
	t1Ident.Identifier.decl = t1.Identifier.decl
	// t1Diff is just like t1 but from a different declaration.
	t1Diff.Identifier.decl = &TypeSpec{Identifier: *id("T1"), Type: intType}
}

func TestTypeIdentical(t *testing.T) {
	// We don't care about the contents of these, just that &pkgA != &pkgB.
	var pkgA, pkgB packageDecl

	tests := []struct {
		u, v  Type
		ident bool
	}{
		{intType, intType, true},
		{byteType, byteType, true},
		{byteType, uint8Type, true},
		{runeType, int32Type, true},
		{runeType, intType, false},
		{intType, t0, false},
		{t0, t0, true},
		{t0, t1, false},
		{t1, t1Ident, true},
		{t1, t1Diff, false},
		{t0, &Star{Target: t0}, false},

		{&Star{Target: intType}, &Star{Target: intType}, true},
		{&Star{Target: int32Type}, &Star{Target: runeType}, true},
		{&Star{Target: int32Type}, &Star{Target: intType}, false},
		{
			&Star{Target: &Star{Target: intType}},
			&Star{Target: &Star{Target: intType}},
			true,
		},
		{
			&Star{Target: &Star{Target: intType}},
			&Star{Target: intType},
			false,
		},
		{
			&SliceType{ElementType: intType},
			&SliceType{ElementType: intType},
			true,
		},
		{
			&SliceType{ElementType: int32Type},
			&SliceType{ElementType: runeType},
			true,
		},
		{
			&SliceType{ElementType: int32Type},
			&SliceType{ElementType: intType},
			false,
		},
		{
			&SliceType{ElementType: &SliceType{ElementType: intType}},
			&SliceType{ElementType: &SliceType{ElementType: intType}},
			true,
		},
		{
			&SliceType{ElementType: &SliceType{ElementType: intType}},
			&SliceType{ElementType: intType},
			false,
		},
		{&SliceType{ElementType: int32Type}, t0, false},
		{
			&ArrayType{Size: intLit("1"), ElementType: intType},
			&ArrayType{Size: intLit("1"), ElementType: intType},
			true,
		},
		{
			&ArrayType{Size: intLit("1"), ElementType: intType},
			&ArrayType{Size: intLit("2"), ElementType: intType},
			false,
		},
		{
			&ArrayType{Size: intLit("1"), ElementType: intType},
			&ArrayType{Size: intLit("1"), ElementType: t0},
			false,
		},
		{
			&ArrayType{
				Size: intLit("1"),
				ElementType: &ArrayType{
					Size:        intLit("1"),
					ElementType: t0,
				},
			},
			&ArrayType{
				Size: intLit("1"),
				ElementType: &ArrayType{
					Size:        intLit("1"),
					ElementType: t0,
				},
			},
			true,
		},
		{
			&ArrayType{
				Size: intLit("2"),
				ElementType: &ArrayType{
					Size:        intLit("1"),
					ElementType: t0,
				},
			},
			&ArrayType{
				Size: intLit("1"),
				ElementType: &ArrayType{
					Size:        intLit("2"),
					ElementType: t0,
				},
			},
			false,
		},
		{
			&ArrayType{Size: intLit("1"), ElementType: intType},
			&SliceType{ElementType: intType},
			false,
		},
		{
			&MapType{Key: intType, Value: t1},
			&MapType{Key: intType, Value: t1Ident},
			true,
		},
		{
			&MapType{Key: intType, Value: t1},
			&MapType{Key: intType, Value: t1Diff},
			false,
		},
		{
			&MapType{Key: t0, Value: t1},
			&MapType{Key: intType, Value: t1},
			false,
		},
		{
			&ChannelType{ElementType: intType},
			&ChannelType{ElementType: intType},
			true,
		},
		{
			&ChannelType{Receive: true, ElementType: intType},
			&ChannelType{Receive: true, ElementType: intType},
			true,
		},
		{
			&ChannelType{Receive: true, Send: true, ElementType: intType},
			&ChannelType{Receive: true, Send: true, ElementType: intType},
			true,
		},
		{
			&ChannelType{ElementType: intType},
			&ChannelType{Send: true, ElementType: intType},
			false,
		},
		{
			&ChannelType{Receive: true, ElementType: intType},
			&ChannelType{ElementType: intType},
			false,
		},
		{
			&ChannelType{ElementType: t1},
			&ChannelType{ElementType: t1Diff},
			false,
		},
		{
			&ChannelType{ElementType: &ChannelType{ElementType: runeType}},
			&ChannelType{ElementType: &ChannelType{ElementType: int32Type}},
			true,
		},
		{
			&ChannelType{Receive: true, ElementType: &ChannelType{ElementType: t0}},
			&ChannelType{ElementType: &ChannelType{Receive: true, ElementType: t0}},
			false,
		},
		{&ChannelType{ElementType: t1}, t0, false},
		{&FunctionType{}, &FunctionType{}, true},
		{
			&FunctionType{Signature: Signature{
				Parameters: []ParameterDecl{
					{Type: intType, Identifier: id("a")},
				},
			}},
			&FunctionType{Signature: Signature{
				Parameters: []ParameterDecl{
					{Type: intType, Identifier: id("a")},
				},
			}},
			true,
		},
		{
			&FunctionType{Signature: Signature{
				Parameters: []ParameterDecl{
					{Type: intType, Identifier: id("a")},
				},
			}},
			&FunctionType{Signature: Signature{
				Parameters: []ParameterDecl{
					{Type: intType, Identifier: id("notA")},
				},
			}},
			true,
		},
		{
			&FunctionType{Signature: Signature{
				Parameters: []ParameterDecl{{Type: t1}},
			}},
			&FunctionType{Signature: Signature{
				Parameters: []ParameterDecl{{Type: t1Ident}},
			}},
			true,
		},
		{
			&FunctionType{Signature: Signature{
				Parameters: []ParameterDecl{{Type: t1}},
			}},
			&FunctionType{Signature: Signature{
				Parameters: []ParameterDecl{{Type: t1Diff}},
			}},
			false,
		},
		{
			&FunctionType{Signature: Signature{
				Parameters: []ParameterDecl{{Type: t0}},
			}},
			&FunctionType{Signature: Signature{
				Results: []ParameterDecl{{Type: t0}},
			}},
			false,
		},
		{
			&FunctionType{Signature: Signature{
				Results: []ParameterDecl{{Type: t0}},
			}},
			&FunctionType{Signature: Signature{
				Results: []ParameterDecl{{Type: t0}},
			}},
			true,
		},
		{
			&FunctionType{Signature: Signature{
				Parameters: []ParameterDecl{{Type: t0}, {Type: t1}},
			}},
			&FunctionType{Signature: Signature{
				Parameters: []ParameterDecl{{Type: t0}, {Type: t1}},
			}},
			true,
		},
		{
			&FunctionType{Signature: Signature{
				Parameters: []ParameterDecl{{Type: t0}, {Type: t1}},
			}},
			&FunctionType{Signature: Signature{
				Parameters: []ParameterDecl{{Type: t1}, {Type: t0}},
			}},
			false,
		},
		{
			&FunctionType{Signature: Signature{
				Parameters: []ParameterDecl{{Type: t0}, {Type: t1, DotDotDot: true}},
			}},
			&FunctionType{Signature: Signature{
				Parameters: []ParameterDecl{{Type: t0}, {Type: t1}},
			}},
			false,
		},
		{
			&FunctionType{Signature: Signature{
				Parameters: []ParameterDecl{{Type: t0}, {Type: t1, DotDotDot: true}},
			}},
			&FunctionType{Signature: Signature{
				Parameters: []ParameterDecl{{Type: t0}, {Type: t1, DotDotDot: true}},
			}},
			true,
		},
		{
			&FunctionType{Signature: Signature{
				Parameters: []ParameterDecl{{Type: t0}, {Type: t1}},
				Results:    []ParameterDecl{{Type: intType}},
			}},
			&FunctionType{Signature: Signature{
				Parameters: []ParameterDecl{{Type: t0}, {Type: t1}},
				Results:    []ParameterDecl{{Type: intType}},
			}},
			true,
		},
		{
			&FunctionType{Signature: Signature{
				Parameters: []ParameterDecl{{Type: intType}},
				Results:    []ParameterDecl{{Type: t0}, {Type: t1}},
			}},
			&FunctionType{Signature: Signature{
				Parameters: []ParameterDecl{{Type: t0}, {Type: t1}},
				Results:    []ParameterDecl{{Type: intType}},
			}},
			false,
		},
		{
			&FunctionType{Signature: Signature{
				Results: []ParameterDecl{{Type: t0}, {Type: t1}},
			}},
			&FunctionType{Signature: Signature{
				Results: []ParameterDecl{{Type: t1}, {Type: t0}},
			}},
			false,
		},
		{&FunctionType{}, t0, false},

		// The Methods field is not used when checking for interface identity, but
		// the unexported methodSet field won't print on test errors, so we kindly
		// populate both.
		{&InterfaceType{}, &InterfaceType{}, true},
		{
			&InterfaceType{
				Methods: []Node{
					&Method{Identifier: *id("a")},
				},
				methodSet: []*Method{
					{Identifier: *id("a"), pkg: &pkgA},
				},
			},
			&InterfaceType{
				Methods: []Node{
					&Method{Identifier: *id("a")},
				},
				methodSet: []*Method{
					{Identifier: *id("a"), pkg: &pkgA},
				},
			},
			true,
		},
		{
			&InterfaceType{
				Methods: []Node{
					&Method{Identifier: *id("a")},
				},
				methodSet: []*Method{
					{Identifier: *id("a"), pkg: &pkgA},
				},
			},
			&InterfaceType{
				Methods: []Node{
					&Method{Identifier: *id("a")},
				},
				methodSet: []*Method{
					{Identifier: *id("a"), pkg: &pkgB},
				},
			},
			false,
		},
		{
			&InterfaceType{
				Methods: []Node{
					&Method{Identifier: *id("A")},
				},
				methodSet: []*Method{
					{Identifier: *id("A"), pkg: &pkgA},
				},
			},
			&InterfaceType{
				Methods: []Node{
					&Method{Identifier: *id("A")},
				},
				methodSet: []*Method{
					{Identifier: *id("A"), pkg: &pkgB},
				},
			},
			true,
		},
		{
			&InterfaceType{
				Methods: []Node{
					&Method{Identifier: *id("a")},
				},
				methodSet: []*Method{
					{Identifier: *id("a"), pkg: &pkgA},
				},
			},
			&InterfaceType{},
			false,
		},
		{
			&InterfaceType{
				Methods: []Node{
					&Method{Identifier: *id("a")},
				},
				methodSet: []*Method{
					{Identifier: *id("a"), pkg: &pkgA},
				},
			},
			&InterfaceType{
				Methods: []Node{
					&Method{Identifier: *id("b")},
				},
				methodSet: []*Method{
					{Identifier: *id("b"), pkg: &pkgA},
				},
			},
			false,
		},
		{
			&InterfaceType{
				Methods: []Node{
					&Method{
						Identifier: *id("a"),
						Signature: Signature{
							Parameters: []ParameterDecl{{Type: t0}, {Type: t1}},
							Results:    []ParameterDecl{{Type: intType}},
						},
					},
				},
				methodSet: []*Method{
					{
						Identifier: *id("a"),
						Signature: Signature{
							Parameters: []ParameterDecl{{Type: t0}, {Type: t1}},
							Results:    []ParameterDecl{{Type: intType}},
						},
						pkg: &pkgA,
					},
				},
			},
			&InterfaceType{
				Methods: []Node{
					&Method{
						Identifier: *id("a"),
						Signature: Signature{
							Parameters: []ParameterDecl{{Type: t0}, {Type: t1}},
							Results:    []ParameterDecl{{Type: intType}},
						},
					},
				},
				methodSet: []*Method{
					{
						Identifier: *id("a"),
						Signature: Signature{
							Parameters: []ParameterDecl{{Type: t0}, {Type: t1}},
							Results:    []ParameterDecl{{Type: intType}},
						},
						pkg: &pkgA,
					},
				},
			},
			true,
		},
		{
			&InterfaceType{
				Methods: []Node{
					&Method{
						Identifier: *id("a"),
						Signature: Signature{
							Results: []ParameterDecl{{Type: intType}},
						},
					},
				},
				methodSet: []*Method{
					{
						Identifier: *id("a"),
						Signature: Signature{
							Results: []ParameterDecl{{Type: intType}},
						},
						pkg: &pkgA,
					},
				},
			},
			&InterfaceType{
				Methods: []Node{
					&Method{
						Identifier: *id("a"),
						Signature: Signature{
							Results: []ParameterDecl{{Type: t0}},
						},
					},
				},
				methodSet: []*Method{
					{
						Identifier: *id("a"),
						Signature: Signature{
							Results: []ParameterDecl{{Type: t0}},
						},
						pkg: &pkgA,
					},
				},
			},
			false,
		},
		{
			&InterfaceType{
				Methods: []Node{
					&Method{Identifier: *id("a")},
				},
				methodSet: []*Method{
					{Identifier: *id("a"), pkg: &pkgA},
				},
			},
			&InterfaceType{
				Methods: []Node{
					&Method{
						Identifier: *id("a"),
						Signature: Signature{
							Parameters: []ParameterDecl{{Type: t0}, {Type: t1}},
							Results:    []ParameterDecl{{Type: intType}},
						},
					},
				},
				methodSet: []*Method{
					{
						Identifier: *id("a"),
						Signature: Signature{
							Parameters: []ParameterDecl{{Type: t0}, {Type: t1}},
							Results:    []ParameterDecl{{Type: intType}},
						},
						pkg: &pkgA,
					},
				},
			},
			false,
		},
		{
			&InterfaceType{
				Methods: []Node{
					&Method{Identifier: *id("A")},
					&Method{Identifier: *id("B")},
				},
				methodSet: []*Method{
					{Identifier: *id("A")},
					{Identifier: *id("B")},
				},
			},
			&InterfaceType{
				Methods: []Node{
					&Method{Identifier: *id("A")},
					&Method{Identifier: *id("B")},
				},
				methodSet: []*Method{
					{Identifier: *id("A")},
					{Identifier: *id("B")},
				},
			},
			true,
		},
		{
			&InterfaceType{
				Methods: []Node{
					&Method{Identifier: *id("A")},
					&Method{Identifier: *id("B")},
				},
				methodSet: []*Method{
					{Identifier: *id("A")},
					{Identifier: *id("B")},
				},
			},
			&InterfaceType{
				Methods: []Node{
					&Method{Identifier: *id("A")},
					&Method{Identifier: *id("C")},
				},
				methodSet: []*Method{
					{Identifier: *id("A")},
					{Identifier: *id("C")},
				},
			},
			false,
		},
		{&InterfaceType{}, t0, false},
		{&StructType{}, &StructType{}, true},
		{
			&StructType{Fields: []FieldDecl{{Type: intType}}},
			&StructType{Fields: []FieldDecl{{Type: intType}}},
			true,
		},
		{
			&StructType{Fields: []FieldDecl{
				{Identifier: id("a"), Type: intType},
			}},
			&StructType{Fields: []FieldDecl{
				{Identifier: id("a"), Type: intType},
			}},
			true,
		},
		{
			&StructType{Fields: []FieldDecl{
				{Identifier: id("x"), Type: intType},
				{Identifier: id("y"), Type: &SliceType{ElementType: t0}},
				{Identifier: id("z"), Type: t1},
			}},
			&StructType{Fields: []FieldDecl{
				{Identifier: id("x"), Type: intType},
				{Identifier: id("y"), Type: &SliceType{ElementType: t0}},
				{Identifier: id("z"), Type: t1},
			}},
			true,
		},
		{
			&StructType{Fields: []FieldDecl{
				{Identifier: id("x"), Type: intType},
				{Identifier: id("y"), Type: &SliceType{ElementType: t0}},
				{Identifier: id("z"), Type: t1},
			}},
			&StructType{Fields: []FieldDecl{
				{Identifier: id("x"), Type: intType},
				{Identifier: id("z"), Type: t1},
			}},
			false,
		},
		{
			&StructType{Fields: []FieldDecl{
				{Identifier: id("x"), Type: intType},
				{Identifier: id("y"), Type: &SliceType{ElementType: t0}},
				{Identifier: id("z"), Type: t1},
			}},
			&StructType{Fields: []FieldDecl{
				{Identifier: id("x"), Type: intType},
				{Identifier: id("z"), Type: t1},
			}},
			false,
		},
		{
			&StructType{Fields: []FieldDecl{
				{Identifier: id("x"), Type: t1},
			}},
			&StructType{Fields: []FieldDecl{
				{Identifier: id("x"), Type: intType},
			}},
			false,
		},
		{
			&StructType{Fields: []FieldDecl{
				{Identifier: id("x"), Type: t1, Tag: strLit("tag")},
			}},
			&StructType{Fields: []FieldDecl{
				{Identifier: id("x"), Type: t1, Tag: strLit("tag")},
			}},
			true,
		},
		{
			&StructType{Fields: []FieldDecl{
				{Identifier: id("x"), Type: t1, Tag: strLit("tag")},
			}},
			&StructType{Fields: []FieldDecl{
				{Identifier: id("x"), Type: t1, Tag: strLit("different tag")},
			}},
			false,
		},
		{
			&StructType{Fields: []FieldDecl{{Identifier: id("x")}}},
			&StructType{Fields: []FieldDecl{{Identifier: id("y")}}},
			false,
		},
		{
			&StructType{Fields: []FieldDecl{{Identifier: id("x"), Type: t0}}},
			&StructType{Fields: []FieldDecl{{Type: t0}}},
			false,
		},
		{
			&StructType{Fields: []FieldDecl{
				{Identifier: id("x"), Type: t0, pkg: &pkgA},
			}},
			&StructType{Fields: []FieldDecl{
				{Identifier: id("x"), Type: t0, pkg: &pkgA},
			}},
			true,
		},
		{
			&StructType{Fields: []FieldDecl{
				{Identifier: id("x"), pkg: &pkgA},
			}},
			&StructType{Fields: []FieldDecl{
				{Identifier: id("x"), pkg: &pkgB},
			}},
			false,
		},
		{&StructType{}, t0, false},
	}
	for _, test := range tests {
		ident := test.u.Identical(test.v)
		if ident != test.ident {
			t.Errorf("(%s).Identical(%s)=%t, want %t", pp.MustString(test.u), pp.MustString(test.v), ident, test.ident)
		}
		identReflect := test.v.Identical(test.u)
		if ident == identReflect {
			continue
		}
		u := pp.MustString(test.u)
		v := pp.MustString(test.v)
		t.Errorf("(%s).Identical(%s)=%t, but (%s).Identical(%s)=%t", u, v, ident, v, u, identReflect)
	}
}

func TestTypeUnderlying(t *testing.T) {
	tests := []struct {
		t Type
		u Type
	}{
		{t: intType, u: intType},
		{t: byteType, u: byteType},
		{t: t0, u: intType},
		{t: t1, u: intType},
		{t: &Star{Target: t0}, u: &Star{Target: t0}},
		{
			t: &SliceType{ElementType: intType},
			u: &SliceType{ElementType: intType},
		},
		{
			t: &ArrayType{Size: intLit("5"), ElementType: intType},
			u: &ArrayType{Size: intLit("5"), ElementType: intType},
		},
		{
			t: &MapType{Key: t0, Value: intType},
			u: &MapType{Key: t0, Value: intType},
		},
		{
			t: &ChannelType{Send: true, ElementType: intType},
			u: &ChannelType{Send: true, ElementType: intType},
		},
		{t: &FunctionType{}, u: &FunctionType{}},
		{t: &InterfaceType{}, u: &InterfaceType{}},
		{t: &StructType{}, u: &StructType{}},
	}
	for _, test := range tests {
		u := test.t.Underlying()
		if eq.Deep(u, test.u) {
			continue
		}
		t.Errorf("(%s).Underlying()=%s, want %s", pp.MustString(test.t), pp.MustString(u), pp.MustString(test.u))
	}
}

func TestPkgDecls(t *testing.T) {
	tests := []struct {
		src []string
		// The top-level identifiers.
		ids []string
		// An expected error, if non-empty.
		err string
	}{
		{
			src: []string{
				`package a; const a = 1`,
				`package a; const B = 1`,
			},
			ids: []string{"a", "B"},
		},
		{
			src: []string{`package a; const a, B = B, 1`},
			ids: []string{"a", "B"},
		},
		{
			src: []string{
				`package a
			const (
				a, B = B, 1
				c, d
			)`,
			},
			ids: []string{"a", "B", "c", "d"},
		},
		{
			src: []string{
				`package a; var a = 1`,
				`package a; var B = 1`,
			},
			ids: []string{"a", "B"},
		},
		{
			src: []string{`package a; var a, B = 1, 2`},
			ids: []string{"a", "B"},
		},
		{
			src: []string{
				`package a
			var (
				a, B = 0, 1
				c, d = 3.0, 4.0
			)`,
			},
			ids: []string{"a", "B", "c", "d"},
		},
		{
			src: []string{
				`package a
			type (
				a struct{}
				B struct { C }
			)`,
			},
			ids: []string{"a", "B"},
		},
		{
			src: []string{
				`package a
			func a(){}
			func B() int { return 0 }`,
			},
			ids: []string{"a", "B"},
		},
		{
			src: []string{
				`package a
			func a(){}
			func B() int { return 0 }`,
				`package a
			func c(e int){}
			func d() (f int) { return 0 }`,
			},
			ids: []string{"a", "B", "c", "d"},
		},
		{
			src: []string{
				`package a
			func (z int) a(){}
			func (z float64) B() int { return 0 }`,
			},
			ids: []string{"a", "B"},
		},
		{
			src: []string{
				`package a
			func (z T0) a(){}
			func (z T1) B() int { return 0 }`,
				`package a
			func (z T2) c(e int){}
			func (z T3) d() (f int) { return 0 }`,
			},
			ids: []string{"a", "B", "c", "d"},
		},

		// Redeclaration errors.
		{
			src: []string{`package a; const a = 1; const a = 2`},
			err: "a redeclared",
		},
		{
			src: []string{`package a; const a = 1; var a = 2`},
			err: "a redeclared",
		},
		{
			src: []string{`package a; const a = 1; func a(){}`},
			err: "a redeclared",
		},
		{
			src: []string{
				`package a; const a = 1`,
				`package a; const a = 2`,
			},
			err: "a redeclared",
		},
		{
			src: []string{`
				package a
				import (
					"fmt"
					"fmt"
				)
			`},
			err: "fmt redeclared",
		},
		{
			src: []string{`
				package a
				import (
					"fmt"
					"foo/bar/fmt"
				)
			`},
			err: "fmt redeclared",
		},
		{
			src: []string{`
				package a
				import fmt "foo"
				import "fmt"
			`},
			err: "fmt redeclared",
		},
		{
			src: []string{`
				package a
				import (
					"fmt"
					fmt "foo/bar"
				)
			`},
			err: "fmt redeclared",
		},
		{
			// Not an error to redeclare the same import in different files.
			src: []string{
				`package a; import "fmt"`,
				`package a; import "fmt"`,
			},
		},

		// The blank identifier is not redeclared.
		{
			src: []string{`package a; const _ = 1; const _ = 2`},
		},
	}

	for _, test := range tests {
		files := parseSrcFiles(t, test.src)
		if files == nil {
			continue
		}
		var ids []string
		seen := make(map[Declaration]bool)
		s, err := pkgDecls(files)
		if test.err != "" {
			if err == nil {
				t.Errorf("pkgDecls(%v), err=nil, want matching %s", test.src, test.err)
			} else if !regexp.MustCompile(test.err).MatchString(err.Error()) {
				t.Errorf("pkgDecls(%v), err=[%v], want matching [%s]", test.src, err, test.err)
			}
			continue
		}
		if err != nil {
			t.Errorf("pkgDecls(%v), unexpected err=[%v]", test.src, err)
			continue
		}
		for id, d := range s.Decls {
			ids = append(ids, id)
			if seen[d] {
				t.Errorf("pkgDecls(%v), multiple idents map to declaration %s", test.src, pp.MustString(d))
			}
			seen[d] = true
		}
		sort.Strings(ids)
		sort.Strings(test.ids)
		if reflect.DeepEqual(ids, test.ids) {
			continue
		}
		t.Errorf("pkgDecls(%v)=%v, want %v", test.src, ids, test.ids)
	}
}

func parseSrcFiles(t *testing.T, srcFiles []string) []*SourceFile {
	var files []*SourceFile
	for _, src := range srcFiles {
		p := NewParser(token.NewLexer("", src))
		file, err := Parse(p)
		if err == nil {
			files = append(files, file)
			continue
		}
		t.Fatalf("Parse(%s), got unexpected error %v", src, err)
		return nil
	}
	return files
}

func TestScopeFind(t *testing.T) {
	src := `
		package testpkg
		const (
			π = 3.1415926535
			name = "eaburns"
		)
		var mayChange int = f()
		var len = 1	// Shadows the predeclared len function.
		func f() int { return 0 }
	`
	p := NewParser(token.NewLexer("", src))
	s, err := pkgDecls([]*SourceFile{parseSourceFile(p)})
	if err != nil {
		panic(err)
	}

	tests := []struct {
		id       string
		declType reflect.Type
	}{
		{"π", reflect.TypeOf(&constSpecView{})},
		{"name", reflect.TypeOf(&constSpecView{})},
		{"mayChange", reflect.TypeOf(&varSpecView{})},
		{"len", reflect.TypeOf(&varSpecView{})},
		{"f", reflect.TypeOf(&FunctionDecl{})},
		{"notDefined", nil},
	}
	for _, test := range tests {
		d := s.Find(test.id)
		typ := reflect.TypeOf(d)
		if (d == nil && test.declType != nil) || (d != nil && typ != test.declType) {
			t.Errorf("Find(%s)=%s, want %s", test.id, typ, test.declType)
		}
	}

	// Our testing source file shadowed len, but let's check that it's
	// still findable in the universal scope.
	d := univScope.Find("len")
	typ := reflect.TypeOf(d)
	lenType := reflect.TypeOf(&predeclaredFunc{})
	if d == nil || typ != lenType {
		t.Errorf("Find(len)=%s, want %s", typ, lenType)
	}

	// Rune and byte are aliases for int32 and uint8 respectively;
	// the have the same declaration.
	if univScope.Find("rune") != univScope.Find("int32") {
		t.Errorf("rune is not an alias for int32")
	}
	if univScope.Find("byte") != univScope.Find("uint8") {
		t.Errorf("byte is not an alias for uint8")
	}
}
