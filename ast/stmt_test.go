package ast

import (
	"testing"

	"github.com/velour/stop/token"
)

func TestStatements(t *testing.T) {
	tests := parserTests{
		{``, empty()},
		{`;`, empty()},
		{`fallthrough`, fallthroughStmt()},
		{`goto a`, gotoStmt(a)},
		{`break`, breakStmt(nil)},
		{`break a`, breakStmt(a)},
		{`continue`, continueStmt(nil)},
		{`continue a`, continueStmt(a)},
		{`return`, returnStmt()},
		{`return a, b, c`, returnStmt(a, b, c)},
		{`go a()`, goStmt(call(a, false))},
		{`defer a()`, deferStmt(call(a, false))},
	}
	tests.runStatements(t)
}

func TestFor(t *testing.T) {
	tests := parserTests{
		{`for { a }`, forLoop(nil, nil, nil, block(expr(a)))},
		{`for a { b }`, forLoop(nil, a, nil, block(expr(b)))},
		{
			`for a := 1; b; a++ { c }`,
			forLoop(shortDecl(ms(a), intLit("1")),
				b,
				incr(a),
				block(expr(c))),
		},
		{
			`for ; b; a++ { c }`,
			forLoop(nil,
				b,
				incr(a),
				block(expr(c))),
		},
		{
			`for a := 1; ; a++ { c }`,
			forLoop(shortDecl(ms(a), intLit("1")),
				nil,
				incr(a),
				block(expr(c))),
		},
		{
			`for a := 1; b;  { c }`,
			forLoop(shortDecl(ms(a), intLit("1")),
				b,
				nil,
				block(expr(c))),
		},
		{`for ; ;  { c }`, forLoop(nil, nil, nil, block(expr(c)))},
		{
			`for a, b, c *= 1, 2, 3; true; a++ { d }`,
			forLoop(assign(token.StarEqual, ms(a, b, c), intLit("1"), intLit("2"), intLit("3")),
				ident("true"),
				incr(a),
				block(expr(d))),
		},
		{
			`for a := range b { c }`,
			forRange(shortDecl(ms(a), b), block(expr(c))),
		},
		{
			`for a, b := range c { d }`,
			forRange(shortDecl(ms(a, b), c), block(expr(d))),
		},
		{
			`for a = range b { c }`,
			forRange(assign(token.Equal, ms(a), b), block(expr(c))),
		},
		{
			`for a, b = range c { d }`,
			forRange(assign(token.Equal, ms(a, b), c), block(expr(d))),
		},

		// Range is unexpected with any assign op other that =.
		{`for a *= range b { c }`, parseErr("range")},
		{`for a, b *= range c { d }`, parseErr("range")},
	}
	tests.runStatements(t)
}

func TestIf(t *testing.T) {
	tests := parserTests{
		{`if a { b }`, ifStmt(nil, a, block(expr(b)), nil)},
		{`if a; b { c }`, ifStmt(expr(a), b, block(expr(c)), nil)},
		{`if a; b { c } else { d }`, ifStmt(expr(a), b, block(expr(c)), block(expr(d)))},
		{
			`if a; b { c } else if d { 1 }`,
			ifStmt(expr(a), b, block(expr(c)),
				ifStmt(nil, d, block(expr(intLit("1"))), nil)),
		},
		{
			`if a { 1 } else if b { 2 } else if c { 3 } else { 4 }`,
			ifStmt(nil, a, block(expr(intLit("1"))),
				ifStmt(nil, b, block(expr(intLit("2"))),
					ifStmt(nil, c, block(expr(intLit("3"))),
						block(expr(intLit("4")))))),
		},
	}
	tests.runStatements(t)
}

func TestBlock(t *testing.T) {
	tests := parserTests{
		{`{}`, block()},
		{`{{{}}}`, block(block(block()))},
		{`{ a = b; c = d; }`, block(assign(token.Equal, ms(a), b), assign(token.Equal, ms(c), d))},
		{`{ a = b; c = d }`, block(assign(token.Equal, ms(a), b), assign(token.Equal, ms(c), d))},
		{`{
			a = b
			c = d
		}`, block(assign(token.Equal, ms(a), b), assign(token.Equal, ms(c), d))},
	}
	tests.runStatements(t)
}

func TestDeclarationStmt(t *testing.T) {
	tests := parserTests{
		{`const a = 5`, decl(cnst(ms(a), nil, intLit("5")))},
		{`const a int = 5`, decl(cnst(ms(a), typeName("", "int"), intLit("5")))},
		{`var(
			a, b int = 5, 6
			c = 7
			d = 8.0
		)`, decl(vars(ms(a, b), typeName("", "int"), intLit("5"), intLit("6")),
			vars(ms(c), nil, intLit("7")),
			vars(ms(d), nil, floatLit("8.0")))},
	}
	tests.runStatements(t)
}

func TestLabeledStmt(t *testing.T) {
	tests := parserTests{
		{`here: a = b`, labeled(ident("here"), assign(token.Equal, ms(a), b))},
		{`here: there: a := b`, labeled(ident("here"), labeled(ident("there"), shortDecl(ms(a), b)))},
	}
	tests.runStatements(t)
}

func TestShortVarDecl(t *testing.T) {
	tests := parserTests{
		{`a := 5`, shortDecl(ms(a), intLit("5"))},
		{`a, b := 5, 6`, shortDecl(ms(a, b), intLit("5"), intLit("6"))},

		// Only allow idents on LHS

		// The following is parsed as an expression statement with
		// a trailing := left for the next parse call.
		//{`a.b := 1`, parseErr("expected")},
		{`a, b.c := 1, 2`, parseErr("expected")},
	}
	tests.runStatements(t)
}

func TestAssignment(t *testing.T) {
	tests := parserTests{
		{`a = b`, assign(token.Equal, ms(a), b)},
		{`a += b`, assign(token.PlusEqual, ms(a), b)},
		{`a -= b`, assign(token.MinusEqual, ms(a), b)},
		{`a |= b`, assign(token.OrEqual, ms(a), b)},
		{`a ^= b`, assign(token.CarrotEqual, ms(a), b)},
		{`a *= b`, assign(token.StarEqual, ms(a), b)},
		{`a /= b`, assign(token.DivideEqual, ms(a), b)},
		{`a %= b`, assign(token.PercentEqual, ms(a), b)},
		{`a <<= b`, assign(token.LessLessEqual, ms(a), b)},
		{`a >>= b`, assign(token.GreaterGreaterEqual, ms(a), b)},
		{`a &= b`, assign(token.AndEqual, ms(a), b)},
		{`a &^= b`, assign(token.AndCarrotEqual, ms(a), b)},

		{`a, b = c, d`, assign(token.Equal, ms(a, b), c, d)},
		{`a.b, c, d *= 5, 6, 7`, assign(token.StarEqual,
			ms(sel(a, b), c, d), intLit("5"), intLit("6"), intLit("7"))},
	}
	tests.runStatements(t)
}

func TestExpressionStmt(t *testing.T) {
	tests := parserTests{
		{`a`, expr(a)},
		{`b`, expr(b)},
		{`a.b`, expr(sel(a, b))},
		{`a[5]`, expr(index(a, intLit("5")))},
	}
	tests.runStatements(t)
}

func TestIncDecStmt(t *testing.T) {
	tests := parserTests{
		{`a++`, incr(a)},
		{`b--`, decr(b)},
		{`a[5]++`, incr(index(a, intLit("5")))},
	}
	tests.runStatements(t)
}

func TestSendStmt(t *testing.T) {
	tests := parserTests{
		{`a <- b`, send(a, b)},
		{`a <- 5`, send(a, intLit("5"))},
		{`a[6] <- b[7]`, send(index(a, intLit("6")), index(b, intLit("7")))},
	}
	tests.runStatements(t)
}
