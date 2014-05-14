package ast

import (
	"bytes"
	"math/big"
	"regexp"
	"testing"

	"github.com/eaburns/eq"
	"github.com/eaburns/pp"
	"github.com/velour/stop/token"
)

var (
	a, b, c, d = id("a"), id("b"), id("c"), id("d")
	x, y, z    = id("x"), id("y"), id("z")
	bigInt     = sel(id("big"), id("Int")).(Type)
)

func runeLit(r rune) *RuneLiteral { return &RuneLiteral{Value: r} }

func strLit(s string) *StringLiteral { return &StringLiteral{Value: s} }

func intLit(s string) *IntegerLiteral {
	var i big.Int
	i.SetString(s, 0)
	return &IntegerLiteral{Value: &i}
}

func floatLit(s string) *FloatLiteral {
	var r big.Rat
	r.SetString(s)
	return &FloatLiteral{Value: &r}
}

func id(s string) *Identifier                         { return &Identifier{Name: s} }
func unOp(op token.Token, l Expression) *UnaryOp      { return &UnaryOp{Op: op, Operand: l} }
func binOp(op token.Token, l, r Expression) *BinaryOp { return &BinaryOp{Op: op, Left: l, Right: r} }
func call(f Expression, ddd bool, args ...Expression) *Call {
	return &Call{Function: f, Arguments: args, DotDotDot: ddd}
}
func assert(e Expression, t Type) *TypeAssertion { return &TypeAssertion{Expression: e, Type: t} }
func index(e, i Expression) *Index               { return &Index{Expression: e, Index: i} }
func slice(e, l, h, m Expression) *Slice         { return &Slice{Expression: e, Low: l, High: h, Max: m} }

func sel(e Expression, ids ...*Identifier) Expression {
	p := e
	for _, i := range ids {
		p = &Selector{Parent: p, Name: i}
	}
	return p
}

func TestParseVarDecl(t *testing.T) {
	parserTests{
		{
			`var a big.Int`,
			Declarations{
				&VarSpec{Names: []Identifier{*a}, Type: bigInt},
			},
		},
		{
			`var a int = 5`,
			Declarations{
				&VarSpec{
					Names:  []Identifier{*a},
					Type:   id("int"),
					Values: []Expression{intLit("5")},
				},
			},
		},
		{
			`var a, b int = 5, 6`,
			Declarations{
				&VarSpec{
					Names:  []Identifier{*a, *b},
					Type:   id("int"),
					Values: []Expression{intLit("5"), intLit("6")},
				},
			},
		},
		{
			`var (
				a int = 7
				b = 3.14
			)`,
			Declarations{
				&VarSpec{
					Names:  []Identifier{*a},
					Type:   id("int"),
					Values: []Expression{intLit("7")},
				},
				&VarSpec{
					Names:  []Identifier{*b},
					Values: []Expression{floatLit("3.14")},
				},
			},
		},
		{
			`var (
				a int = 7
				b, c, d = 12, 13, 14
				x, y, z float64 = 3.0, 4.0, 5.0
			)`,
			Declarations{
				&VarSpec{
					Names:  []Identifier{*a},
					Type:   id("int"),
					Values: []Expression{intLit("7")},
				},
				&VarSpec{
					Names:  []Identifier{*b, *c, *d},
					Values: []Expression{intLit("12"), intLit("13"), intLit("14")},
				},
				&VarSpec{
					Names:  []Identifier{*x, *y, *z},
					Type:   id("float64"),
					Values: []Expression{floatLit("3.0"), floatLit("4.0"), floatLit("5.0")},
				},
			},
		},

		// If there is no type then there must be an expr list.
		{`var b`, parseError{"expected"}},
	}.run(t, func(p *Parser) Node { return parseDeclarations(p) })
}

func TestParseConstDecl(t *testing.T) {
	parserTests{
		{
			`const a`,
			Declarations{
				&ConstSpec{Names: []Identifier{*a}},
			},
		},
		{
			`const a big.Int`,
			Declarations{
				&ConstSpec{Names: []Identifier{*a}, Type: bigInt},
			},
		},
		{
			`const a int = 5`,
			Declarations{
				&ConstSpec{
					Names:  []Identifier{*a},
					Type:   id("int"),
					Values: []Expression{intLit("5")},
				},
			},
		},
		{
			`const a, b int = 5, 6`,
			Declarations{
				&ConstSpec{
					Names:  []Identifier{*a, *b},
					Type:   id("int"),
					Values: []Expression{intLit("5"), intLit("6")},
				},
			},
		},
		{
			`const (
				a int = 7
				b
			)`,
			Declarations{
				&ConstSpec{
					Names:  []Identifier{*a},
					Type:   id("int"),
					Values: []Expression{intLit("7")},
				},
				&ConstSpec{Names: []Identifier{*b}},
			},
		},
		{
			`const (
				a int = 7
				b, c, d
				x, y, z float64 = 3.0, 4.0, 5.0
			)`,
			Declarations{
				&ConstSpec{
					Names:  []Identifier{*a},
					Type:   id("int"),
					Values: []Expression{intLit("7")},
				},
				&ConstSpec{Names: []Identifier{*b, *c, *d}},
				&ConstSpec{
					Names:  []Identifier{*x, *y, *z},
					Type:   id("float64"),
					Values: []Expression{floatLit("3.0"), floatLit("4.0"), floatLit("5.0")},
				},
			},
		},
	}.run(t, func(p *Parser) Node { return parseDeclarations(p) })
}

func TestParseTypeDecl(t *testing.T) {
	parserTests{
		{
			`type a big.Int`,
			Declarations{
				&TypeSpec{Name: *a, Type: bigInt},
			},
		},
		{
			`type (
				a big.Int
				b struct{}
			)`,
			Declarations{
				&TypeSpec{Name: *a, Type: bigInt},
				&TypeSpec{Name: *b, Type: &StructType{}},
			},
		},
	}.run(t, func(p *Parser) Node { return parseDeclarations(p) })
}

func TestParseType(t *testing.T) {
	parserTests{
		{`a`, a},
		{`(a)`, a},
		{`((a))`, a},
		{`*(a)`, &Star{Target: a}},
		{`[](a)`, &SliceType{Type: a}},
		{`[]*(a)`, &SliceType{Type: &Star{Target: a}}},
		{`*[](a)`, &Star{Target: &SliceType{Type: a}}},
		{`map[a]b`, &MapType{Key: a, Type: b}},

		{`[]func()`, &SliceType{Type: &FunctionType{}}},
		{
			`[]func()<-chan int`,
			&SliceType{
				Type: &FunctionType{Signature: Signature{
					ParameterList{},
					ParameterList{Parameters: []ParameterDecl{
						{Type: &ChannelType{Receive: true, Type: id("int")}},
					}},
				}},
			},
		},
		{
			`[]interface{ c()<-chan[5]big.Int }`,
			&SliceType{
				Type: &InterfaceType{Methods: []Node{
					&Method{
						Name: *c,
						Signature: Signature{
							ParameterList{},
							ParameterList{Parameters: []ParameterDecl{
								{Type: &ChannelType{
									Receive: true,
									Type: &ArrayType{
										Size: intLit("5"),
										Type: bigInt,
									},
								}},
							}},
						},
					},
				}},
			},
		},
		{`1`, parseError{"expected"}},
	}.run(t, func(p *Parser) Node { return parseType(p) })
}

func TestParseStructType(t *testing.T) {
	parserTests{
		{`struct{}`, &StructType{Fields: []FieldDecl{}}},
		{`struct{a}`, &StructType{Fields: []FieldDecl{
			{Type: a},
		}}},
		{`struct{a; b; c}`, &StructType{Fields: []FieldDecl{
			{Type: a},
			{Type: b},
			{Type: c},
		}}},
		{`struct{a, b c}`, &StructType{Fields: []FieldDecl{
			{Identifiers: []Identifier{*a, *b}, Type: c},
		}}},
		{`struct{a b; c}`, &StructType{Fields: []FieldDecl{
			{Identifiers: []Identifier{*a}, Type: b},
			{Type: c},
		}}},
		{`struct{a; b c}`, &StructType{Fields: []FieldDecl{
			{Type: a},
			{Identifiers: []Identifier{*b}, Type: c},
		}}},
		{`struct{big.Int}`, &StructType{Fields: []FieldDecl{
			{Type: bigInt},
		}}},
		{`struct{big.Int; a}`, &StructType{Fields: []FieldDecl{
			{Type: bigInt},
			{Type: a},
		}}},
		{`struct{big.Int; a b}`, &StructType{Fields: []FieldDecl{
			{Type: bigInt},
			{Identifiers: []Identifier{*a}, Type: b},
		}}},
		{`struct{big.Int; a big.Int}`, &StructType{Fields: []FieldDecl{
			{Type: bigInt},
			{Identifiers: []Identifier{*a}, Type: bigInt},
		}}},
		{`struct{*big.Int}`, &StructType{Fields: []FieldDecl{
			{Type: &Star{Target: bigInt}},
		}}},
		{`struct{*big.Int; a}`, &StructType{Fields: []FieldDecl{
			{Type: &Star{Target: bigInt}},
			{Type: a},
		}}},
		{`struct{a; *big.Int}`, &StructType{Fields: []FieldDecl{
			{Type: a},
			{Type: &Star{Target: bigInt}},
		}}},

		// Tagged.
		{`struct{a "your it"}`, &StructType{Fields: []FieldDecl{
			{Type: a, Tag: strLit("your it")},
		}}},
		{`struct{*a "your it"}`, &StructType{Fields: []FieldDecl{
			{Type: &Star{Target: a}, Tag: strLit("your it")},
		}}},
		{`struct{big.Int "your it"}`, &StructType{Fields: []FieldDecl{
			{Type: bigInt, Tag: strLit("your it")},
		}}},
		{`struct{*big.Int "your it"}`, &StructType{Fields: []FieldDecl{
			{Type: &Star{Target: bigInt}, Tag: strLit("your it")},
		}}},
		{`struct{a "your it"; b}`, &StructType{Fields: []FieldDecl{
			{Type: a, Tag: strLit("your it")},
			{Type: b},
		}}},
		{`struct{a b "your it"}`, &StructType{Fields: []FieldDecl{
			{Identifiers: []Identifier{*a}, Type: b, Tag: strLit("your it")},
		}}},

		// Trailing ;
		{`struct{a;}`, &StructType{Fields: []FieldDecl{
			{Type: a},
		}}},
		{`struct{a; b; c;}`, &StructType{Fields: []FieldDecl{
			{Type: a},
			{Type: b},
			{Type: c},
		}}},

		// Embedded stars must be type names.
		{`struct{**big.Int}`, parseError{"expected.*got \\*"}},
		{`struct{*[]big.Int}`, parseError{"expected.*got \\["}},
	}.run(t, func(p *Parser) Node { return parseType(p) })
}

func TestParseInterfaceType(t *testing.T) {
	parserTests{
		{`interface{}`, &InterfaceType{}},
		{`interface{a; b; c}`, &InterfaceType{Methods: []Node{a, b, c}}},
		{`interface{a; b; big.Int}`, &InterfaceType{Methods: []Node{a, b, bigInt}}},
		{`interface{a; big.Int; b}`, &InterfaceType{Methods: []Node{a, bigInt, b}}},
		{`interface{big.Int; a; b}`, &InterfaceType{Methods: []Node{bigInt, a, b}}},
		{`interface{a; b; c()}`, &InterfaceType{Methods: []Node{a, b, &Method{Name: *c}}}},
		{`interface{a; b(); c}`, &InterfaceType{Methods: []Node{a, &Method{Name: *b}, c}}},
		{`interface{a(); b; c}`, &InterfaceType{Methods: []Node{&Method{Name: *a}, b, c}}},
		{`interface{a; big.Int; c()}`, &InterfaceType{Methods: []Node{a, bigInt, &Method{Name: *c}}}},

		// Trailing ;
		{`interface{a; b; c;}`, &InterfaceType{Methods: []Node{a, b, c}}},
	}.run(t, func(p *Parser) Node { return parseType(p) })
}

func TestParseFunctionType(t *testing.T) {
	parserTests{
		{`func()`, &FunctionType{}},
		{
			`func(a) b`,
			&FunctionType{Signature: Signature{
				ParameterList{Parameters: []ParameterDecl{
					{Type: a},
				}},
				ParameterList{Parameters: []ParameterDecl{
					{Type: b},
				}},
			}},
		},
		{
			`func(...a) b`,
			&FunctionType{Signature: Signature{
				ParameterList{Parameters: []ParameterDecl{
					{Type: a, DotDotDot: true},
				}},
				ParameterList{Parameters: []ParameterDecl{
					{Type: b},
				}},
			}},
		},
		{
			`func(a, b) c`,
			&FunctionType{Signature: Signature{
				ParameterList{Parameters: []ParameterDecl{
					{Type: a},
					{Type: b},
				}},
				ParameterList{Parameters: []ParameterDecl{
					{Type: c},
				}},
			}},
		},
		{
			`func(a) (b, c)`,
			&FunctionType{Signature: Signature{
				ParameterList{Parameters: []ParameterDecl{
					{Type: a},
				}},
				ParameterList{Parameters: []ParameterDecl{
					{Type: b},
					{Type: c},
				}},
			}},
		},
		{
			`func(...a) (b, c)`,
			&FunctionType{Signature: Signature{
				ParameterList{Parameters: []ParameterDecl{
					{Type: a, DotDotDot: true},
				}},
				ParameterList{Parameters: []ParameterDecl{
					{Type: b},
					{Type: c},
				}},
			}},
		},

		// Invalid, but will have to be caught during type checking.
		{
			`func(a) (b, ...c)`,
			&FunctionType{Signature: Signature{
				ParameterList{Parameters: []ParameterDecl{
					{Type: a},
				}},
				ParameterList{Parameters: []ParameterDecl{
					{Type: b},
					{Type: c, DotDotDot: true},
				}},
			}},
		},
	}.run(t, func(p *Parser) Node { return parseType(p) })
}

func TestParseParameterList(t *testing.T) {
	parserTests{
		{`()`, &ParameterList{}},

		// Parameter declarations without any identifiers.
		{
			`(a, b, c)`,
			&ParameterList{Parameters: []ParameterDecl{
				{Type: a},
				{Type: b},
				{Type: c},
			}},
		},
		{
			`(a, b, ...c)`,
			&ParameterList{Parameters: []ParameterDecl{
				{Type: a},
				{Type: b},
				{Type: c, DotDotDot: true},
			}},
		},
		{
			`(a, b, big.Int)`,
			&ParameterList{Parameters: []ParameterDecl{
				{Type: a},
				{Type: b},
				{Type: bigInt.(Type)},
			}},
		},
		{
			`(a, b, ...big.Int)`,
			&ParameterList{Parameters: []ParameterDecl{
				{Type: a},
				{Type: b},
				{Type: bigInt.(Type), DotDotDot: true},
			}},
		},
		{
			`(a, b, []c)`,
			&ParameterList{Parameters: []ParameterDecl{
				{Type: a},
				{Type: b},
				{Type: &SliceType{Type: c}},
			}},
		},
		{
			`(a, b, ...[]c)`,
			&ParameterList{Parameters: []ParameterDecl{
				{Type: a},
				{Type: b},
				{Type: &SliceType{Type: c}, DotDotDot: true},
			}},
		},
		{
			`([]a, b, c)`,
			&ParameterList{Parameters: []ParameterDecl{
				{Type: &SliceType{Type: a}},
				{Type: b},
				{Type: c},
			}},
		},
		{
			`([]a, b, ...c)`,
			&ParameterList{Parameters: []ParameterDecl{
				{Type: &SliceType{Type: a}},
				{Type: b},
				{Type: c, DotDotDot: true},
			}},
		},
		{
			`(...a)`,
			&ParameterList{Parameters: []ParameterDecl{
				{Type: a, DotDotDot: true},
			}},
		},

		// Parameter declarations with identifiers
		{
			`(a, b c)`,
			&ParameterList{Parameters: []ParameterDecl{
				{Identifiers: []Identifier{*a, *b}, Type: c},
			}},
		},
		{
			`(a, b ...c)`,
			&ParameterList{Parameters: []ParameterDecl{
				{Identifiers: []Identifier{*a, *b}, Type: c, DotDotDot: true},
			}},
		},
		{
			`(a, b big.Int)`,
			&ParameterList{Parameters: []ParameterDecl{
				{Identifiers: []Identifier{*a, *b}, Type: bigInt.(Type)},
			}},
		},
		{
			`(a, b []int)`,
			&ParameterList{Parameters: []ParameterDecl{
				{Identifiers: []Identifier{*a, *b}, Type: &SliceType{Type: id("int")}},
			}},
		},
		{
			`(a, b ...[]int)`,
			&ParameterList{Parameters: []ParameterDecl{
				{Identifiers: []Identifier{*a, *b}, Type: &SliceType{Type: id("int")}, DotDotDot: true},
			}},
		},
		{
			`(a, b []int, c, d ...[]int)`,
			&ParameterList{Parameters: []ParameterDecl{
				{Identifiers: []Identifier{*a, *b}, Type: &SliceType{Type: id("int")}},
				{Identifiers: []Identifier{*c, *d}, Type: &SliceType{Type: id("int")}, DotDotDot: true},
			}},
		},

		// Trailing comma is OK.
		{
			`(a, b, c,)`,
			&ParameterList{Parameters: []ParameterDecl{
				{Type: a},
				{Type: b},
				{Type: c},
			}},
		},
		{
			`(a, []b, c,)`,
			&ParameterList{Parameters: []ParameterDecl{
				{Type: a},
				{Type: &SliceType{Type: b}},
				{Type: c},
			}},
		},
		{
			`(a, b []int,)`,
			&ParameterList{Parameters: []ParameterDecl{
				{Identifiers: []Identifier{*a, *b}, Type: &SliceType{Type: id("int")}},
			}},
		},
		{
			`(a, b []int, c, d ...[]int,)`,
			&ParameterList{Parameters: []ParameterDecl{
				{Identifiers: []Identifier{*a, *b}, Type: &SliceType{Type: id("int")}},
				{Identifiers: []Identifier{*c, *d}, Type: &SliceType{Type: id("int")}, DotDotDot: true},
			}},
		},

		// Strange, but OK.
		{
			`(int float64, float64 int)`,
			&ParameterList{Parameters: []ParameterDecl{
				{Identifiers: []Identifier{*id("int")}, Type: id("float64")},
				{Identifiers: []Identifier{*id("float64")}, Type: id("int")},
			}},
		},

		// ... types must be the last in the list.
		{`(...a, b)`, parseError{""}},
		{`([]a, ...b, c)`, parseError{""}},
		{`(a, ...b, c)`, parseError{""}},
		{`(a ...b, c int)`, parseError{""}},

		// Can't mix declarations with identifiers with those without.
		{`([]a, b c)`, parseError{""}},
		{`(a b, c, d)`, parseError{""}},
	}.run(t, func(p *Parser) Node {
		parms := parseParameterList(p)
		return &parms
	})
}

func TestParseChannelType(t *testing.T) {
	parserTests{
		{`chan a`, &ChannelType{Send: true, Receive: true, Type: a}},
		{`<-chan a`, &ChannelType{Receive: true, Type: a}},
		{`chan<- a`, &ChannelType{Send: true, Type: a}},
		{`chan<- <- chan a`, &ChannelType{Send: true, Type: &ChannelType{Receive: true, Type: a}}},
		{`chan<- chan a`, &ChannelType{Send: true, Type: &ChannelType{Send: true, Receive: true, Type: a}}},
		{`<- chan <-chan a`, &ChannelType{Receive: true, Type: &ChannelType{Receive: true, Type: a}}},

		{`<- chan<- a`, parseError{"expected"}},
		{`chan<- <- a`, parseError{"expected"}},
		{`chan<- <- <- chan a`, parseError{"expected"}},
	}.run(t, func(p *Parser) Node { return parseType(p) })
}

func TestParseMapType(t *testing.T) {
	parserTests{
		{`map[int]a`, &MapType{Key: id("int"), Type: a}},
		{`map[*int]a.b`, &MapType{Key: &Star{Target: id("int")}, Type: sel(a, b).(Type)}},
		{`map[*int]map[string]int]`, &MapType{
			Key:  &Star{Target: id("int")},
			Type: &MapType{Key: id("string"), Type: id("int")},
		}},
	}.run(t, func(p *Parser) Node { return parseType(p) })
}

func TestParseArrayType(t *testing.T) {
	parserTests{
		{`[4]a`, &ArrayType{Size: intLit("4"), Type: a}},
		{`[4]a.b`, &ArrayType{Size: intLit("4"), Type: sel(a, b).(Type)}},
		{`[4](a)`, &ArrayType{Size: intLit("4"), Type: a}},
		{`[4][]a`, &ArrayType{Size: intLit("4"), Type: &SliceType{Type: a}}},
		{`[4][42*b]a`, &ArrayType{
			Size: intLit("4"),
			Type: &ArrayType{Size: binOp(token.Star, intLit("42"), b), Type: a},
		}},
		// [...]Type notation is only allowed in composite literals,
		// not types in general
		{`[...]int`, parseError{"expected.*got \\.\\.\\."}},
	}.run(t, func(p *Parser) Node { return parseType(p) })
}

func TestParseSliceType(t *testing.T) {
	parserTests{
		{`[]a`, &SliceType{Type: a}},
		{`[]a.b`, &SliceType{Type: sel(a, b).(Type)}},
		{`[](a)`, &SliceType{Type: a}},
		{`[][]a`, &SliceType{Type: &SliceType{Type: a}}},
	}.run(t, func(p *Parser) Node { return parseType(p) })
}

func TestParsePointerType(t *testing.T) {
	parserTests{
		{`*a`, &Star{Target: a}},
		{`*a.b`, &Star{Target: sel(a, b)}},
		{`*(a)`, &Star{Target: a}},
		{`**(a)`, &Star{Target: &Star{Target: a}}},

		{`α.`, parseError{"expected.*Identifier.*got EOF"}},
	}.run(t, func(p *Parser) Node { return parseType(p) })
}

func TestParseTypeName(t *testing.T) {
	parserTests{
		{`a`, a},
		{`a.b`, sel(a, b)},
		{`α.b`, sel(id("α"), b)},
	}.run(t, func(p *Parser) Node { return parseType(p) })
}

func TestParseCompositeLiteral(t *testing.T) {
	parserTests{
		{`struct{ a int }{ a: 4 }`, &CompositeLiteral{
			Type: &StructType{Fields: []FieldDecl{
				{Identifiers: []Identifier{*a}, Type: id("int")},
			}},
			Elements: []Element{
				{Key: a, Value: intLit("4")},
			},
		}},
		{`struct{ a, b int }{ a: 4, b: 5}`, &CompositeLiteral{
			Type: &StructType{Fields: []FieldDecl{
				{Identifiers: []Identifier{*a, *b}, Type: id("int")},
			}},
			Elements: []Element{
				{Key: a, Value: intLit("4")},
				{Key: b, Value: intLit("5")},
			},
		}},
		{`struct{ a []int }{ a: { 4, 5 } }`, &CompositeLiteral{
			Type: &StructType{Fields: []FieldDecl{
				{Identifiers: []Identifier{*a}, Type: &SliceType{Type: id("int")}},
			}},
			Elements: []Element{
				{Key: a, Value: &CompositeLiteral{
					Elements: []Element{
						{Value: intLit("4")}, {Value: intLit("5")},
					}},
				},
			},
		}},
		{`[][]int{ {4, 5} }`, &CompositeLiteral{
			Type: &SliceType{Type: &SliceType{Type: id("int")}},
			Elements: []Element{
				{Value: &CompositeLiteral{
					Elements: []Element{{Value: intLit("4")}, {Value: intLit("5")}},
				}},
			},
		}},
		{`[...]int{ 4, 5 }`, &CompositeLiteral{
			Type: &ArrayType{Type: id("int")},
			Elements: []Element{
				{Value: intLit("4")}, {Value: intLit("5")},
			},
		}},
		// Trailing ,
		{`struct{ a, b int }{ a: 4, b: 5,}`, &CompositeLiteral{
			Type: &StructType{Fields: []FieldDecl{
				{Identifiers: []Identifier{*a, *b}, Type: id("int")},
			}},
			Elements: []Element{
				{Key: a, Value: intLit("4")},
				{Key: b, Value: intLit("5")},
			},
		}},
		{`a{b: 5, c: 6}`, &CompositeLiteral{
			Type: a,
			Elements: []Element{
				{Key: b, Value: intLit("5")},
				{Key: c, Value: intLit("6")},
			},
		}},
	}.run(t, func(p *Parser) Node { return parseExpr(p) })
}

func TestParseTypeSwitchGuard(t *testing.T) {
	parserTests{
		{`a.(type)`, assert(a, nil)},
		{`a.(b).(type)`, assert(assert(a, b), nil)},
		{`a.b.(type)`, assert(sel(a, b), nil)},
		{`a[5].(type)`, assert(index(a, intLit("5")), nil)},

		// A type switch guard cannot be an operand.
		{`-a.(type)`, parseError{""}},
		{`5 * a.(type)`, parseError{""}},

		// This is OK.  It parses the type switch guard, leaving the *
		// as the next token.
		{`a.(type) * 5`, assert(a, nil)},
	}.run(t, func(p *Parser) Node { return parseExpression(p, true) })

	// Check that type switch guards aren't allowed by the
	// standard expression parser.
	parserTests{
		{`a.(type)`, parseError{"type"}},
		{`-a.(type)`, parseError{"type"}},
		{`5 * a.(type)`, parseError{"type"}},
		{`a.(type) * 5`, parseError{"type"}},
	}.run(t, func(p *Parser) Node { return parseExpr(p) })
}

func TestParseConversionExpr(t *testing.T) {
	parserTests{
		{`(int)(a)`, call(id("int"), false, a)},
		{`(struct{x int})(a)`, call(&StructType{
			Fields: []FieldDecl{{Identifiers: []Identifier{*id("x")}, Type: id("int")}},
		}, false, a)},
		{`(chan <- a)(b)`, call(&ChannelType{Send: true, Type: a}, false, b)},
		{`chan <- a(b)`, call(&ChannelType{Send: true, Type: a}, false, b)},
	}.run(t, func(p *Parser) Node { return parseExpr(p) })
}

func TestParseBuiltInCall(t *testing.T) {
	parserTests{
		{`make(chan <- a)`, call(id("make"), false, &ChannelType{Send: true, Type: a})},
		{`make(chan <- a, 5)`, call(id("make"), false, &ChannelType{Send: true, Type: a}, intLit("5"))},
		{`(make)(chan <- a, 5)`, call(id("make"), false, &ChannelType{Send: true, Type: a}, intLit("5"))},
	}.run(t, func(p *Parser) Node { return parseExpr(p) })
}

func TestParsePrimaryExpr(t *testing.T) {
	parserTests{
		// Operand
		{`5`, intLit("5")},

		// Call
		{`a()`, call(a, false)},
		{`a("bar")`, call(a, false, strLit("bar"))},
		{`a(b, c, d)`, call(a, false, b, c, d)},
		{`a(b, c, d...)`, call(a, true, b, c, d)},
		{`(a*b)(c, d)`, call(binOp(token.Star, a, b), false, c, d)},
		{`a(b*c-d)`, call(a, false, binOp(token.Minus, binOp(token.Star, b, c), d))},
		{`a(b*c, d)`, call(a, false, binOp(token.Star, b, c), d)},
		{`a(b(c), d)`, call(a, false, call(b, false, c), d)},
		{`math.Atan2(3.14/2, 0.5)`, call(sel(id("math"), id("Atan2")), false, binOp(token.Divide, floatLit("3.14"), intLit("2")), floatLit("0.5"))},

		// Selector
		{`a`, a},
		{`a.b`, sel(a, b)},
		{`a.b.c`, sel(sel(a, b), c)},
		{`a.b.c.d`, sel(sel(sel(a, b), c), d)},
		{`a(b).c`, sel(call(a, false, b), c)},
		{`a(b).c(d)`, call(sel(call(a, false, b), c), false, d)},

		// TypeAssertion
		{`a.(b)`, assert(a, b)},
		{`a.b.(c)`, assert(sel(a, b), c)},
		{`a.(b).(c)`, assert(assert(a, b), c)},
		{`a.(b).(c).d`, sel(assert(assert(a, b), c), d)},

		// Index
		{`a[b]`, index(a, b)},
		{`a[b][c]`, index(index(a, b), c)},
		{`a[b[c]]`, index(a, index(b, c))},

		// Slice
		{`a[:]`, slice(a, nil, nil, nil)},
		{`a[b:]`, slice(a, b, nil, nil)},
		{`a[:b]`, slice(a, nil, b, nil)},
		{`a[b:c:d]`, slice(a, b, c, d)},
		{`a[:b:c]`, slice(a, nil, b, c)},
		{`a[b[c]:d]`, slice(a, index(b, c), d, nil)},
		{`a[:b[c]:d]`, slice(a, nil, index(b, c), d)},
		{`a[:b:c[d]]`, slice(a, nil, b, index(c, d))},

		// Errors
		{`a[:b:]`, parseError{"expected operand"}},
		{`a[::b]`, parseError{"expected operand"}},
		{`a[5`, parseError{"expected"}},
		{`a.`, parseError{"expected.*OpenParen or Identifier"}},
		{`a[4`, parseError{"expected.*CloseBracket or Colon"}},

		// Disallow type switch guards outside of a type switch.
		{`a.(type)`, parseError{"type"}},
		{`a.(foo).(type)`, parseError{"type"}},
		{`a.b.(type)`, parseError{"type"}},
	}.run(t, func(p *Parser) Node { return parseExpr(p) })
}

func TestParseBinaryExpr(t *testing.T) {
	parserTests{
		{`a + b`, binOp(token.Plus, a, b)},
		{`a + b + c`, binOp(token.Plus, binOp(token.Plus, a, b), c)},
		{`a + b + c + d`, binOp(token.Plus, binOp(token.Plus, binOp(token.Plus, a, b), c), d)},
		{`a * b + c`, binOp(token.Plus, binOp(token.Star, a, b), c)},
		{`a + b * c`, binOp(token.Plus, a, binOp(token.Star, b, c))},
		{`a + -b`, binOp(token.Plus, a, unOp(token.Minus, b))},
		{`-a + b`, binOp(token.Plus, unOp(token.Minus, a), b)},
		{`a || b && c`, binOp(token.OrOr, a, binOp(token.AndAnd, b, c))},
		{`a && b || c`, binOp(token.OrOr, binOp(token.AndAnd, a, b), c)},
		{`a && b == c`, binOp(token.AndAnd, a, binOp(token.EqualEqual, b, c))},
		{`a == b && c`, binOp(token.AndAnd, binOp(token.EqualEqual, a, b), c)},
		{`a && b != c`, binOp(token.AndAnd, a, binOp(token.BangEqual, b, c))},
		{`a && b < c`, binOp(token.AndAnd, a, binOp(token.Less, b, c))},
		{`a && b <= c`, binOp(token.AndAnd, a, binOp(token.LessEqual, b, c))},
		{`a && b > c`, binOp(token.AndAnd, a, binOp(token.Greater, b, c))},
		{`a && b >= c`, binOp(token.AndAnd, a, binOp(token.GreaterEqual, b, c))},
		{`(a + b) * c`, binOp(token.Star, binOp(token.Plus, a, b), c)},
		{`(a || b) && c`, binOp(token.AndAnd, binOp(token.OrOr, a, b), c)},
	}.run(t, func(p *Parser) Node { return parseExpr(p) })
}

func TestParseUnaryExpr(t *testing.T) {
	parserTests{
		{`+a`, unOp(token.Plus, a)},
		{`-a.b`, unOp(token.Minus, sel(a, b))},
		{`!0`, unOp(token.Bang, intLit("0"))},
		{`^(a)`, unOp(token.Carrot, a)},
		{`*5.1`, &Star{Target: floatLit("5.1")}},
		{`&!1`, unOp(token.And, unOp(token.Bang, intLit("1")))},
		{`<-a`, unOp(token.LessMinus, a)},
		{`<-!-a`, unOp(token.LessMinus, unOp(token.Bang, unOp(token.Minus, a)))},
	}.run(t, func(p *Parser) Node { return parseExpr(p) })
}

func TestParseIdentifier(t *testing.T) {
	parserTests{
		{"_abc123", id("_abc123")},
		{"_αβξ123", id("_αβξ123")},
	}.run(t, func(p *Parser) Node { return parseExpr(p) })
}

func TestParseIntegerLiteral(t *testing.T) {
	parserTests{
		{"1", intLit("1")},
		{"010", intLit("8")},
		{"0x10", intLit("16")},
		{"08", parseError{"malformed.*integer"}},
	}.run(t, func(p *Parser) Node { return parseExpr(p) })
}

func TestParseFloatLiteral(t *testing.T) {
	parserTests{
		{"1.", floatLit("1.0")},
		{"1.0", floatLit("1.0")},
		{"0.1", floatLit("0.1")},
		{"0.1000", floatLit("0.1")},
		{"1e1", floatLit("10.0")},
		{"1e-1", floatLit("0.1")},
	}.run(t, func(p *Parser) Node { return parseExpr(p) })
}

func TestParseImaginaryLiteral(t *testing.T) {
	i := func(s string) Node {
		var r big.Rat
		r.SetString(s)
		return &ImaginaryLiteral{Value: &r}
	}
	parserTests{
		{"0.i", i("0.0i")},
		{"1.i", i("1.0i")},
		{"1.0i", i("1.0i")},
		{"0.1i", i("0.1i")},
		{"0.1000i", i("0.1i")},
		{"1e1i", i("10.0i")},
		{"1e-1i", i("0.1i")},
	}.run(t, func(p *Parser) Node { return parseExpr(p) })
}

func TestParseStringLiteral(t *testing.T) {
	parserTests{
		{`""`, strLit("")},
		{`"abc"`, strLit("abc")},
		{`"αβξ"`, strLit("αβξ")},
		{`"\x0A""`, strLit("\n")},
		{`"\n\r\t\v\""`, strLit("\n\r\t\v\"")},

		// Cannot have a newline in an interpreted string.
		{"\x22\x0A\x22", parseError{"unexpected"}},
		// Cannot escape a single quote in an interpreted string lit.
		{`"\'""`, parseError{"unexpected.*'"}},

		{"\x60\x60", strLit("")},
		{"\x60\x5C\x60", strLit("\\")},
		{"\x60\x0A\x60", strLit("\n")},
		{"\x60abc\x60", strLit("abc")},
		// Strip \r from raw string literals.
		{"\x60\x0D\x60", strLit("")},
		{"\x60αβξ\x60", strLit("αβξ")},
	}.run(t, func(p *Parser) Node { return parseExpr(p) })
}

func TestParseRuneLiteral(t *testing.T) {
	parserTests{
		{`'a'`, runeLit('a')},
		{`'α'`, runeLit('α')},
		{`'\''`, runeLit('\'')},
		{`'\000'`, runeLit('\000')},
		{`'\x00'`, runeLit('\x00')},
		{`'\xFF'`, runeLit('\xFF')},
		{`'\u00FF'`, runeLit('\u00FF')},
		{`'\U000000FF'`, runeLit('\U000000FF')},
		{`'\U0010FFFF'`, runeLit('\U0010FFFF')},

		{`'\"'`, parseError{"unexpected.*\""}},
		{`'\008'`, parseError{"unexpected"}},
		{`'\U00110000'`, parseError{"malformed"}},
	}.run(t, func(p *Parser) Node { return parseExpr(p) })
}

type parserTest struct {
	text string
	node Node
}

func (test parserTest) run(t *testing.T, production func(*Parser) Node) {
	n, err := parse(NewParser(token.NewLexer("", test.text)), production)
	if pe, ok := test.node.(parseError); ok {
		if err == nil {
			t.Errorf("parse(%s): expected error matching %s, got\n%s", test.text, pe.re, str(n))
		} else if !regexp.MustCompile(pe.re).MatchString(err.Error()) {
			t.Errorf("parse(%s): expected an error matching %s, got %s", test.text, pe.re, err.Error())
		}
		return
	}
	if err != nil {
		t.Errorf("parse(%s): unexpected error: %s", test.text, err.Error())
	}
	if !eq.Deep(n, test.node) {
		t.Errorf("parse(%s): %s,\nexpected: %s", test.text, str(n), str(test.node))
	}
}

func parse(p *Parser, production func(*Parser) Node) (n Node, err error) {
	defer func() {
		r := recover()
		if r == nil {
			return
		}
		switch e := r.(type) {
		case *SyntaxError:
			err = e
		case *MalformedLiteral:
			err = e
		default:
			panic(r)
		}

	}()
	n = production(p)
	return n, err
}

func str(u interface{}) string {
	buf := bytes.NewBuffer(nil)
	if err := pp.Print(buf, u); err != nil {
		panic(err)
	}
	return buf.String()
}

// A special AST Node, denoting that a test expects to have a parse error.
type parseError struct{ re string }

func (p parseError) Start() token.Location { panic("unimplemented") }
func (p parseError) End() token.Location   { panic("unimplemented") }

type parserTests []parserTest

func (tests parserTests) run(t *testing.T, production func(*Parser) Node) {
	for _, test := range tests {
		test.run(t, production)
	}
}
