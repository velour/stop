package ast

import (
	"testing"
)

func TestParseOperandName(t *testing.T) {
	tests := parserTests{
		{"_abc123", opName("", "_abc123")},
		{"os.Stderr", opName("os", "Stderr")},
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
