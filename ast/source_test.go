package ast

import (
	"math/big"
	"testing"

	"github.com/eaburns/pretty"
	"github.com/velour/stop/token"
)

// BUG(eaburns): Test NilLiteral, BoolLiteral, and ComplexLiteral.
func TestExpressionSource(t *testing.T) {
	tests := []struct{ expr, want string }{
		{`"Hello, World!"`, `"Hello, World!"`},
		{`"αβξ"`, `"αβξ"`},
		{"\x60\\n\x0a\\n\x60", `"\\n\n\\n"`},
		{`'a'`, `'a'`},
		{`'☺'`, `'☺'`},
		{`0`, `0`},
		{`5`, `5`},
		{`0.0`, `0.0`},
		{`5.0`, `5.0`},
		{`3.1415926535`, `3.1415926535`},
		{`1e4`, `10000.0`},
		{`1e-4`, `0.0001`},
		// BUG(eaburns): ComplexLiteral tests don't test ComplexLiterals with a
		// real component, since they are never returned by the parser.
		{`0i`, `0.0i`},
		{`5i`, `5.0i`},
		{`3.1415926535i`, `3.1415926535i`},
		{`a`, `a`},
		{`abc_xyz`, `abc_xyz`},
		{`-1`, `-1`},
		{`+-1`, `+-1`},
		{`<-abc_xyz`, `<-abc_xyz`},
		{`**a`, `**a`},
		{`1 + 1`, `1 + 1`},
		{`1 * abc`, `1 * abc`},
		{`-x / *b`, `-x / *b`},
		{`a + b / c + d`, `a + b / c + d`},
		{`(a + b) / (c + d)`, `(a + b) / (c + d)`},
		{`a()`, `a()`},
		{`a(b, c)`, `a(b, c)`},
		{`a(b, c...)`, `a(b, c...)`},
		{`*a(b, (c + d) / e)`, `*a(b, (c + d) / e)`},
		{`(*int32)(4)`, `(*int32)(4)`},
		{`a.b.c`, `a.b.c`},
		{`a(b, c).d`, `a(b, c).d`},
		{`-a(b, c).d + 12`, `-a(b, c).d + 12`},
		{`a.(b)`, `a.(b)`},
		{`a().(b.c).(e)`, `a().(b.c).(e)`},
		{`a[:]`, `a[:]`},
		{`a[1:]`, `a[1:]`},
		{`a[:2]`, `a[:2]`},
		{`a[1:2]`, `a[1:2]`},
		{`a[1:2:3]`, `a[1:2:3]`},
		{`a[:len(a) - 1]`, `a[:len(a) - 1]`},
		{`a[0]`, `a[0]`},
		{`a[len(a) - 1]`, `a[len(a) - 1]`},
		{`point{a, b}`, `point{a, b}`},
		{`point{x: a, y: b}`, `point{x: a, y: b}`},
		{`(*point){x: a, y: b}`, `(*point){x: a, y: b}`},
		{`[]int`, `[]int`},
		{`[][]int32{}`, `[][]int32{}`},
		{`[2]float64`, `[2]float64`},
		{`[...]float64{0, 0}`, `[...]float64{0, 0}`},
		{`map[string]int`, `map[string]int`},
		{`map[string]*[]float64`, `map[string]*[]float64`},
		{`chan int`, `chan int`},
		{`<-chan int`, `<-chan int`},
		{`chan<- int`, `chan<- int`},
		{`chan<- <-chan int`, `chan<- <-chan int`},
		{`func(){}`, `func(){…}`},
		{`func(int){}`, `func(int){…}`},
		{`func(int)bool{}`, `func(int)bool{…}`},
		{`func(int, float64){}`, `func(int, float64){…}`},
		{`func()(int, int){}`, `func()(int, int){…}`},
		{`func(int, ...float64)(int, int){}`, `func(int, ...float64)(int, int){…}`},
		{`func(x, y int)(z int){}`, `func(x int, y int)(z int){…}`},
		{`func(x, y ...int)(z int, err error){}`, `func(x int, y ...int)(z int, err error){…}`},
		{`struct{ x, y int }{}`, `struct{…}{}`},
		{`interface{ x(); y(int)bool }{}`, `interface{…}{}`},
	}
	for _, test := range tests {
		l := token.NewLexer("", test.expr)
		p := NewParser(l)
		e := parseExpr(p)
		got := e.Source()
		if test.want != got {
			t.Errorf("(%s).Source()=[%s], want [%s]", test.expr, got, test.want)
		}
	}

	// These can't be tested above, because the parser never returns them.
	var zero big.Rat
	nodeTests := []struct {
		expr Expression
		want string
	}{
		{&NilLiteral{}, "nil"},
		{&BoolLiteral{}, "false"},
		{&BoolLiteral{Value: true}, "true"},
		{&ComplexLiteral{Real: &zero, Imaginary: &zero}, "0.0i"},
		{&ComplexLiteral{Real: &zero, Imaginary: big.NewRat(5, 1)}, "5.0i"},
		{&ComplexLiteral{Real: big.NewRat(5, 1), Imaginary: &zero}, "5.0+0.0i"},
		{&ComplexLiteral{Real: big.NewRat(5, 1), Imaginary: big.NewRat(5, 1)}, "5.0+5.0i"},
		{&ComplexLiteral{Real: big.NewRat(5, 1), Imaginary: big.NewRat(-5, 1)}, "5.0-5.0i"},
		{&ComplexLiteral{Real: big.NewRat(-5, 1), Imaginary: big.NewRat(-5, 1)}, "-5.0-5.0i"},
	}
	for _, test := range nodeTests {
		got := test.expr.Source()
		if got != test.want {
			t.Errorf("(%s).Source()=[%s], want [%s]", pretty.String(test.expr), got, test.want)
		}
	}
}
