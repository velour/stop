package lexer

import (
	"reflect"
	"strings"
	"testing"
)

type singleTokenTests []struct {
	text string
	want TokenType
}

func (tests singleTokenTests) run(t *testing.T) {
	for i, test := range tests {
		lex := New(strings.NewReader(test.text))
		got := lex.Next()
		if got.Type != test.want {
			t.Errorf("test %d: %s got %v, wanted %v", i, test.text, got, test.want)
		}
	}
}

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
			t.Errorf("test %d: %s got %v, wanted %v", i, test.text, got, test.want)
		}
	}
}

type locTests []struct {
	text string
	want [][2]int
}

func (tests locTests) run(t *testing.T, loc func(*Token) [2]int) {
	for i, test := range tests {
		lex := New(strings.NewReader(test.text))
		got := make([][2]int, 0, len(test.want))
		for {
			tok := lex.Next()
			if tok.Type == EOF {
				break
			}
			got = append(got, loc(tok))
		}
		if !reflect.DeepEqual(got, test.want) {
			t.Errorf("test %d: %s got %v, wanted %v", i, test.text, got, test.want)
		}
	}
}

func TestEOF(t *testing.T) {
	tests := multiTokenTests{
		{"", []TokenType{EOF}},
		{"\n", []TokenType{Whitespace, EOF}},
	}
	tests.run(t)
}

func TestIsNewline(t *testing.T) {
	tests := []struct {
		text      string
		isNewline bool
	}{
		{"\n", true},
		{"//\n", true},
		{"//foo\n", true},
		{"//foo", true},
		{"/*f\noo*/", true},
		{"", false},
		{"/*foo*/", false},
		{" ", false},
		{"\r", false},
	}
	for i, test := range tests {
		lex := New(strings.NewReader(test.text))
		got := lex.Next()
		if got.IsNewline() != test.isNewline {
			t.Errorf("test %d: %s got %v, wanted %v", i, test.text, got.IsNewline(), test.isNewline)
		}
	}
}

func TestLineComment(t *testing.T) {
	tests := multiTokenTests{
		{"//\n", []TokenType{Comment, EOF}},
		{"// hello\n", []TokenType{Comment, EOF}},
		{"//", []TokenType{Comment, EOF}},
		{"\n//", []TokenType{Whitespace, Comment, EOF}},
	}
	tests.run(t)
}

func TestMultiLineComment(t *testing.T) {
	tests := multiTokenTests{
		{"/* comment */", []TokenType{Comment, EOF}},
		{"/*\ncomment */", []TokenType{Comment, EOF}},
		{"/*\ncomment\n*/", []TokenType{Comment, EOF}},
		{"/* // comment */", []TokenType{Comment, EOF}},
		{"/* // /*comment*/ */", []TokenType{Comment, Whitespace, Star, Divide, EOF}},
	}
	tests.run(t)
}

func TestSemicolonInsertion(t *testing.T) {
	tests := multiTokenTests{
		{"identifier\n", []TokenType{Identifier, Semicolon, Whitespace, EOF}},
		{"++\n", []TokenType{PlusPlus, Semicolon, Whitespace, EOF}},
		{"++//hi", []TokenType{PlusPlus, Semicolon, Comment, EOF}},
		{"++/**/", []TokenType{PlusPlus, Comment, EOF}},
		{"5\n", []TokenType{IntegerLiteral, Semicolon, Whitespace, EOF}},
		{"'a'\n", []TokenType{RuneLiteral, Semicolon, Whitespace, EOF}},
		{"\"hello\"\n", []TokenType{StringLiteral, Semicolon, Whitespace, EOF}},
	}
	tests.run(t)
}

func TestKeywords(t *testing.T) {
	tests := singleTokenTests{
		{"break", Break},
		{"default", Default},
		{"func", Func},
		{"interface", Interface},
		{"select", Select},
		{"case", Case},
		{"defer", Defer},
		{"go", Go},
		{"map", Map},
		{"struct", Struct},
		{"chan", Chan},
		{"else", Else},
		{"goto", Goto},
		{"package", Package},
		{"switch", Switch},
		{"const", Const},
		{"fallthrough", Fallthrough},
		{"if", If},
		{"range", Range},
		{"type", Type},
		{"continue", Continue},
		{"for", For},
		{"import", Import},
		{"return", Return},
		{"var", Var},
	}
	tests.run(t)
}

func TestOperators(t *testing.T) {
	tests := singleTokenTests{
		{"+", Plus},
		{"&", And},
		{"+=", PlusEqual},
		{"&=", AndEqual},
		{"&&", AndAnd},
		{"==", EqualEqual},
		{"!=", BangEqual},
		{"(", OpenParen},
		{")", CloseParen},
		{"-", Minus},
		{"|", Or},
		{"-=", MinusEqual},
		{"|=", OrEqual},
		{"||", OrOr},
		{"<", Less},
		{"<=", LessEqual},
		{"[", OpenBracket},
		{"]", CloseBracket},
		{"*", Star},
		{"^", Carrot},
		{"*=", StarEqual},
		{"^=", CarrotEqual},
		{"<-", LessMinus},
		{">", Greater},
		{">=", GreaterEqual},
		{"{", OpenBrace},
		{"}", CloseBrace},
		{"/", Divide},
		{"<<", LessLess},
		{"/=", DivideEqual},
		{"<<=", LessLessEqual},
		{"++", PlusPlus},
		{"=", Equal},
		{":=", ColonEqual},
		{",", Comma},
		{";", Semicolon},
		{"%", Percent},
		{">>", GreaterGreater},
		{"%=", PercentEqual},
		{">>=", GreaterGreaterEqual},
		{"--", MinusMinus},
		{"!", Bang},
		{"...", DotDotDot},
		{".", Dot},
		{":", Colon},
		{"&^", AndCarrot},
		{"&^=", AndCarrotEqual},
	}
	tests.run(t)
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
		{"'\\''", RuneLiteral},
		{`'\"'`, RuneLiteral},
		{"'\\b'", RuneLiteral},
		{"'\\r'", RuneLiteral},
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

func TestLineNumbers(t *testing.T) {
	tests := locTests{
		{"\n", [][2]int{{1, 2}}},
		{"\n\n", [][2]int{{1, 3}}},
		{"// foo \n", [][2]int{{1, 2}}},
		{"/* foo */", [][2]int{{1, 1}}},
		{"/* foo \n*/", [][2]int{{1, 2}}},
		{"/* foo \n*/\n", [][2]int{{1, 2}, {2, 3}}},
		{"\nident", [][2]int{{1, 2}, {2, 2}}},
		{"\nα", [][2]int{{1, 2}, {2, 2}}},
		{"\nident\n&&", [][2]int{{1, 2}, {2, 2}, {2, 2}, {2, 3}, {3, 3}}},
	}
	tests.run(t, func(tok *Token) [2]int {
		return [2]int{tok.Span[0].Line, tok.Span[1].Line}
	})
}

func TestRuneNumbers(t *testing.T) {
	tests := locTests{
		{"\n", [][2]int{{1, 2}}},
		{"hello", [][2]int{{1, 6}}},
		{"\t\t\t\t\t", [][2]int{{1, 6}}},
		{"α", [][2]int{{1, 2}}},
		{"αβ", [][2]int{{1, 3}}},
		{"α β", [][2]int{{1, 2}, {2, 3}, {3, 4}}},
		{"α\nβ", [][2]int{{1, 2}, {2, 2}, {2, 3}, {3, 4}}},
	}
	tests.run(t, func(tok *Token) [2]int {
		return [2]int{tok.Span[0].Rune, tok.Span[1].Rune}
	})
}
