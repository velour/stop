package ast

import (
	"testing"

	"github.com/velour/stop/token"
)

//func TestDeclStatement(t *testing.T) {
//	tests := parserTests{
//		{`const a = 5`, decl(cnst(ms(a), intLit("5")))},
//	}
//	tests.runStatements(t)
//}
//
//func TestLabeledStatement(t *testing.T) {
//	tests := parserTests{
//		{`here: a = b`, labeled(ident("here"),
//			assign(token.Equal, ms(a), b))},
//	}
//	tests.runStatements(t)
//}

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

func TestExpressionStatement(t *testing.T) {
	tests := parserTests{
		{`a`, expr(a)},
		{`b`, expr(b)},
		{`a.b`, expr(sel(a, b))},
		{`a[5]`, expr(index(a, intLit("5")))},
	}
	tests.runStatements(t)
}

func TestIncDecStatement(t *testing.T) {
	tests := parserTests{
		{`a++`, incr(a)},
		{`b--`, decr(b)},
		{`a[5]++`, incr(index(a, intLit("5")))},
	}
	tests.runStatements(t)
}

func TestSendStatement(t *testing.T) {
	tests := parserTests{
		{`a <- b`, send(a, b)},
		{`a <- 5`, send(a, intLit("5"))},
		{`a[6] <- b[7]`, send(index(a, intLit("6")), index(b, intLit("7")))},
	}
	tests.runStatements(t)
}
