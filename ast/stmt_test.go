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

		// Range is disallowed outside of a for loop.
		{`a := range b`, parseErr("range")},
	}
	tests.runStatements(t)
}

func TestSelect(t *testing.T) {
	tests := parserTests{
		{`select{}`, selectStmt()},
		{
			`select{ case a <- b: c() }`,
			selectStmt(commMatcher{
				send:  send(a, b),
				stmts: ms(expr(call(c, false))),
			}),
		},
		{
			`select{ case a := <- b: c() }`,
			selectStmt(commMatcher{
				recv:  recv(ms(a), token.ColonEqual, b),
				stmts: ms(expr(call(c, false))),
			}),
		},
		{
			`select{ case a, b := <- c: d() }`,
			selectStmt(commMatcher{
				recv:  recv(ms(a, b), token.ColonEqual, c),
				stmts: ms(expr(call(d, false))),
			}),
		},
		{
			`select{ case a, b = <- c: d() }`,
			selectStmt(commMatcher{
				recv:  recv(ms(a, b), token.Equal, c),
				stmts: ms(expr(call(d, false))),
			}),
		},
		{
			`select{ case a() = <- b: c() }`,
			selectStmt(commMatcher{
				recv:  recv(ms(call(a, false)), token.Equal, b),
				stmts: ms(expr(call(c, false))),
			}),
		},
		{
			`select{ default: a() }`,
			selectStmt(commMatcher{
				stmts: ms(expr(call(a, false))),
			}),
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
			selectStmt(
				commMatcher{
					send:  send(a, b),
					stmts: ms(expr(call(c, false))),
				},
				commMatcher{
					recv:  recv(ms(call(a, false)), token.Equal, b),
					stmts: ms(expr(call(c, false)), expr(call(d, false))),
				},
				commMatcher{
					stmts: ms(expr(call(a, false))),
				},
			),
		},

		// := cannot appear after non-identifiers.
		{`select{ case a() := <- b: c() }`, parseErr(":=")},
		{`select{ case a, a() := <- b: c() }`, parseErr(":=")},

		// Only a receive expression can appear in a receive statement.
		{`select{ case a := b: c() }`, parseErr("LessMinus")},
		{`select{ case a = b: c() }`, parseErr("LessMinus")},
		{`select{ case a, b = *d: c() }`, parseErr("LessMinus")},
	}
	tests.runStatements(t)
}

func TestSwitch(t *testing.T) {
	tests := parserTests{
		{`switch {}`, exprSwitch(nil, nil)},
		{`switch a {}`, exprSwitch(nil, a)},
		{`switch ; a {}`, exprSwitch(nil, a)},
		{`switch a(b, c) {}`, exprSwitch(nil, call(a, false, b, c))},
		{`switch a(); b {}`, exprSwitch(expr(call(a, false)), b)},
		{`switch a = b; c {}`, exprSwitch(assign(token.Equal, ms(a), b), c)},
		{`switch a := b; c {}`, exprSwitch(shortDecl(ms(a), b), c)},
		{`switch a, b := b, a; c {}`, exprSwitch(shortDecl(ms(a, b), b, a), c)},
		{`switch a, b := b, a; a(b, c) {}`, exprSwitch(shortDecl(ms(a, b), b, a), call(a, false, b, c))},

		{
			`switch { case a: b() }`,
			exprSwitch(nil, nil, caseMatcher{
				guards: ms(a),
				stmts:  ms(expr(call(b, false))),
			}),
		},
		{
			`switch { case a, 5: b() }`,
			exprSwitch(nil, nil, caseMatcher{
				guards: ms(a, intLit("5")),
				stmts:  ms(expr(call(b, false))),
			}),
		},
		{
			`switch { default: a() }`,
			exprSwitch(nil, nil, caseMatcher{
				guards: ms(),
				stmts:  ms(expr(call(a, false))),
			}),
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
			exprSwitch(nil, nil,
				caseMatcher{
					guards: ms(a, intLit("5")),
					stmts:  ms(expr(call(b, false)), assign(token.Equal, ms(c), d)),
				},
				caseMatcher{
					guards: ms(ident("true"), ident("false")),
					stmts:  ms(expr(call(d, false)), fallthroughStmt()),
				},
				caseMatcher{
					guards: ms(),
					stmts:  ms(returnStmt(intLit("42"))),
				},
			),
		},

		{`switch a.(type) {}`, typeSwitch(nil, nil, a)},
		{`switch ; a.(type) {}`, typeSwitch(nil, nil, a)},
		{`switch a.b.(type) {}`, typeSwitch(nil, nil, sel(a, b))},
		{`switch a(b).(type) {}`, typeSwitch(nil, nil, call(a, false, b))},
		{`switch a.(b).(type) {}`, typeSwitch(nil, nil, tAssert(a, ident("b")))},
		{`switch a := b.(type) {}`, typeSwitch(nil, a, b)},
		{`switch a(); b := c.(type) {}`, typeSwitch(expr(call(a, false)), b, c)},

		{
			`switch a.(type) { case int: b() }`,
			typeSwitch(nil, nil, a, caseMatcher{
				guards: ms(ident("int")),
				stmts:  ms(expr(call(b, false))),
			}),
		},
		{
			`switch a.(type) { case int, float64: b() }`,
			typeSwitch(nil, nil, a, caseMatcher{
				guards: ms(ident("int"), ident("float64")),
				stmts:  ms(expr(call(b, false))),
			}),
		},
		{
			`switch a.(type) { default: b() }`,
			typeSwitch(nil, nil, a, caseMatcher{
				guards: ms(),
				stmts:  ms(expr(call(b, false))),
			}),
		},
		{
			`switch a.(type) { 
				case int, float64:
					b()
					c = d
				case interface{}:
					d()
					fallthrough
				default:
					return 42
			}`,
			typeSwitch(nil, nil, a,
				caseMatcher{
					guards: ms(ident("int"), ident("float64")),
					stmts:  ms(expr(call(b, false)), assign(token.Equal, ms(c), d)),
				},
				caseMatcher{
					guards: ms(ifaceType()),
					stmts:  ms(expr(call(d, false)), fallthroughStmt()),
				},
				caseMatcher{
					guards: ms(),
					stmts:  ms(returnStmt(intLit("42"))),
				},
			),
		},

		// Bad type switches.
		{`switch a.(type); b.(type) {}`, parseErr("")},
		{`switch a, b := c.(type) {}`, parseErr("")},
		{`switch a := b, c.(type) {}`, parseErr("")},
		{`switch a := b.(type), c {}`, parseErr("")},
		{`switch a = b.(type) {}`, parseErr("")},

		// Switches and composite literals.
		{`switch (struct{}{}) {}`, exprSwitch(nil, compLit(structType()))},
		{`switch (struct{}{}); 5 {}`, exprSwitch(expr(compLit(structType())), intLit("5"))},
		{`switch (struct{}{}).(type) {}`, typeSwitch(nil, nil, compLit(structType()))},
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

		// Only a single expression allow on LHS of a range.
		{`for a := range b, c { d }`, parseErr("")},
		{`for a, b := range c, d { 1 }`, parseErr("")},
		{`for a = range b, c { d }`, parseErr("")},
		{`for a, b = range c, d { 1 }`, parseErr("")},

		// Labels are not a simple statement.
		{`for label:; a < 100; a++`, parseErr(":")},

		// For loops and composite literals.
		{`for (struct{}{}) {}`, forLoop(nil, compLit(structType()), nil, block())},
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

		// If statements and composite literals.
		{`if (struct{}{}) {}`, ifStmt(nil, compLit(structType()), block(), nil)},
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
		{`const a int = 5`, decl(cnst(ms(a), ident("int"), intLit("5")))},
		{`var(
			a, b int = 5, 6
			c = 7
			d = 8.0
		)`, decl(vars(ms(a, b), ident("int"), intLit("5"), intLit("6")),
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
		{`a := b.(type)`, parseErr("type")},
		{`a, b := c.(type), 5`, parseErr("type")},
		{`a, b := 5, c.(type)`, parseErr("type")},
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
