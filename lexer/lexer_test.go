package lexer

import (
	"reflect"
	"strings"
	"testing"
)

type multiTokenTests []struct {
	text string
	want []TokenType
}

func (tests multiTokenTests) run(t *testing.T) {
	for i, test := range tests {
		lex := New(strings.NewReader(test.text))
		got := make([]TokenType, 0, len(test.want))
		for len(got) == 0 || got[len(got)-1] != EOF {
			got = append(got, lex.Next().Type)
		}
		if !reflect.DeepEqual(got, test.want) {
			t.Errorf("test %d: %s got %s, wanted %s", i, test.text, got, test.want)
		}
	}
}

func TestEOF(t *testing.T) {
	tests := multiTokenTests{
		{"", []TokenType{Newline, EOF}},
		{"\n", []TokenType{Newline, EOF}},
	}
	tests.run(t)
}

func TestLineComment(t *testing.T) {
	tests := multiTokenTests{
		{"//\n", []TokenType{Newline, EOF}},
		{"// hello\n", []TokenType{Newline, EOF}},
		{"//", []TokenType{Newline, EOF}},
		{"\n//", []TokenType{Newline, Newline, EOF}},
	}
	tests.run(t)
}

func TestMultiLineComment(t *testing.T) {
	tests := multiTokenTests{
		{"/* comment */", []TokenType{Whitespace, Newline, EOF}},
		{"/*\ncomment */", []TokenType{Newline, EOF}},
		{"/*\ncomment\n*/", []TokenType{Newline, EOF}},
		{"/* // comment */", []TokenType{Whitespace, Newline, EOF}},
		{"/* // /*comment*/ */", []TokenType{Whitespace, Whitespace, Star, Divide, Newline, EOF}},
	}
	tests.run(t)
}

func TestSemicolonInsertion(t *testing.T) {
	tests := multiTokenTests{
		{"identifier\n", []TokenType{Identifier, Semicolon, Newline, EOF}},
		{"identifier", []TokenType{Identifier, Semicolon, Newline, EOF}},
		{"++", []TokenType{PlusPlus, Semicolon, Newline, EOF}},
		{"++//hi", []TokenType{PlusPlus, Semicolon, Newline, EOF}},
	}
	tests.run(t)
}

func TestKeywords(t *testing.T) {
	tests := multiTokenTests{
		{"break", []TokenType{Break, Semicolon, Newline, EOF}},
		{"default", []TokenType{Default, Newline, EOF}},
		{"func", []TokenType{Func, Newline, EOF}},
		{"interface", []TokenType{Interface, Newline, EOF}},
		{"select", []TokenType{Select, Newline, EOF}},
		{"case", []TokenType{Case, Newline, EOF}},
		{"defer", []TokenType{Defer, Newline, EOF}},
		{"go", []TokenType{Go, Newline, EOF}},
		{"map", []TokenType{Map, Newline, EOF}},
		{"struct", []TokenType{Struct, Newline, EOF}},
		{"chan", []TokenType{Chan, Newline, EOF}},
		{"else", []TokenType{Else, Newline, EOF}},
		{"goto", []TokenType{Goto, Newline, EOF}},
		{"package", []TokenType{Package, Newline, EOF}},
		{"switch", []TokenType{Switch, Newline, EOF}},
		{"const", []TokenType{Const, Newline, EOF}},
		{"fallthrough", []TokenType{Fallthrough, Semicolon, Newline, EOF}},
		{"if", []TokenType{If, Newline, EOF}},
		{"range", []TokenType{Range, Newline, EOF}},
		{"type", []TokenType{Type, Newline, EOF}},
		{"continue", []TokenType{Continue, Semicolon, Newline, EOF}},
		{"for", []TokenType{For, Newline, EOF}},
		{"import", []TokenType{Import, Newline, EOF}},
		{"return", []TokenType{Return, Semicolon, Newline, EOF}},
		{"var", []TokenType{Var, Newline, EOF}},
	}
	tests.run(t)
}

func TestOperators(t *testing.T) {
	tests := multiTokenTests{
		{"+", []TokenType{Plus, Newline, EOF}},
		{"&", []TokenType{And, Newline, EOF}},
		{"+=", []TokenType{PlusEqual, Newline, EOF}},
		{"&=", []TokenType{AndEqual, Newline, EOF}},
		{"&&", []TokenType{AndAnd, Newline, EOF}},
		{"==", []TokenType{EqualEqual, Newline, EOF}},
		{"!=", []TokenType{BangEqual, Newline, EOF}},
		{"(", []TokenType{OpenParen, Newline, EOF}},
		{")", []TokenType{CloseParen, Semicolon, Newline, EOF}},
		{"-", []TokenType{Minus, Newline, EOF}},
		{"|", []TokenType{Or, Newline, EOF}},
		{"-=", []TokenType{MinusEqual, Newline, EOF}},
		{"|=", []TokenType{OrEqual, Newline, EOF}},
		{"||", []TokenType{OrOr, Newline, EOF}},
		{"<", []TokenType{Less, Newline, EOF}},
		{"<=", []TokenType{LessEqual, Newline, EOF}},
		{"[", []TokenType{OpenBracket, Newline, EOF}},
		{"]", []TokenType{CloseBracket, Semicolon, Newline, EOF}},
		{"*", []TokenType{Star, Newline, EOF}},
		{"^", []TokenType{Carrot, Newline, EOF}},
		{"*=", []TokenType{StarEqual, Newline, EOF}},
		{"^=", []TokenType{CarrotEqual, Newline, EOF}},
		{"<-", []TokenType{LessMinus, Newline, EOF}},
		{">", []TokenType{Greater, Newline, EOF}},
		{">=", []TokenType{GreaterEqual, Newline, EOF}},
		{"{", []TokenType{OpenBrace, Newline, EOF}},
		{"}", []TokenType{CloseBrace, Semicolon, Newline, EOF}},
		{"/", []TokenType{Divide, Newline, EOF}},
		{"<<", []TokenType{LessLess, Newline, EOF}},
		{"/=", []TokenType{DivideEqual, Newline, EOF}},
		{"<<=", []TokenType{LessLessEqual, Newline, EOF}},
		{"++", []TokenType{PlusPlus, Semicolon, Newline, EOF}},
		{"=", []TokenType{Equal, Newline, EOF}},
		{":=", []TokenType{ColonEqual, Newline, EOF}},
		{",", []TokenType{Comma, Newline, EOF}},
		{";", []TokenType{Semicolon, Newline, EOF}},
		{"%", []TokenType{Percent, Newline, EOF}},
		{">>", []TokenType{GreaterGreater, Newline, EOF}},
		{"%=", []TokenType{PercentEqual, Newline, EOF}},
		{">>=", []TokenType{GreaterGreaterEqual, Newline, EOF}},
		{"--", []TokenType{MinusMinus, Semicolon, Newline, EOF}},
		{"!", []TokenType{Bang, Newline, EOF}},
		{"...", []TokenType{DotDotDot, Newline, EOF}},
		{".", []TokenType{Dot, Newline, EOF}},
		{":", []TokenType{Colon, Newline, EOF}},
		{"&^", []TokenType{AndCarrot, Newline, EOF}},
		{"&^=", []TokenType{AndCarrotEqual, Newline, EOF}},
	}
	tests.run(t)
}

type singleTokenTests []struct {
	text string
	want TokenType
}

func (tests singleTokenTests) run(t *testing.T) {
	for i, test := range tests {
		lex := New(strings.NewReader(test.text))
		got := lex.Next()
		if got.Type != test.want {
			t.Errorf("test %d: %s got %s, wanted %s", i, test.text, got, test.want)
		}
	}
}

func TestIdentifier(t *testing.T) {
	tests := singleTokenTests{
		{"a", Identifier},
		{"_x9", Identifier},
		{"ThisVariableIsExported", Identifier},
		{"αβ", Identifier},
	}
	tests.run(t)
}

func TestIntegerLiteral(t *testing.T) {
	tests := singleTokenTests{
		{"42", IntegerLiteral},
		{"0600", IntegerLiteral},
		{"0xBadFace", IntegerLiteral},
		{"170141183460469231731687303715884105727", IntegerLiteral},
		{"0", IntegerLiteral},
		{"01", IntegerLiteral},
		{"0777", IntegerLiteral},
		{"0xF", IntegerLiteral},
		{"0XF", IntegerLiteral},
		{"9832", IntegerLiteral},
	}
	tests.run(t)
}

func TestFloatLiteral(t *testing.T) {
	tests := singleTokenTests{
		{"0.", FloatLiteral},
		{"72.40", FloatLiteral},
		{"072.40", FloatLiteral},
		{"2.71828", FloatLiteral},
		{"1.e+0", FloatLiteral},
		{"6.67428e-11", FloatLiteral},
		{"1E6", FloatLiteral},
		{".25", FloatLiteral},
		{".12345E+5", FloatLiteral},
	}
	tests.run(t)
}

func TestImaginaryLiteral(t *testing.T) {
	tests := singleTokenTests{
		{"0i", ImaginaryLiteral},
		{"011i", ImaginaryLiteral},
		{"0.i", ImaginaryLiteral},
		{"2.71828i", ImaginaryLiteral},
		{"1.e+0i", ImaginaryLiteral},
		{"6.67428e-11i", ImaginaryLiteral},
		{"1E6i", ImaginaryLiteral},
		{".25i", ImaginaryLiteral},
		{".12345E+5i", ImaginaryLiteral},
	}
	tests.run(t)
}

func TestRuneLiteral(t *testing.T) {
	tests := singleTokenTests{
		{"'a'", RuneLiteral},
		{"'ä'", RuneLiteral},
		{"'本'", RuneLiteral},
		{"'\\t'", RuneLiteral},
		{"'\\000'", RuneLiteral},
		{"'\\007'", RuneLiteral},
		{"'\\377'", RuneLiteral},
		{"'\\x07'", RuneLiteral},
		{"'\\xff'", RuneLiteral},
		{"'\\u12e4'", RuneLiteral},
		{"'\\U00101234'", RuneLiteral},

		{"'aa'", Error},
		{"'\\xa'", Error},
		{"'\\0'", Error},

		// The following two should be errors, but we don't validate the Unicode code points.
		{"'\\uDFFF'", RuneLiteral},
		{"'\\U00110000'", RuneLiteral},
	}
	tests.run(t)
}

func TestInterpretedStringLiteral(t *testing.T) {
	tests := singleTokenTests{
		{`"\n"`, StringLiteral},
		{`""`, StringLiteral},
		{`"Hello, world!\n"`, StringLiteral},
		{`"日本語"`, StringLiteral},
		{`"\u65e5本\U00008a9e"`, StringLiteral},
		{`"\xff\u00FF"`, StringLiteral},
		{`"日本語"`, StringLiteral},
		{`"\u65e5\u672c\u8a9e"`, StringLiteral},
		{`"\U000065e5\U0000672c\U00008a9e"`, StringLiteral},
		{`"\xe6\x97\xa5\xe6\x9c\xac\xe8\xaa\x9e"`, StringLiteral},

		{`"not terminated`, Error},

		// The following two should be errors, but we don`t validate the Unicode code points.
		{`"\uD800"`, StringLiteral},
		{`"\U00110000"`, StringLiteral},
	}
	tests.run(t)
}

func TestRawStringLiteral(t *testing.T) {
	tests := singleTokenTests{
		// `abc`
		{"\x60abc\x60", StringLiteral},
		// '\n
		// \n'
		{"\x60\\n\x0a\\n\x60", StringLiteral},
		{"\x60日本語\x60", StringLiteral},

		{"\x60not terminated", Error},
	}
	tests.run(t)
}
