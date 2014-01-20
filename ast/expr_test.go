package ast

import (
	"testing"

	"bitbucket.org/eaburns/stop/token"
)

func TestParseOperandName(t *testing.T) {
	tests := parserTests{
		{"_abc123", opName("", "_abc123")},
		{"os.Stderr", opName("os", "Stderr")},
	}
	tests.run(t)
}

func TestParseBinaryExpr(t *testing.T) {
	a, b, c, d := opName("", "a"), opName("", "b"), opName("", "c"), opName("", "d")
	tests := parserTests{
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
	}
	tests.run(t)

}

func TestParseUnaryExpr(t *testing.T) {
	tests := parserTests{
		{`+a`, unOp(token.Plus, opName("", "a"))},
		{`-math.b`, unOp(token.Minus, opName("math", "b"))},
		{`!0`, unOp(token.Bang, intLit("0"))},
		{`^(a)`, unOp(token.Carrot, opName("", "a"))},
		{`*5.1`, unOp(token.Star, floatLit("5.1"))},
		{`&!1`, unOp(token.And, unOp(token.Bang, intLit("1")))},
		{`<-z`, unOp(token.LessMinus, opName("", "z"))},
		{`<-!-z`, unOp(token.LessMinus, unOp(token.Bang, unOp(token.Minus, opName("", "z"))))},
	}
	tests.run(t)
}

func TestParseIntegerLiteral(t *testing.T) {
	tests := parserTests{
		{"1", intLit("1")},
		{"010", intLit("8")},
		{"0x10", intLit("16")},

		{"08", parseErr("malformed.*integer")},
	}
	tests.run(t)
}

func TestParseFloatLiteral(t *testing.T) {
	tests := parserTests{
		{"1.", floatLit("1.0")},
		{"1.0", floatLit("1.0")},
		{"0.1", floatLit("0.1")},
		{"0.1000", floatLit("0.1")},
		{"1e1", floatLit("10.0")},
		{"1e-1", floatLit("0.1")},
	}
	tests.run(t)
}

func TestParseImaginaryLiteral(t *testing.T) {
	tests := parserTests{
		{"0.i", imaginaryLit("0.0i")},
		{"1.i", imaginaryLit("1.0i")},
		{"1.0i", imaginaryLit("1.0i")},
		{"0.1i", imaginaryLit("0.1i")},
		{"0.1000i", imaginaryLit("0.1i")},
		{"1e1i", imaginaryLit("10.0i")},
		{"1e-1i", imaginaryLit("0.1i")},
	}
	tests.run(t)
}

func TestParseStringLiteral(t *testing.T) {
	tests := parserTests{
		{`""`, strLit("")},
		{`"abc"`, strLit("abc")},
		{`"αβξ"`, strLit("αβξ")},
		{`"\x0A""`, strLit("\n")},
		{`"\n\r\t\v\""`, strLit("\n\r\t\v\"")},

		// Cannot have a newline in an interpreted string.
		{"\x22\x0A\x22", parseErr("unexpected")},
		// Cannot escape a single quote in an interpreted string lit.
		{`"\'""`, parseErr("unexpected.*'")},

		{"\x60\x60", strLit("")},
		{"\x60\x5C\x60", strLit("\\")},
		{"\x60\x0A\x60", strLit("\n")},
		{"\x60abc\x60", strLit("abc")},
		// Strip \r from raw string literals.
		{"\x60\x0D\x60", strLit("")},
		{"\x60αβξ\x60", strLit("αβξ")},
	}
	tests.run(t)
}

func TestParseRuneLiteral(t *testing.T) {
	tests := parserTests{
		{`'a'`, runeLit('a')},
		{`'α'`, runeLit('α')},
		{`'\''`, runeLit('\'')},
		{`'\000'`, runeLit('\000')},
		{`'\x00'`, runeLit('\x00')},
		{`'\xFF'`, runeLit('\xFF')},
		{`'\u00FF'`, runeLit('\u00FF')},
		{`'\U000000FF'`, runeLit('\U000000FF')},
		{`'\U0010FFFF'`, runeLit('\U0010FFFF')},

		{`'\"'`, parseErr("unexpected.*\"")},
		{`'\008'`, parseErr("unexpected")},
		{`'\U00110000'`, parseErr("malformed")},
	}
	tests.run(t)
}
