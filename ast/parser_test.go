package ast

import (
	"go/parser"
	stdtoken "go/token"
	"math/big"
	"reflect"
	"regexp"
	"testing"

	"github.com/eaburns/eq"
	"github.com/eaburns/pp"
	"github.com/velour/stop/test"
	"github.com/velour/stop/token"
)

var (
	a, b, c, d             = id("a"), id("b"), id("c"), id("d")
	t0, t1, t2, t3, t4, t5 = id("T0"), id("T1"), id("T2"), id("T3"), id("T4"), id("T5")
	aStmt                  = &ExpressionStmt{Expression: a}
	bStmt                  = &ExpressionStmt{Expression: b}
	cStmt                  = &ExpressionStmt{Expression: c}
	dStmt                  = &ExpressionStmt{Expression: d}
	oneStmt                = &ExpressionStmt{Expression: intLit("1")}
	twoStmt                = &ExpressionStmt{Expression: intLit("2")}
	threeStmt              = &ExpressionStmt{Expression: intLit("3")}
	fourStmt               = &ExpressionStmt{Expression: intLit("4")}
	x, y, z                = id("x"), id("y"), id("z")
	bigInt                 = sel(id("big"), id("Int")).(Type)
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

func id(s string) *Identifier { return &Identifier{Name: s} }
func ids(ss ...string) []Identifier {
	ids := make([]Identifier, len(ss))
	for i, s := range ss {
		ids[i] = *id(s)
	}
	return ids
}
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
		p = &Selector{Parent: p, Identifier: i}
	}
	return p
}

var specDeclarationTests = parserTests{
	{`type T1 string`, Declarations{&TypeSpec{Identifier: *t1, Type: id("string")}}},
	{`type T2 T1`, Declarations{&TypeSpec{Identifier: *t2, Type: t1}}},
	{`type T3 []T1`, Declarations{&TypeSpec{Identifier: *t3, Type: &SliceType{Type: t1}}}},
	{`type T4 T3`, Declarations{&TypeSpec{Identifier: *t4, Type: t3}}},
	{
		`type Lock interface {
			Lock()
			Unlock()
		}`,
		Declarations{
			&TypeSpec{
				Identifier: *id("Lock"),
				Type: &InterfaceType{Methods: []Node{
					&Method{Identifier: *id("Lock")},
					&Method{Identifier: *id("Unlock")},
				}},
			},
		},
	},
	{
		`type ReadWrite interface {
			Read(b Buffer) bool
			Write(b Buffer) bool
		}`,
		Declarations{
			&TypeSpec{
				Identifier: *id("ReadWrite"),
				Type: &InterfaceType{Methods: []Node{
					&Method{
						Identifier: *id("Read"),
						Signature: Signature{
							Parameters: []ParameterDecl{{Identifier: b, Type: id("Buffer")}},
							Results:    []ParameterDecl{{Type: id("bool")}},
						},
					},
					&Method{
						Identifier: *id("Write"),
						Signature: Signature{
							Parameters: []ParameterDecl{{Identifier: b, Type: id("Buffer")}},
							Results:    []ParameterDecl{{Type: id("bool")}},
						},
					},
				}},
			},
		},
	},
	{
		`type File interface {
			ReadWrite
			Lock
			Close()
		}`,
		Declarations{
			&TypeSpec{
				Identifier: *id("File"),
				Type: &InterfaceType{Methods: []Node{
					id("ReadWrite"),
					id("Lock"),
					&Method{Identifier: *id("Close")},
				}},
			},
		},
	},
	{
		`type (
			T0 []string
			T1 []string
			T2 struct{ a, b int }
			T3 struct{ a, c int }
			T4 func(int, float64) *T0
			T5 func(x int, y float64) *[]string
		)`,
		Declarations{
			&TypeSpec{Identifier: *t0, Type: &SliceType{Type: id("string")}},
			&TypeSpec{Identifier: *t1, Type: &SliceType{Type: id("string")}},
			&TypeSpec{
				Identifier: *t2,
				Type: &StructType{Fields: []FieldDecl{
					{Identifier: a, Type: id("int")},
					{Identifier: b, Type: id("int")},
				}},
			},
			&TypeSpec{
				Identifier: *t3,
				Type: &StructType{Fields: []FieldDecl{
					{Identifier: a, Type: id("int")},
					{Identifier: c, Type: id("int")},
				}},
			},
			&TypeSpec{
				Identifier: *t4,
				Type: &FunctionType{Signature: Signature{
					Parameters: []ParameterDecl{
						{Type: id("int")}, {Type: id("float64")},
					},
					Results: []ParameterDecl{
						{Type: &Star{Target: t0}},
					},
				}},
			},
			&TypeSpec{
				Identifier: *t5,
				Type: &FunctionType{Signature: Signature{
					Parameters: []ParameterDecl{
						{Identifier: x, Type: id("int")},
						{Identifier: y, Type: id("float64")},
					},
					Results: []ParameterDecl{
						{Type: &Star{Target: &SliceType{Type: id("string")}}},
					},
				}},
			},
		},
	},
	{
		`const Pi float64 = 3.14159265358979323846`,
		Declarations{
			&ConstSpec{
				Type:        id("float64"),
				Identifiers: ids("Pi"),
				Values:      []Expression{floatLit("3.14159265358979323846")},
			},
		},
	},
	{
		`const zero = 0.0`,
		Declarations{
			&ConstSpec{
				Identifiers: ids("zero"),
				Values:      []Expression{floatLit("0.0")},
			},
		},
	},
	{
		`const (
			size int64 = 1024
			eof        = -1
		)`,
		Declarations{
			&ConstSpec{
				Type:        id("int64"),
				Identifiers: ids("size"),
				Values:      []Expression{intLit("1024")},
			},
			&ConstSpec{
				Identifiers: ids("eof"),
				Values:      []Expression{unOp(token.Minus, intLit("1"))},
			},
		},
	},
	{
		`const a, b, c = 3, 4, "foo"`,
		Declarations{
			&ConstSpec{
				Identifiers: ids("a", "b", "c"),
				Values:      []Expression{intLit("3"), intLit("4"), strLit("foo")},
			},
		},
	},
	{
		`const u, v float32 = 0, 3`,
		Declarations{
			&ConstSpec{
				Type:        id("float32"),
				Identifiers: ids("u", "v"),
				Values:      []Expression{intLit("0"), intLit("3")},
			},
		},
	},
	{
		`const (
			Sunday = iota
			Monday
			Tuesday
			Wednesday
			Thursday
			Friday
			Partyday
			numberOfDays
		)`,
		Declarations{
			&ConstSpec{
				Identifiers: ids("Sunday"),
				Values:      []Expression{id("iota")},
			},
			&ConstSpec{Identifiers: ids("Monday")},
			&ConstSpec{Identifiers: ids("Tuesday")},
			&ConstSpec{Identifiers: ids("Wednesday")},
			&ConstSpec{Identifiers: ids("Thursday")},
			&ConstSpec{Identifiers: ids("Friday")},
			&ConstSpec{Identifiers: ids("Partyday")},
			&ConstSpec{Identifiers: ids("numberOfDays")},
		},
	},
	{
		`const (
			c0 = iota
			c1 = iota
			c2 = iota
		)`,
		Declarations{
			&ConstSpec{
				Identifiers: ids("c0"),
				Values:      []Expression{id("iota")},
			},
			&ConstSpec{
				Identifiers: ids("c1"),
				Values:      []Expression{id("iota")},
			},
			&ConstSpec{
				Identifiers: ids("c2"),
				Values:      []Expression{id("iota")},
			},
		},
	},
	{
		`const (
			a = 1 << iota
			b = 1 << iota
			c = 1 << iota
		)`,
		Declarations{
			&ConstSpec{
				Identifiers: ids("a"),
				Values:      []Expression{binOp(token.LessLess, intLit("1"), id("iota"))},
			},
			&ConstSpec{
				Identifiers: ids("b"),
				Values:      []Expression{binOp(token.LessLess, intLit("1"), id("iota"))},
			},
			&ConstSpec{
				Identifiers: ids("c"),
				Values:      []Expression{binOp(token.LessLess, intLit("1"), id("iota"))},
			},
		},
	},
	{
		`const (
			u         = iota * 42
			v float64 = iota * 42
			w         = iota * 42
		)`,
		Declarations{
			&ConstSpec{
				Identifiers: ids("u"),
				Values:      []Expression{binOp(token.Star, id("iota"), intLit("42"))},
			},
			&ConstSpec{
				Identifiers: ids("v"),
				Type:        id("float64"),
				Values:      []Expression{binOp(token.Star, id("iota"), intLit("42"))},
			},
			&ConstSpec{
				Identifiers: ids("w"),
				Values:      []Expression{binOp(token.Star, id("iota"), intLit("42"))},
			},
		},
	},
	{
		`const (
			bit0, mask0 = 1 << iota, 1<<iota - 1
			bit1, mask1
			_, _
			bit3, mask3
		)`,
		Declarations{
			&ConstSpec{
				Identifiers: ids("bit0", "mask0"),
				Values: []Expression{
					binOp(token.LessLess, intLit("1"), id("iota")),
					binOp(token.Minus,
						binOp(token.LessLess, intLit("1"), id("iota")),
						intLit("1")),
				},
			},
			&ConstSpec{Identifiers: ids("bit1", "mask1")},
			&ConstSpec{Identifiers: ids("_", "_")},
			&ConstSpec{Identifiers: ids("bit3", "mask3")},
		},
	},
	{
		`type IntArray [16]int`,
		Declarations{
			&TypeSpec{
				Identifier: *id("IntArray"),
				Type: &ArrayType{
					Size: intLit("16"),
					Type: id("int"),
				},
			},
		},
	},
	{
		`type (
			Point struct{ x, y float64 }
			Polar Point
		)`,
		Declarations{
			&TypeSpec{
				Identifier: *id("Point"),
				Type: &StructType{Fields: []FieldDecl{
					{Identifier: x, Type: id("float64")},
					{Identifier: y, Type: id("float64")},
				}},
			},
			&TypeSpec{Identifier: *id("Polar"), Type: id("Point")},
		},
	},
	{
		`type TreeNode struct {
			left, right *TreeNode
			value *Comparable
		}`,
		Declarations{
			&TypeSpec{
				Identifier: *id("TreeNode"),
				Type: &StructType{Fields: []FieldDecl{
					{Identifier: id("left"), Type: &Star{Target: id("TreeNode")}},
					{Identifier: id("right"), Type: &Star{Target: id("TreeNode")}},
					{Identifier: id("value"), Type: &Star{Target: id("Comparable")}},
				}},
			},
		},
	},
	{
		`type Block interface {
			BlockSize() int
			Encrypt(src, dst []byte)
			Decrypt(src, dst []byte)
		}`,
		Declarations{
			&TypeSpec{
				Identifier: *id("Block"),
				Type: &InterfaceType{Methods: []Node{
					&Method{
						Identifier: *id("BlockSize"),
						Signature: Signature{
							Results: []ParameterDecl{{Type: id("int")}},
						},
					},
					&Method{
						Identifier: *id("Encrypt"),
						Signature: Signature{
							Parameters: []ParameterDecl{
								{Identifier: id("src"), Type: &SliceType{Type: id("byte")}},
								{Identifier: id("dst"), Type: &SliceType{Type: id("byte")}},
							},
						},
					},
					&Method{
						Identifier: *id("Decrypt"),
						Signature: Signature{
							Parameters: []ParameterDecl{
								{Identifier: id("src"), Type: &SliceType{Type: id("byte")}},
								{Identifier: id("dst"), Type: &SliceType{Type: id("byte")}},
							},
						},
					},
				}},
			},
		},
	},
}

var specTypeTests = parserTests{
	{`[32]byte`, &ArrayType{Size: intLit("32"), Type: id("byte")}},
	{
		`[2*N] struct { x, y int32 }`,
		&ArrayType{
			Size: binOp(token.Star, intLit("2"), id("N")),
			Type: &StructType{
				Fields: []FieldDecl{
					{Identifier: x, Type: id("int32")},
					{Identifier: y, Type: id("int32")},
				},
			},
		},
	},
	{
		`[1000]*float64`,
		&ArrayType{Size: intLit("1000"), Type: &Star{Target: id("float64")}},
	},
	{
		`[3][5]int`,
		&ArrayType{
			Size: intLit("3"),
			Type: &ArrayType{Size: intLit("5"), Type: id("int")},
		},
	},
	{
		`[2][2][2]float64`,
		&ArrayType{
			Size: intLit("2"),
			Type: &ArrayType{
				Size: intLit("2"),
				Type: &ArrayType{
					Size: intLit("2"),
					Type: id("float64"),
				},
			},
		},
	},
	{
		`[2]([2]([2]float64))`,
		&ArrayType{
			Size: intLit("2"),
			Type: &ArrayType{
				Size: intLit("2"),
				Type: &ArrayType{
					Size: intLit("2"),
					Type: id("float64"),
				},
			},
		},
	},
	{`struct {}`, &StructType{}},
	{
		`struct {
				x, y int
				u float32
				_ float32  // padding
				A *[]int
				F func()
			}`,
		&StructType{
			Fields: []FieldDecl{
				{Identifier: x, Type: id("int")},
				{Identifier: y, Type: id("int")},
				{Identifier: id("u"), Type: id("float32")},
				{Identifier: id("_"), Type: id("float32")},
				{Identifier: id("A"), Type: &Star{Target: &SliceType{Type: id("int")}}},
				{Identifier: id("F"), Type: &FunctionType{}},
			},
		},
	},
	{
		`struct {
				T1        // field name is T1
				*T2       // field name is T2
				P.T3      // field name is T3
				*P.T4     // field name is T4
				x, y int  // field names are x and y
			}`,
		&StructType{
			Fields: []FieldDecl{
				{Type: t1},
				{Type: &Star{Target: t2}},
				{Type: sel(id("P"), t3).(Type)},
				{Type: &Star{Target: sel(id("P"), t4).(Type)}},
				{Identifier: x, Type: id("int")},
				{Identifier: y, Type: id("int")},
			},
		},
	},
	{
		// Illegal, but it should parse.
		`struct {
				T     // conflicts with anonymous field *T and *P.T
				*T    // conflicts with anonymous field T and *P.T
				*P.T  // conflicts with anonymous field T and *T
			}`,
		&StructType{
			Fields: []FieldDecl{
				{Type: id("T")},
				{Type: &Star{Target: id("T")}},
				{Type: &Star{Target: sel(id("P"), id("T")).(Type)}},
			},
		},
	},
	{
		`struct {
				microsec  uint64 "field 1"
				serverIP6 uint64 "field 2"
				process   string "field 3"
			}`,
		&StructType{
			Fields: []FieldDecl{
				{
					Identifier: id("microsec"),
					Type:       id("uint64"),
					Tag:        strLit("field 1"),
				},
				{
					Identifier: id("serverIP6"),
					Type:       id("uint64"),
					Tag:        strLit("field 2"),
				},
				{
					Identifier: id("process"),
					Type:       id("string"),
					Tag:        strLit("field 3"),
				},
			},
		},
	},
	{`*Point`, &Star{Target: id("Point")}},
	{`*[4]int`, &Star{Target: &ArrayType{Size: intLit("4"), Type: id("int")}}},
	{`func()`, &FunctionType{}},
	{
		`func(x int) int`,
		&FunctionType{Signature: Signature{
			Parameters: []ParameterDecl{{Identifier: x, Type: id("int")}},
			Results:    []ParameterDecl{{Type: id("int")}},
		}},
	},
	{
		`func(a, _ int, z float32) bool`,
		&FunctionType{Signature: Signature{
			Parameters: []ParameterDecl{
				{Identifier: a, Type: id("int")},
				{Identifier: id("_"), Type: id("int")},
				{Identifier: z, Type: id("float32")},
			},
			Results: []ParameterDecl{{Type: id("bool")}},
		}},
	},
	{
		`func(a, b int, z float32) (bool)`,
		&FunctionType{Signature: Signature{
			Parameters: []ParameterDecl{
				{Identifier: a, Type: id("int")},
				{Identifier: b, Type: id("int")},
				{Identifier: z, Type: id("float32")},
			},
			Results: []ParameterDecl{{Type: id("bool")}},
		}},
	},
	{
		`func(prefix string, values ...int)`,
		&FunctionType{Signature: Signature{
			Parameters: []ParameterDecl{
				{Identifier: id("prefix"), Type: id("string")},
				{Identifier: id("values"), DotDotDot: true, Type: id("int")},
			},
		}},
	},
	{
		`func(a, b int, z float64, opt ...interface{}) (success bool)`,
		&FunctionType{Signature: Signature{
			Parameters: []ParameterDecl{
				{Identifier: a, Type: id("int")},
				{Identifier: b, Type: id("int")},
				{Identifier: z, Type: id("float64")},
				{Identifier: id("opt"), DotDotDot: true, Type: &InterfaceType{}},
			},
			Results: []ParameterDecl{{Identifier: id("success"), Type: id("bool")}},
		}},
	},
	{
		`func(int, int, float64) (float64, *[]int)`,
		&FunctionType{Signature: Signature{
			Parameters: []ParameterDecl{
				{Type: id("int")}, {Type: id("int")}, {Type: id("float64")},
			},
			Results: []ParameterDecl{
				{Type: id("float64")},
				{Type: &Star{Target: &SliceType{Type: id("int")}}},
			},
		}},
	},
	{
		`func(n int) func(p *T)`,
		&FunctionType{Signature: Signature{
			Parameters: []ParameterDecl{{Identifier: id("n"), Type: id("int")}},
			Results: []ParameterDecl{
				{
					Type: &FunctionType{Signature: Signature{
						Parameters: []ParameterDecl{
							{Identifier: id("p"), Type: &Star{Target: id("T")}},
						},
					}},
				},
			},
		}},
	},
	{
		`interface {
				Read(b Buffer) bool
				Write(b Buffer) bool
				Close()
			}`,
		&InterfaceType{Methods: []Node{
			&Method{
				Identifier: *id("Read"),
				Signature: Signature{
					Parameters: []ParameterDecl{{Identifier: b, Type: id("Buffer")}},
					Results:    []ParameterDecl{{Type: id("bool")}},
				},
			},
			&Method{
				Identifier: *id("Write"),
				Signature: Signature{
					Parameters: []ParameterDecl{{Identifier: b, Type: id("Buffer")}},
					Results:    []ParameterDecl{{Type: id("bool")}},
				},
			},
			&Method{Identifier: *id("Close")},
		}},
	},
	{`interface{}`, &InterfaceType{}},
	{`map[string]int`, &MapType{Key: id("string"), Value: id("int")}},
	{
		`map[*T]struct{ x, y float64 }`,
		&MapType{
			Key: &Star{Target: id("T")},
			Value: &StructType{Fields: []FieldDecl{
				{Identifier: x, Type: id("float64")},
				{Identifier: y, Type: id("float64")},
			}},
		},
	},
	{`map[string]interface{}`, &MapType{Key: id("string"), Value: &InterfaceType{}}},
	{`chan T`, &ChannelType{Send: true, Receive: true, Type: id("T")}},
	{`chan<- float64`, &ChannelType{Send: true, Type: id("float64")}},
	{`<-chan int`, &ChannelType{Receive: true, Type: id("int")}},
	{
		`chan<- chan int`,
		&ChannelType{
			Send: true,
			Type: &ChannelType{Send: true, Receive: true, Type: id("int")},
		},
	},
	{
		`chan<- <-chan int`,
		&ChannelType{
			Send: true,
			Type: &ChannelType{Receive: true, Type: id("int")},
		},
	},
	{
		`<-chan <-chan int`,
		&ChannelType{
			Receive: true,
			Type:    &ChannelType{Receive: true, Type: id("int")},
		},
	},
	{
		`chan (<-chan int)`,
		&ChannelType{
			Send:    true,
			Receive: true,
			Type:    &ChannelType{Receive: true, Type: id("int")},
		},
	},
}

// These are freebie examples given to us right from the Go spec itself.
func TestParseSpecExamples(t *testing.T) {
	specDeclarationTests.run(t, func(p *Parser) Node { return parseDeclarations(p) })
	specTypeTests.run(t, func(p *Parser) Node { return parseType(p) })
}

func TestParseSourceFile(t *testing.T) {
	parserTests{
		{`package main`, &SourceFile{PackageName: *id("main")}},
		{
			`
				package main
				import "fmt"
			`,
			&SourceFile{
				PackageName: *id("main"),
				Imports: []ImportDecl{
					{Imports: []ImportSpec{{Path: *strLit("fmt")}}},
				},
			},
		},
		{
			`
				package main
				import "fmt"
				func main() {}
			`,
			&SourceFile{
				PackageName: *id("main"),
				Imports: []ImportDecl{
					{Imports: []ImportSpec{{Path: *strLit("fmt")}}},
				},
				Declarations: Declarations{
					&FunctionDecl{
						Identifier: *id("main"),
					},
				},
			},
		},
	}.run(t, func(p *Parser) Node { return parseSourceFile(p) })
}

func TestParseImportDecl(t *testing.T) {
	parserTests{
		{
			`import "fmt"`,
			&ImportDecl{
				Imports: []ImportSpec{
					{Path: *strLit("fmt")},
				},
			},
		},
		{
			`import . "fmt"`,
			&ImportDecl{
				Imports: []ImportSpec{
					{Dot: true, Path: *strLit("fmt")},
				},
			},
		},
		{
			`import f "fmt"`,
			&ImportDecl{
				Imports: []ImportSpec{
					{Identifier: id("f"), Path: *strLit("fmt")},
				},
			},
		},
		{
			`import (
				"fmt"
			)`,
			&ImportDecl{
				Imports: []ImportSpec{
					{Path: *strLit("fmt")},
				},
			},
		},
		{
			`import (
				. "fmt"
			)`,
			&ImportDecl{
				Imports: []ImportSpec{
					{Dot: true, Path: *strLit("fmt")},
				},
			},
		},
		{
			`import (
				f "fmt"
			)`,
			&ImportDecl{
				Imports: []ImportSpec{
					{Identifier: id("f"), Path: *strLit("fmt")},
				},
			},
		},
		{
			`import (
				f "fmt"
				o "os";
				. "github.com/eaburns/pp"
				"code.google.com/p/plotinum"
			)`,
			&ImportDecl{
				Imports: []ImportSpec{
					{Identifier: id("f"), Path: *strLit("fmt")},
					{Identifier: id("o"), Path: *strLit("os")},
					{Dot: true, Path: *strLit("github.com/eaburns/pp")},
					{Path: *strLit("code.google.com/p/plotinum")},
				},
			},
		},
		{
			`import ( "fmt"; "os" )`,
			&ImportDecl{
				Imports: []ImportSpec{
					{Path: *strLit("fmt")},
					{Path: *strLit("os")},
				},
			},
		},
		{`import ( "fmt" "os" )`, parseError{";"}},
	}.run(t, func(p *Parser) Node { return parseImportDecl(p) })
}

func TestParseStatements(t *testing.T) {
	parserTests{
		{``, nil},
		{`;`, nil},
		{`fallthrough`, &FallthroughStmt{}},
		{`goto a`, &GotoStmt{Label: *a}},
		{`break`, &BreakStmt{}},
		{`break a`, &BreakStmt{Label: a}},
		{`continue`, &ContinueStmt{}},
		{`continue a`, &ContinueStmt{Label: a}},
		{`return`, &ReturnStmt{}},
		{`return a, b, c`, &ReturnStmt{Expressions: []Expression{a, b, c}}},
		{`go a()`, &GoStmt{Expression: call(a, false)}},
		{`defer a()`, &DeferStmt{Expression: call(a, false)}},

		// Range is disallowed outside of a for loop.
		{`a := range b`, parseError{"range"}},

		{
			`{continue; break}`,
			&BlockStmt{
				Statements: []Statement{
					&ContinueStmt{},
					&BreakStmt{},
				},
			},
		},
		{`{continue break}`, parseError{";"}},
	}.run(t, func(p *Parser) Node { return parseStatement(p) })
}

func TestParseSelect(t *testing.T) {
	parserTests{
		{`select{}`, &Select{}},
		{
			`select{ case a <- b: c() }`,
			&Select{
				Cases: []CommCase{
					{
						Send: &SendStmt{Channel: a, Expression: b},
						Statements: []Statement{
							&ExpressionStmt{Expression: call(c, false)},
						},
					},
				},
			},
		},
		{
			`select{ case a := <- b: c() }`,
			&Select{
				Cases: []CommCase{
					{
						Receive: &RecvStmt{
							Op:    token.ColonEqual,
							Left:  []Expression{a},
							Right: *unOp(token.LessMinus, b),
						},
						Statements: []Statement{
							&ExpressionStmt{Expression: call(c, false)},
						},
					},
				},
			},
		},
		{
			`select{ case a, b := <- c: d() }`,
			&Select{
				Cases: []CommCase{
					{
						Receive: &RecvStmt{
							Op:    token.ColonEqual,
							Left:  []Expression{a, b},
							Right: *unOp(token.LessMinus, c),
						},
						Statements: []Statement{
							&ExpressionStmt{Expression: call(d, false)},
						},
					},
				},
			},
		},
		{
			`select{ case a, b = <- c: d() }`,
			&Select{
				Cases: []CommCase{
					{
						Receive: &RecvStmt{
							Op:    token.Equal,
							Left:  []Expression{a, b},
							Right: *unOp(token.LessMinus, c),
						},
						Statements: []Statement{
							&ExpressionStmt{Expression: call(d, false)},
						},
					},
				},
			},
		},
		{
			`select{ case a() = <- b: c() }`,
			&Select{
				Cases: []CommCase{
					{
						Receive: &RecvStmt{
							Op:    token.Equal,
							Left:  []Expression{call(a, false)},
							Right: *unOp(token.LessMinus, b),
						},
						Statements: []Statement{
							&ExpressionStmt{Expression: call(c, false)},
						},
					},
				},
			},
		},
		{
			`select{ default: a() }`,
			&Select{
				Cases: []CommCase{
					{
						Statements: []Statement{
							&ExpressionStmt{Expression: call(a, false)},
						},
					},
				},
			},
		},
		{
			`select{
				case a <- b:
					c()
				case a() = <- b:
					c()
					d()
				default:
					a()
			}`,
			&Select{
				Cases: []CommCase{
					{
						Send: &SendStmt{Channel: a, Expression: b},
						Statements: []Statement{
							&ExpressionStmt{Expression: call(c, false)},
						},
					},
					{
						Receive: &RecvStmt{
							Op:    token.Equal,
							Left:  []Expression{call(a, false)},
							Right: *unOp(token.LessMinus, b),
						},
						Statements: []Statement{
							&ExpressionStmt{Expression: call(c, false)},
							&ExpressionStmt{Expression: call(d, false)},
						},
					},
					{
						Statements: []Statement{
							&ExpressionStmt{Expression: call(a, false)},
						},
					},
				},
			},
		},

		// := cannot appear after non-identifiers.
		{`select{ case a() := <- b: c() }`, parseError{":="}},
		{`select{ case a, a() := <- b: c() }`, parseError{":="}},

		// Only a receive expression can appear in a receive statement.
		{`select{ case a := b: c() }`, parseError{"<-"}},
		{`select{ case a = b: c() }`, parseError{"<-"}},
		{`select{ case a, b = *d: c() }`, parseError{"<-"}},
	}.run(t, func(p *Parser) Node { return parseStatement(p) })
}

func TestParseSwitch(t *testing.T) {
	parserTests{
		{`switch {}`, &ExprSwitch{}},
		{`switch a {}`, &ExprSwitch{Expression: a}},
		{`switch ; a {}`, &ExprSwitch{Expression: a}},
		{
			`switch a(b, c) {}`,
			&ExprSwitch{Expression: call(a, false, b, c)},
		},
		{
			`switch a(); b {}`,
			&ExprSwitch{
				Initialization: &ExpressionStmt{Expression: call(a, false)},
				Expression:     b,
			},
		},
		{
			`switch a = b; c {}`,
			&ExprSwitch{
				Initialization: &Assignment{
					Op:    token.Equal,
					Left:  []Expression{a},
					Right: []Expression{b},
				},
				Expression: c,
			},
		},
		{
			`switch a := b; c {}`,
			&ExprSwitch{
				Initialization: &ShortVarDecl{
					Left:  ids("a"),
					Right: []Expression{b},
				},
				Expression: c,
			},
		},
		{
			`switch a, b := b, a; c {}`,
			&ExprSwitch{
				Initialization: &ShortVarDecl{
					Left:  ids("a", "b"),
					Right: []Expression{b, a},
				},
				Expression: c,
			},
		},
		{
			`switch a, b := b, a; a(b, c) {}`,
			&ExprSwitch{
				Initialization: &ShortVarDecl{
					Left:  ids("a", "b"),
					Right: []Expression{b, a},
				},
				Expression: call(a, false, b, c),
			},
		},
		{
			`switch { case a: b() }`,
			&ExprSwitch{
				Cases: []ExprCase{
					{
						Expressions: []Expression{a},
						Statements: []Statement{
							&ExpressionStmt{Expression: call(b, false)},
						},
					},
				},
			},
		},
		{
			`switch { case a, 5: b() }`,
			&ExprSwitch{
				Cases: []ExprCase{
					{
						Expressions: []Expression{a, intLit("5")},
						Statements: []Statement{
							&ExpressionStmt{Expression: call(b, false)},
						},
					},
				},
			},
		},
		{
			`switch { default: a() }`,
			&ExprSwitch{
				Cases: []ExprCase{
					{
						Statements: []Statement{
							&ExpressionStmt{Expression: call(a, false)},
						},
					},
				},
			},
		},
		{
			`switch { 
				case a, 5:
					b()
					c = d
				case true, false:
					d()
					fallthrough
				default:
					return 42
			}`,
			&ExprSwitch{
				Cases: []ExprCase{
					{
						Expressions: []Expression{a, intLit("5")},
						Statements: []Statement{
							&ExpressionStmt{Expression: call(b, false)},
							&Assignment{
								Op:    token.Equal,
								Left:  []Expression{c},
								Right: []Expression{d},
							},
						},
					},
					{
						Expressions: []Expression{id("true"), id("false")},
						Statements: []Statement{
							&ExpressionStmt{Expression: call(d, false)},
							&FallthroughStmt{},
						},
					},
					{
						Statements: []Statement{
							&ReturnStmt{
								Expressions: []Expression{intLit("42")},
							},
						},
					},
				},
			},
		},

		{`switch a.(type) {}`, &TypeSwitch{Expression: a}},
		{`switch ; a.(type) {}`, &TypeSwitch{Expression: a}},
		{`switch a.b.(type) {}`, &TypeSwitch{Expression: sel(a, b)}},
		{`switch a(b).(type) {}`, &TypeSwitch{Expression: call(a, false, b)}},
		{`switch a.(b).(type) {}`, &TypeSwitch{Expression: assert(a, b)}},
		{`switch a := b.(type) {}`, &TypeSwitch{Declaration: a, Expression: b}},
		{
			`switch a(); b := c.(type) {}`,
			&TypeSwitch{
				Initialization: &ExpressionStmt{Expression: call(a, false)},
				Declaration:    b,
				Expression:     c,
			},
		},
		{
			`switch a.(type) { case big.Int: b() }`,
			&TypeSwitch{
				Expression: a,
				Cases: []TypeCase{
					{
						Types: []Type{bigInt},
						Statements: []Statement{
							&ExpressionStmt{Expression: call(b, false)},
						},
					},
				},
			},
		},
		{
			`switch a.(type) { case big.Int, float64: b() }`,
			&TypeSwitch{
				Expression: a,
				Cases: []TypeCase{
					{
						Types: []Type{bigInt, id("float64")},
						Statements: []Statement{
							&ExpressionStmt{Expression: call(b, false)},
						},
					},
				},
			},
		},
		{
			`switch a.(type) { default: b() }`,
			&TypeSwitch{
				Expression: a,
				Cases: []TypeCase{
					{
						Statements: []Statement{
							&ExpressionStmt{Expression: call(b, false)},
						},
					},
				},
			},
		},
		{
			`switch a.(type) { 
				case big.Int, float64:
					b()
					c = d
				case interface{}:
					d()
					fallthrough
				default:
					return 42
			}`,
			&TypeSwitch{
				Expression: a,
				Cases: []TypeCase{
					{
						Types: []Type{bigInt, id("float64")},
						Statements: []Statement{
							&ExpressionStmt{Expression: call(b, false)},
							&Assignment{
								Op:    token.Equal,
								Left:  []Expression{c},
								Right: []Expression{d},
							},
						},
					},
					{
						Types: []Type{&InterfaceType{}},
						Statements: []Statement{
							&ExpressionStmt{Expression: call(d, false)},
							&FallthroughStmt{},
						},
					},
					{
						Statements: []Statement{
							&ReturnStmt{Expressions: []Expression{intLit("42")}},
						},
					},
				},
			},
		},

		// Bad type switches.
		{`switch a.(type); b.(type) {}`, parseError{""}},
		{`switch a, b := c.(type) {}`, parseError{""}},
		{`switch a := b, c.(type) {}`, parseError{""}},
		{`switch a := b.(type), c {}`, parseError{""}},
		{`switch a = b.(type) {}`, parseError{""}},

		// Switches and composite literals.
		{
			`switch (struct{}{}) {}`,
			&ExprSwitch{
				Expression: &CompositeLiteral{Type: &StructType{}},
			},
		},
		{
			`switch (struct{}{}); 5 {}`,
			&ExprSwitch{
				Initialization: &ExpressionStmt{
					Expression: &CompositeLiteral{Type: &StructType{}},
				},
				Expression: intLit("5"),
			},
		},
		{
			`switch (struct{}{}).(type) {}`,
			&TypeSwitch{
				Expression: &CompositeLiteral{Type: &StructType{}},
			},
		},
	}.run(t, func(p *Parser) Node { return parseStatement(p) })
}

func TestParseFor(t *testing.T) {
	parserTests{
		{
			`for { a }`,
			&ForStmt{
				Block: BlockStmt{Statements: []Statement{aStmt}},
			},
		},
		{
			`for a { b }`,
			&ForStmt{
				Condition: a,
				Block:     BlockStmt{Statements: []Statement{bStmt}},
			},
		},
		{
			`for a := 1; b; a++ { c }`,
			&ForStmt{
				Initialization: &ShortVarDecl{
					Left:  ids("a"),
					Right: []Expression{intLit("1")},
				},
				Condition: b,
				Post:      &IncDecStmt{Op: token.PlusPlus, Expression: a},
				Block:     BlockStmt{Statements: []Statement{cStmt}},
			},
		},
		{
			`for ; b; a++ { c }`,
			&ForStmt{
				Condition: b,
				Post:      &IncDecStmt{Op: token.PlusPlus, Expression: a},
				Block:     BlockStmt{Statements: []Statement{cStmt}},
			},
		},
		{
			`for a := 1; ; a++ { c }`,
			&ForStmt{
				Initialization: &ShortVarDecl{
					Left:  ids("a"),
					Right: []Expression{intLit("1")},
				},
				Post:  &IncDecStmt{Op: token.PlusPlus, Expression: a},
				Block: BlockStmt{Statements: []Statement{cStmt}},
			},
		},
		{
			`for a := 1; b; { c }`,
			&ForStmt{
				Initialization: &ShortVarDecl{
					Left:  ids("a"),
					Right: []Expression{intLit("1")},
				},
				Condition: b,
				Block:     BlockStmt{Statements: []Statement{cStmt}},
			},
		},
		{
			`for ; ; { c }`,
			&ForStmt{
				Block: BlockStmt{Statements: []Statement{cStmt}},
			},
		},
		{
			`for a := range b { c }`,
			&ForStmt{
				Range: &ShortVarDecl{
					Left:  ids("a"),
					Right: []Expression{b},
				},
				Block: BlockStmt{Statements: []Statement{cStmt}},
			},
		},
		{
			`for a, b := range c { d }`,
			&ForStmt{
				Range: &ShortVarDecl{
					Left:  ids("a", "b"),
					Right: []Expression{c},
				},
				Block: BlockStmt{Statements: []Statement{dStmt}},
			},
		},
		{
			`for a = range b { c }`,
			&ForStmt{
				Range: &Assignment{
					Op:    token.Equal,
					Left:  []Expression{a},
					Right: []Expression{b},
				},
				Block: BlockStmt{Statements: []Statement{cStmt}},
			},
		},
		{
			`for a, b = range c { d }`,
			&ForStmt{
				Range: &Assignment{
					Op:    token.Equal,
					Left:  []Expression{a, b},
					Right: []Expression{c},
				},
				Block: BlockStmt{Statements: []Statement{dStmt}},
			},
		},

		// Range is unexpected with any assign op other that =.
		{`for a *= range b { c }`, parseError{"range"}},
		{`for a, b *= range c { d }`, parseError{"range"}},

		// Only a single expression allow on LHS of a range.
		{`for a := range b, c { d }`, parseError{""}},
		{`for a, b := range c, d { 1 }`, parseError{""}},
		{`for a = range b, c { d }`, parseError{""}},
		{`for a, b = range c, d { 1 }`, parseError{""}},

		// Labels are not a simple statement.
		{`for label:; a < 100; a++`, parseError{":"}},

		// For loops and composite literals.
		{
			`for (struct{}{}) {}`,
			&ForStmt{
				Condition: &CompositeLiteral{Type: &StructType{}},
			},
		},
	}.run(t, func(p *Parser) Node { return parseStatement(p) })
}

func TestParseIf(t *testing.T) {
	parserTests{
		{
			`if a { b }`,
			&IfStmt{
				Condition: a,
				Block:     BlockStmt{Statements: []Statement{bStmt}},
			},
		},
		{
			`if a; b { c }`,
			&IfStmt{
				Statement: aStmt,
				Condition: b,
				Block:     BlockStmt{Statements: []Statement{cStmt}},
			},
		},
		{
			`if a; b { c } else { d }`,
			&IfStmt{
				Statement: aStmt,
				Condition: b,
				Block:     BlockStmt{Statements: []Statement{cStmt}},
				Else:      &BlockStmt{Statements: []Statement{dStmt}},
			},
		},
		{
			`if a; b { c } else if d { 1 }`,
			&IfStmt{
				Statement: aStmt,
				Condition: b,
				Block: BlockStmt{
					Statements: []Statement{cStmt},
				},
				Else: &IfStmt{
					Condition: d,
					Block:     BlockStmt{Statements: []Statement{oneStmt}},
				},
			},
		},
		{
			`if a { 1 } else if b { 2 } else if c { 3 } else { 4 }`,
			&IfStmt{
				Condition: a,
				Block: BlockStmt{
					Statements: []Statement{oneStmt},
				},
				Else: &IfStmt{
					Condition: b,
					Block: BlockStmt{
						Statements: []Statement{twoStmt},
					},
					Else: &IfStmt{
						Condition: c,
						Block: BlockStmt{
							Statements: []Statement{threeStmt},
						},
						Else: &BlockStmt{
							Statements: []Statement{fourStmt},
						},
					},
				},
			},
		},

		// If statements and composite literals.
		{
			`if (struct{}{}){}`,
			&IfStmt{
				Condition: &CompositeLiteral{Type: &StructType{}},
				Block:     BlockStmt{},
			},
		},
	}.run(t, func(p *Parser) Node { return parseStatement(p) })
}

func TestParseBlock(t *testing.T) {
	parserTests{
		{`{}`, &BlockStmt{}},
		{
			`{{{}}}`,
			&BlockStmt{
				Statements: []Statement{
					&BlockStmt{Statements: []Statement{&BlockStmt{}}},
				},
			},
		},
		{
			`{ a = b; c = d; }`,
			&BlockStmt{
				Statements: []Statement{
					&Assignment{
						Op:    token.Equal,
						Left:  []Expression{a},
						Right: []Expression{b},
					},
					&Assignment{
						Op:    token.Equal,
						Left:  []Expression{c},
						Right: []Expression{d},
					},
				},
			},
		},
		{
			`{ a = b; c = d }`,
			&BlockStmt{
				Statements: []Statement{
					&Assignment{
						Op:    token.Equal,
						Left:  []Expression{a},
						Right: []Expression{b},
					},
					&Assignment{
						Op:    token.Equal,
						Left:  []Expression{c},
						Right: []Expression{d},
					},
				},
			},
		},
		{
			`{
				a = b
				c = d
			}`,
			&BlockStmt{
				Statements: []Statement{
					&Assignment{
						Op:    token.Equal,
						Left:  []Expression{a},
						Right: []Expression{b},
					},
					&Assignment{
						Op:    token.Equal,
						Left:  []Expression{c},
						Right: []Expression{d},
					},
				},
			},
		},
	}.run(t, func(p *Parser) Node { return parseStatement(p) })
}

func TestParseDeclarationStmt(t *testing.T) {
	parserTests{
		{
			`const a = 5`,
			&DeclarationStmt{
				Declarations: Declarations{
					&ConstSpec{
						Identifiers: ids("a"),
						Values:      []Expression{intLit("5")},
					},
				},
			},
		},
		{
			`const a big.Int = 5`,
			&DeclarationStmt{
				Declarations: Declarations{
					&ConstSpec{
						Type:        bigInt,
						Identifiers: ids("a"),
						Values:      []Expression{intLit("5")},
					},
				},
			},
		},
		{
			`var(
				a, b big.Int = 5, 6
				c = 7
				d = 8.0
			)`,
			&DeclarationStmt{
				Declarations: Declarations{
					&VarSpec{
						Type:        bigInt,
						Identifiers: ids("a", "b"),
						Values:      []Expression{intLit("5"), intLit("6")},
					},
					&VarSpec{
						Identifiers: ids("c"),
						Values:      []Expression{intLit("7")},
					},
					&VarSpec{
						Identifiers: ids("d"),
						Values:      []Expression{floatLit("8.0")},
					},
				},
			},
		},
	}.run(t, func(p *Parser) Node { return parseStatement(p) })
}

func TestParseLabeledStmt(t *testing.T) {
	parserTests{
		{
			`here: a = b`,
			&LabeledStmt{
				Label: *id("here"),
				Statement: &Assignment{
					Op:    token.Equal,
					Left:  []Expression{a},
					Right: []Expression{b},
				},
			},
		},
		{
			`here: there: a := b`,
			&LabeledStmt{
				Label: *id("here"),
				Statement: &LabeledStmt{
					Label: *id("there"),
					Statement: &ShortVarDecl{
						Left:  ids("a"),
						Right: []Expression{b},
					},
				},
			},
		},
	}.run(t, func(p *Parser) Node { return parseStatement(p) })
}

func TestParseShortVarDecl(t *testing.T) {
	parserTests{
		{
			`a := 5`,
			&ShortVarDecl{
				Left:  ids("a"),
				Right: []Expression{intLit("5")},
			},
		},
		{
			`a, b := 5, 6`,
			&ShortVarDecl{
				Left:  ids("a", "b"),
				Right: []Expression{intLit("5"), intLit("6")},
			},
		},
		// Only allow idents on LHS

		// The following is parsed as an expression statement with
		// a trailing := left for the next parse call.
		//{`a.b := 1`, parseError{"expected"}},
		{`a, b.c := 1, 2`, parseError{"expected"}},
		{`a := b.(type)`, parseError{"type"}},
		{`a, b := c.(type), 5`, parseError{"type"}},
		{`a, b := 5, c.(type)`, parseError{"type"}},
	}.run(t, func(p *Parser) Node { return parseStatement(p) })
}

func TestParseAssignment(t *testing.T) {
	parserTests{
		{
			`a = b`,
			&Assignment{
				Op:    token.Equal,
				Left:  []Expression{a},
				Right: []Expression{b},
			},
		},
		{
			`a += b`,
			&Assignment{
				Op:    token.PlusEqual,
				Left:  []Expression{a},
				Right: []Expression{b},
			},
		},
		{
			`a -= b`,
			&Assignment{
				Op:    token.MinusEqual,
				Left:  []Expression{a},
				Right: []Expression{b},
			},
		},
		{
			`a |= b`,
			&Assignment{
				Op:    token.OrEqual,
				Left:  []Expression{a},
				Right: []Expression{b},
			},
		},
		{
			`a ^= b`,
			&Assignment{
				Op:    token.CarrotEqual,
				Left:  []Expression{a},
				Right: []Expression{b},
			},
		},
		{
			`a *= b`,
			&Assignment{
				Op:    token.StarEqual,
				Left:  []Expression{a},
				Right: []Expression{b},
			},
		},
		{
			`a /= b`,
			&Assignment{
				Op:    token.DivideEqual,
				Left:  []Expression{a},
				Right: []Expression{b},
			},
		},
		{
			`a %= b`,
			&Assignment{
				Op:    token.PercentEqual,
				Left:  []Expression{a},
				Right: []Expression{b},
			},
		},
		{
			`a <<= b`,
			&Assignment{
				Op:    token.LessLessEqual,
				Left:  []Expression{a},
				Right: []Expression{b},
			},
		},
		{
			`a >>= b`,
			&Assignment{
				Op:    token.GreaterGreaterEqual,
				Left:  []Expression{a},
				Right: []Expression{b},
			},
		},
		{
			`a &= b`,
			&Assignment{
				Op:    token.AndEqual,
				Left:  []Expression{a},
				Right: []Expression{b},
			},
		},
		{
			`a &^= b`,
			&Assignment{
				Op:    token.AndCarrotEqual,
				Left:  []Expression{a},
				Right: []Expression{b},
			},
		},

		{
			`a, b = c, d`,
			&Assignment{
				Op:    token.Equal,
				Left:  []Expression{a, b},
				Right: []Expression{c, d},
			},
		},
		{
			`a.b, c, d *= 5, 6, 7`,
			&Assignment{
				Op:    token.StarEqual,
				Left:  []Expression{sel(a, b), c, d},
				Right: []Expression{intLit("5"), intLit("6"), intLit("7")},
			},
		},
	}.run(t, func(p *Parser) Node { return parseStatement(p) })
}

func TestParseExpressionStmt(t *testing.T) {
	parserTests{
		{`a`, &ExpressionStmt{Expression: a}},
		{`b`, &ExpressionStmt{Expression: b}},
		{`a.b`, &ExpressionStmt{Expression: sel(a, b)}},
		{`a[5]`, &ExpressionStmt{Expression: index(a, intLit("5"))}},
	}.run(t, func(p *Parser) Node { return parseStatement(p) })
}

func TestParseIncDecStmt(t *testing.T) {
	inc, dec := token.PlusPlus, token.MinusMinus
	parserTests{
		{`a++`, &IncDecStmt{Op: inc, Expression: a}},
		{`b--`, &IncDecStmt{Op: dec, Expression: b}},
		{`a[5]++`, &IncDecStmt{Op: inc, Expression: index(a, intLit("5"))}},
	}.run(t, func(p *Parser) Node { return parseStatement(p) })
}

func TestParseSendStmt(t *testing.T) {
	parserTests{
		{`a <- b`, &SendStmt{Channel: a, Expression: b}},
		{`a <- 5`, &SendStmt{Channel: a, Expression: intLit("5")}},
		{
			`a[6] <- b[7]`,
			&SendStmt{
				Channel:    index(a, intLit("6")),
				Expression: index(b, intLit("7")),
			},
		},
	}.run(t, func(p *Parser) Node { return parseStatement(p) })
}

func TestParseMethodDecl(t *testing.T) {
	parserTests{
		{
			`func (a b) method(c, d big.Int) big.Int { return c }`,
			Declarations{
				&MethodDecl{
					Receiver:     *a,
					BaseTypeName: *b,
					Identifier:   *id("method"),
					Signature: Signature{
						Parameters: []ParameterDecl{
							{Identifier: c, Type: bigInt},
							{Identifier: d, Type: bigInt},
						},
						Results: []ParameterDecl{{Type: bigInt}},
					},
					Body: BlockStmt{
						Statements: []Statement{
							&ReturnStmt{
								Expressions: []Expression{c},
							},
						},
					},
				},
			},
		},
		{
			`func (a *b) method() {}`,
			Declarations{
				&MethodDecl{
					Receiver:     *a,
					Pointer:      true,
					BaseTypeName: *b,
					Identifier:   *id("method"),
				},
			},
		},
	}.run(t, func(p *Parser) Node { return parseTopLevelDecl(p) })
}

func TestParseFunctionDecl(t *testing.T) {
	parserTests{
		{
			`func function(a, b big.Int) big.Int { return a }`,
			Declarations{
				&FunctionDecl{
					Identifier: *id("function"),
					Signature: Signature{
						Parameters: []ParameterDecl{
							{Identifier: a, Type: bigInt},
							{Identifier: b, Type: bigInt},
						},
						Results: []ParameterDecl{{Type: bigInt}},
					},
					Body: BlockStmt{
						Statements: []Statement{
							&ReturnStmt{
								Expressions: []Expression{a},
							},
						},
					},
				},
			},
		},
	}.run(t, func(p *Parser) Node { return parseTopLevelDecl(p) })
}

func TestParseVarDecl(t *testing.T) {
	parserTests{
		{
			`var a big.Int`,
			Declarations{
				&VarSpec{Identifiers: ids("a"), Type: bigInt},
			},
		},
		{
			`var a int = 5`,
			Declarations{
				&VarSpec{
					Identifiers: ids("a"),
					Type:        id("int"),
					Values:      []Expression{intLit("5")},
				},
			},
		},
		{
			`var a, b int = 5, 6`,
			Declarations{
				&VarSpec{
					Identifiers: ids("a", "b"),
					Type:        id("int"),
					Values:      []Expression{intLit("5"), intLit("6")},
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
					Identifiers: ids("a"),
					Type:        id("int"),
					Values:      []Expression{intLit("7")},
				},
				&VarSpec{
					Identifiers: ids("b"),
					Values:      []Expression{floatLit("3.14")},
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
					Identifiers: ids("a"),
					Type:        id("int"),
					Values:      []Expression{intLit("7")},
				},
				&VarSpec{
					Identifiers: ids("b", "c", "d"),
					Values:      []Expression{intLit("12"), intLit("13"), intLit("14")},
				},
				&VarSpec{
					Identifiers: ids("x", "y", "z"),
					Type:        id("float64"),
					Values:      []Expression{floatLit("3.0"), floatLit("4.0"), floatLit("5.0")},
				},
			},
		},

		// If there is no type then there must be an expr list.
		{`var b`, parseError{"expected"}},
	}.run(t, func(p *Parser) Node { return parseTopLevelDecl(p) })
}

func TestParseConstDecl(t *testing.T) {
	parserTests{
		{
			`const a`,
			Declarations{
				&ConstSpec{Identifiers: ids("a")},
			},
		},
		{
			`const a big.Int`,
			Declarations{
				&ConstSpec{Identifiers: ids("a"), Type: bigInt},
			},
		},
		{
			`const a int = 5`,
			Declarations{
				&ConstSpec{
					Identifiers: ids("a"),
					Type:        id("int"),
					Values:      []Expression{intLit("5")},
				},
			},
		},
		{
			`const a, b int = 5, 6`,
			Declarations{
				&ConstSpec{
					Identifiers: ids("a", "b"),
					Type:        id("int"),
					Values:      []Expression{intLit("5"), intLit("6")},
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
					Identifiers: ids("a"),
					Type:        id("int"),
					Values:      []Expression{intLit("7")},
				},
				&ConstSpec{Identifiers: ids("b")},
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
					Identifiers: ids("a"),
					Type:        id("int"),
					Values:      []Expression{intLit("7")},
				},
				&ConstSpec{Identifiers: ids("b", "c", "d")},
				&ConstSpec{
					Identifiers: ids("x", "y", "z"),
					Type:        id("float64"),
					Values:      []Expression{floatLit("3.0"), floatLit("4.0"), floatLit("5.0")},
				},
			},
		},
	}.run(t, func(p *Parser) Node { return parseTopLevelDecl(p) })
}

func TestParseTypeDecl(t *testing.T) {
	parserTests{
		{
			`type a big.Int`,
			Declarations{
				&TypeSpec{Identifier: *a, Type: bigInt},
			},
		},
		{
			`type (
				a big.Int
				b struct{}
			)`,
			Declarations{
				&TypeSpec{Identifier: *a, Type: bigInt},
				&TypeSpec{Identifier: *b, Type: &StructType{}},
			},
		},

		{
			`type ( a big.Int; b struct{} )`,
			Declarations{
				&TypeSpec{Identifier: *a, Type: bigInt},
				&TypeSpec{Identifier: *b, Type: &StructType{}},
			},
		},
		{`type ( a big.Int b struct{} )`, parseError{";"}},
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
		{`map[a]b`, &MapType{Key: a, Value: b}},

		{`[]func()`, &SliceType{Type: &FunctionType{}}},
		{
			`[]func()<-chan int`,
			&SliceType{
				Type: &FunctionType{Signature: Signature{
					Results: []ParameterDecl{
						{Type: &ChannelType{Receive: true, Type: id("int")}},
					},
				}},
			},
		},
		{
			`[]interface{ c()<-chan[5]big.Int }`,
			&SliceType{
				Type: &InterfaceType{Methods: []Node{
					&Method{
						Identifier: *c,
						Signature: Signature{
							Results: []ParameterDecl{
								{Type: &ChannelType{
									Receive: true,
									Type: &ArrayType{
										Size: intLit("5"),
										Type: bigInt,
									},
								}},
							},
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
			{Identifier: a, Type: c},
			{Identifier: b, Type: c},
		}}},
		{`struct{a b; c}`, &StructType{Fields: []FieldDecl{
			{Identifier: a, Type: b},
			{Type: c},
		}}},
		{`struct{a; b c}`, &StructType{Fields: []FieldDecl{
			{Type: a},
			{Identifier: b, Type: c},
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
			{Identifier: a, Type: b},
		}}},
		{`struct{big.Int; a big.Int}`, &StructType{Fields: []FieldDecl{
			{Type: bigInt},
			{Identifier: a, Type: bigInt},
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
			{Identifier: a, Type: b, Tag: strLit("your it")},
		}}},
		{`struct{a, b c "your it"}`, &StructType{Fields: []FieldDecl{
			{Identifier: a, Type: c, Tag: strLit("your it")},
			{Identifier: b, Type: c, Tag: strLit("your it")},
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

		{
			`struct {a int; b int}`,
			&StructType{Fields: []FieldDecl{
				{Identifier: a, Type: id("int")},
				{Identifier: b, Type: id("int")},
			},
			},
		},
		{`struct {a int b int}`, parseError{";"}},
	}.run(t, func(p *Parser) Node { return parseType(p) })
}

func TestParseInterfaceType(t *testing.T) {
	parserTests{
		{`interface{}`, &InterfaceType{}},
		{`interface{a; b; c}`, &InterfaceType{Methods: []Node{a, b, c}}},
		{`interface{a; b; big.Int}`, &InterfaceType{Methods: []Node{a, b, bigInt}}},
		{`interface{a; big.Int; b}`, &InterfaceType{Methods: []Node{a, bigInt, b}}},
		{`interface{big.Int; a; b}`, &InterfaceType{Methods: []Node{bigInt, a, b}}},
		{`interface{a; b; c()}`, &InterfaceType{Methods: []Node{a, b, &Method{Identifier: *c}}}},
		{`interface{a; b(); c}`, &InterfaceType{Methods: []Node{a, &Method{Identifier: *b}, c}}},
		{`interface{a(); b; c}`, &InterfaceType{Methods: []Node{&Method{Identifier: *a}, b, c}}},
		{`interface{a; big.Int; c()}`, &InterfaceType{Methods: []Node{a, bigInt, &Method{Identifier: *c}}}},

		// Trailing ;
		{`interface{a; b; c;}`, &InterfaceType{Methods: []Node{a, b, c}}},

		{
			`interface{ a(); b()}`,
			&InterfaceType{
				Methods: []Node{
					&Method{Identifier: *a},
					&Method{Identifier: *b},
				},
			},
		},
		{`interface{ a() b()}`, parseError{";"}},
	}.run(t, func(p *Parser) Node { return parseType(p) })
}

func TestParseFunctionType(t *testing.T) {
	parserTests{
		{`func()`, &FunctionType{}},
		{
			`func(a) b`,
			&FunctionType{Signature: Signature{
				Parameters: []ParameterDecl{{Type: a}},
				Results:    []ParameterDecl{{Type: b}},
			}},
		},
		{
			`func(...a) b`,
			&FunctionType{Signature: Signature{
				Parameters: []ParameterDecl{{Type: a, DotDotDot: true}},
				Results:    []ParameterDecl{{Type: b}},
			}},
		},
		{
			`func(a, b) c`,
			&FunctionType{Signature: Signature{
				Parameters: []ParameterDecl{{Type: a}, {Type: b}},
				Results:    []ParameterDecl{{Type: c}},
			}},
		},
		{
			`func(a) (b, c)`,
			&FunctionType{Signature: Signature{
				Parameters: []ParameterDecl{{Type: a}},
				Results:    []ParameterDecl{{Type: b}, {Type: c}},
			}},
		},
		{
			`func(...a) (b, c)`,
			&FunctionType{Signature: Signature{
				Parameters: []ParameterDecl{{Type: a, DotDotDot: true}},
				Results:    []ParameterDecl{{Type: b}, {Type: c}},
			}},
		},

		// Invalid, but will have to be caught during type checking.
		{
			`func(a) (b, ...c)`,
			&FunctionType{Signature: Signature{
				Parameters: []ParameterDecl{{Type: a}},
				Results:    []ParameterDecl{{Type: b}, {Type: c, DotDotDot: true}},
			}},
		},
	}.run(t, func(p *Parser) Node { return parseType(p) })
}

// Tests parameter list parsing via parseSignature, because there is no AST node for
// a parameter list itself.
func TestParseParameterList(t *testing.T) {
	parserTests{
		// Parameter declarations without any identifiers.
		{
			`(a, b, c)`,
			&Signature{Parameters: []ParameterDecl{
				{Type: a},
				{Type: b},
				{Type: c},
			}},
		},
		{
			`(a, b, ...c)`,
			&Signature{Parameters: []ParameterDecl{
				{Type: a},
				{Type: b},
				{Type: c, DotDotDot: true},
			}},
		},
		{
			`(a, b, big.Int)`,
			&Signature{Parameters: []ParameterDecl{
				{Type: a},
				{Type: b},
				{Type: bigInt.(Type)},
			}},
		},
		{
			`(a, b, ...big.Int)`,
			&Signature{Parameters: []ParameterDecl{
				{Type: a},
				{Type: b},
				{Type: bigInt.(Type), DotDotDot: true},
			}},
		},
		{
			`(a, b, []c)`,
			&Signature{Parameters: []ParameterDecl{
				{Type: a},
				{Type: b},
				{Type: &SliceType{Type: c}},
			}},
		},
		{
			`(a, b, ...[]c)`,
			&Signature{Parameters: []ParameterDecl{
				{Type: a},
				{Type: b},
				{Type: &SliceType{Type: c}, DotDotDot: true},
			}},
		},
		{
			`([]a, b, c)`,
			&Signature{Parameters: []ParameterDecl{
				{Type: &SliceType{Type: a}},
				{Type: b},
				{Type: c},
			}},
		},
		{
			`([]a, b, ...c)`,
			&Signature{Parameters: []ParameterDecl{
				{Type: &SliceType{Type: a}},
				{Type: b},
				{Type: c, DotDotDot: true},
			}},
		},
		{
			`(...a)`,
			&Signature{Parameters: []ParameterDecl{
				{Type: a, DotDotDot: true},
			}},
		},

		// Parameter declarations with identifiers
		{
			`(a, b c)`,
			&Signature{Parameters: []ParameterDecl{
				{Identifier: a, Type: c},
				{Identifier: b, Type: c},
			}},
		},
		{
			`(a, b ...c)`,
			&Signature{Parameters: []ParameterDecl{
				{Identifier: a, Type: c},
				{Identifier: b, Type: c, DotDotDot: true},
			}},
		},
		{
			`(a, b big.Int)`,
			&Signature{Parameters: []ParameterDecl{
				{Identifier: a, Type: bigInt.(Type)},
				{Identifier: b, Type: bigInt.(Type)},
			}},
		},
		{
			`(a, b []int)`,
			&Signature{Parameters: []ParameterDecl{
				{Identifier: a, Type: &SliceType{Type: id("int")}},
				{Identifier: b, Type: &SliceType{Type: id("int")}},
			}},
		},
		{
			`(a, b ...[]int)`,
			&Signature{Parameters: []ParameterDecl{
				{Identifier: a, Type: &SliceType{Type: id("int")}},
				{Identifier: b, Type: &SliceType{Type: id("int")}, DotDotDot: true},
			}},
		},
		{
			`(a, b []int, c, d ...[]int)`,
			&Signature{Parameters: []ParameterDecl{
				{Identifier: a, Type: &SliceType{Type: id("int")}},
				{Identifier: b, Type: &SliceType{Type: id("int")}},
				{Identifier: c, Type: &SliceType{Type: id("int")}},
				{Identifier: d, Type: &SliceType{Type: id("int")}, DotDotDot: true},
			}},
		},

		// Trailing comma is OK.
		{
			`(a, b, c,)`,
			&Signature{Parameters: []ParameterDecl{
				{Type: a},
				{Type: b},
				{Type: c},
			}},
		},
		{
			`(a, []b, c,)`,
			&Signature{Parameters: []ParameterDecl{
				{Type: a},
				{Type: &SliceType{Type: b}},
				{Type: c},
			}},
		},
		{
			`(a, b []int,)`,
			&Signature{Parameters: []ParameterDecl{
				{Identifier: a, Type: &SliceType{Type: id("int")}},
				{Identifier: b, Type: &SliceType{Type: id("int")}},
			}},
		},
		{
			`(a, b []int, c, d ...[]int,)`,
			&Signature{Parameters: []ParameterDecl{
				{Identifier: a, Type: &SliceType{Type: id("int")}},
				{Identifier: b, Type: &SliceType{Type: id("int")}},
				{Identifier: c, Type: &SliceType{Type: id("int")}},
				{Identifier: d, Type: &SliceType{Type: id("int")}, DotDotDot: true},
			}},
		},

		// Strange, but OK.
		{
			`(int float64, float64 int)`,
			&Signature{Parameters: []ParameterDecl{
				{Identifier: id("int"), Type: id("float64")},
				{Identifier: id("float64"), Type: id("int")},
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
		s := parseSignature(p)
		return &s
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
		{`map[int]a`, &MapType{Key: id("int"), Value: a}},
		{`map[*int]a.b`, &MapType{Key: &Star{Target: id("int")}, Value: sel(a, b).(Type)}},
		{`map[*int]map[string]int]`, &MapType{
			Key:   &Star{Target: id("int")},
			Value: &MapType{Key: id("string"), Value: id("int")},
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

func TestParseTypeIdentifier(t *testing.T) {
	parserTests{
		{`a`, a},
		{`a.b`, sel(a, b)},
		{`α.b`, sel(id("α"), b)},
	}.run(t, func(p *Parser) Node { return parseType(p) })
}

func TestParseFunctionLiteral(t *testing.T) {
	parserTests{
		{`func(){}`, &FunctionLiteral{}},
		{
			`func()big.Int{}`,
			&FunctionLiteral{
				Signature: Signature{
					Results: []ParameterDecl{{Type: bigInt}},
				},
			},
		},
		{
			`func(a){}`,
			&FunctionLiteral{
				Signature: Signature{
					Parameters: []ParameterDecl{{Type: a}},
				},
			},
		},
		{
			`func(a b){}`,
			&FunctionLiteral{
				Signature: Signature{
					Parameters: []ParameterDecl{{Identifier: a, Type: b}},
				},
			},
		},
		{
			`func(a b)big.Int{}`,
			&FunctionLiteral{
				Signature: Signature{
					Parameters: []ParameterDecl{{Identifier: a, Type: b}},
					Results:    []ParameterDecl{{Type: bigInt}},
				},
			},
		},
		{
			`func(){
				a
				b
				c
			}`,
			&FunctionLiteral{
				Body: BlockStmt{
					Statements: []Statement{aStmt, bStmt, cStmt},
				},
			},
		},
		{
			`func(a b)big.Int{
				a
				b
				return c
			}`,
			&FunctionLiteral{
				Signature: Signature{
					Parameters: []ParameterDecl{{Identifier: a, Type: b}},
					Results:    []ParameterDecl{{Type: bigInt}},
				},
				Body: BlockStmt{
					Statements: []Statement{
						aStmt,
						bStmt,
						&ReturnStmt{Expressions: []Expression{c}},
					},
				},
			},
		},
	}.run(t, func(p *Parser) Node { return parseExpr(p) })
}

func TestParseCompositeLiteral(t *testing.T) {
	parserTests{
		{`struct{ a int }{ a: 4 }`, &CompositeLiteral{
			Type: &StructType{Fields: []FieldDecl{
				{Identifier: a, Type: id("int")},
			}},
			Elements: []Element{
				{Key: a, Value: intLit("4")},
			},
		}},
		{`struct{ a, b int }{ a: 4, b: 5}`, &CompositeLiteral{
			Type: &StructType{Fields: []FieldDecl{
				{Identifier: a, Type: id("int")},
				{Identifier: b, Type: id("int")},
			}},
			Elements: []Element{
				{Key: a, Value: intLit("4")},
				{Key: b, Value: intLit("5")},
			},
		}},
		{`struct{ a []int }{ a: { 4, 5 } }`, &CompositeLiteral{
			Type: &StructType{Fields: []FieldDecl{
				{Identifier: a, Type: &SliceType{Type: id("int")}},
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
				{Identifier: a, Type: id("int")},
				{Identifier: b, Type: id("int")},
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
			Fields: []FieldDecl{{Identifier: x, Type: id("int")}},
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
		{`a.`, parseError{"expected.*\\( or Identifier"}},
		{`a[4`, parseError{"expected.*\\] or :"}},

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
			t.Errorf("parse(%s): expected error matching %s, got\n%s", test.text, pe.re, pp.MustString(n))
		} else if !regexp.MustCompile(pe.re).MatchString(err.Error()) {
			t.Errorf("parse(%s): expected an error matching %s, got %s", test.text, pe.re, err.Error())
		}
		return
	}
	if err != nil {
		t.Errorf("parse(%s): unexpected error: %s", test.text, err.Error())
	}
	if !eq.Deep(n, test.node) {
		t.Errorf("parse(%s): %s,\nexpected: %s", test.text, pp.MustString(n), pp.MustString(test.node))
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

type commentTests []struct {
	src string
	// Cmnts are the comments that should appear before each
	// subsequent identifier in the source.
	cmnts [][]string
}

func (tests commentTests) run(t *testing.T) {
	for i, test := range tests {
		p := NewParser(token.NewLexer("", test.src))
		n := 0
		for p.tok != token.EOF {
			if p.tok != token.Identifier {
				goto next
			}
			n++
			if n > len(test.cmnts) {
				break
			}
			// Test lengths, because nil != []string{}.
			if len(p.cmnts) == 0 && len(test.cmnts[n-1]) == 0 {
				goto next
			}
			if reflect.DeepEqual(p.cmnts, test.cmnts[n-1]) {
				goto next
			}
			t.Errorf("test %d: expected comments %v, got %v", i, test.cmnts[n-1], p.cmnts)
		next:
			p.next()
		}
		if n != len(test.cmnts) {
			t.Fatalf("test %d: expected %d identifiers, got %d", i, len(test.cmnts), n)
		}
	}
}

func TestComments(t *testing.T) {
	tests := commentTests{
		{`a`, [][]string{{}}},
		{"// a\na", [][]string{{"// a"}}},
		{"/* a */a", [][]string{{"/* a */"}}},
		{"/* a */\na", [][]string{{"/* a */"}}},
		{"// a\n// b\na", [][]string{{"// a", "// b"}}},
		{"// a\n/* b */\na", [][]string{{"// a", "/* b */"}}},
		{"// a\na\n// b\nb", [][]string{{"// a"}, {"// b"}}},
		{"a\n// b\nb", [][]string{{}, {"// b"}}},
		{"// a\na\nb", [][]string{{"// a"}, {}}},
		{"// a\na\n// b\n\nb", [][]string{{"// a"}, {}}},
		{"// a\na\n// b\n\n// c\nb", [][]string{{"// a"}, {"// c"}}},
	}
	tests.run(t)
}

func BenchmarkParser(b *testing.B) {
	for i := 0; i < b.N; i++ {
		p := NewParser(token.NewLexer("", test.Prog))
		_, err := Parse(p)
		if err != nil {
			b.Fatalf("parse error: %s", err)
		}
	}
}

func BenchmarkStandardLibraryParser(b *testing.B) {
	fset := stdtoken.NewFileSet()
	for i := 0; i < b.N; i++ {
		_, err := parser.ParseFile(fset, "", test.Prog, 0)
		if err != nil {
			b.Fatalf("parse error: %s", err)
		}
	}
}
