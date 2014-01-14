package token

import (
	"reflect"
	"testing"
)

type singleTokenTests []struct {
	text string
	want Token
}

func (tests singleTokenTests) run(t *testing.T) {
	for i, test := range tests {
		lex := NewLexer("", test.text)
		got := lex.Next()
		if got != test.want {
			t.Errorf("test %d: %s got %s, wanted %v", i, test.text, got, test.want)
		}
	}
}

type multiTokenTests []struct {
	text string
	want []Token
}

func (tests multiTokenTests) run(t *testing.T) {
	for i, test := range tests {
		lex := NewLexer("", test.text)
		got := make([]Token, 0, len(test.want))
		for len(got) == 0 || got[len(got)-1] != EOF {
			got = append(got, lex.Next())
		}
		if !reflect.DeepEqual(got, test.want) {
			t.Errorf("test %d: %s got %s, wanted %v", i, test.text, got, test.want)
		}
	}
}

type locTests []struct {
	text string
	want [][2]int
}

func (tests locTests) run(t *testing.T, loc func(s, e Location) [2]int) {
	for i, test := range tests {
		lex := NewLexer("", test.text)
		got := make([][2]int, 0, len(test.want))
		for {
			tok := lex.Next()
			if tok == EOF {
				break
			}
			got = append(got, loc(lex.Start, lex.End))
		}
		if !reflect.DeepEqual(got, test.want) {
			t.Errorf("test %d: %s got %v, wanted %v", i, test.text, got, test.want)
		}
	}
}

func TestReplace(t *testing.T) {
	l := NewLexer("", "αβ")
	if l.rune() != 'α' {
		t.Fatalf("first rune was not α")
	}
	l.replace()
	if l.rune() != 'α' {
		t.Errorf("expected α")
	}
	if l.rune() != 'β' {
		t.Errorf("expected β")
	}
}

func TestError(t *testing.T) {
	tests := []struct {
		text       string
		unexpected rune
	}{
		{`'\z'`, 'z'},
		{`/\`, '\\'},
		{`*\`, '\\'},
		{`"\'"`, '\''},
		{`'\"'`, '"'},
	}
	for i, test := range tests {
		lex := NewLexer("", test.text)
		got := Semicolon
		for got != EOF && got != Error {
			got = lex.Next()
		}
		if got == EOF {
			t.Fatalf("no error token")
		}
		if lex.Text() != string([]rune{test.unexpected}) {
			t.Errorf("test %d: %s got %s, wanted %c", i, test.text, lex.Text(), test.unexpected)
		}
	}
}

func TestEOF(t *testing.T) {
	tests := multiTokenTests{
		{"", []Token{EOF}},
		{"\n", []Token{Whitespace, EOF}},
	}
	tests.run(t)
}

func TestLineComment(t *testing.T) {
	tests := multiTokenTests{
		{"//\n", []Token{Comment, EOF}},
		{"// hello\n", []Token{Comment, EOF}},
		{"//", []Token{Comment, EOF}},
		{"\n//", []Token{Whitespace, Comment, EOF}},
	}
	tests.run(t)
}

func TestMultiLineComment(t *testing.T) {
	tests := multiTokenTests{
		{"/* comment */", []Token{Comment, EOF}},
		{"/*\ncomment */", []Token{Comment, EOF}},
		{"/*\ncomment\n*/", []Token{Comment, EOF}},
		{"/* // comment */", []Token{Comment, EOF}},
		{"/* // /*comment*/ */", []Token{Comment, Whitespace, Star, Divide, EOF}},
	}
	tests.run(t)
}

func TestSemicolonInsertion(t *testing.T) {
	tests := multiTokenTests{
		{"identifier\n", []Token{Identifier, Semicolon, Whitespace, EOF}},
		{"++\n", []Token{PlusPlus, Semicolon, Whitespace, EOF}},
		{"++//hi", []Token{PlusPlus, Comment, Semicolon, EOF}},
		{"++/**/", []Token{PlusPlus, Comment, Semicolon, EOF}},
		{"++ ", []Token{PlusPlus, Whitespace, Semicolon, EOF}},
		{"5\n", []Token{IntegerLiteral, Semicolon, Whitespace, EOF}},
		{"'a'\n", []Token{RuneLiteral, Semicolon, Whitespace, EOF}},
		{"\"hello\"\n", []Token{StringLiteral, Semicolon, Whitespace, EOF}},
		// No semicolon inserted on blank lines.
		{"\n\n", []Token{Whitespace, EOF}},
		{"\n \n", []Token{Whitespace, EOF}},
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

		{"..*", Error},
	}
	tests.run(t)
}

func TestIdentifier(t *testing.T) {
	tests := singleTokenTests{
		{"a", Identifier},
		{"_x9", Identifier},
		{"ThisVariableIsExported", Identifier},
		{"αβ", Identifier},

		// Pre-declared identifiers.
		{"bool", Identifier},
		{"byte", Identifier},
		{"complex64", Identifier},
		{"complex128", Identifier},
		{"error", Identifier},
		{"float32", Identifier},
		{"float64", Identifier},
		{"int", Identifier},
		{"int8", Identifier},
		{"int16", Identifier},
		{"int32", Identifier},
		{"int64", Identifier},
		{"rune", Identifier},
		{"string", Identifier},
		{"uint", Identifier},
		{"uint8", Identifier},
		{"uint16", Identifier},
		{"uint32", Identifier},
		{"uint64", Identifier},
		{"uintptr", Identifier},
		{"true", Identifier},
		{"false", Identifier},
		{"iota", Identifier},
		{"nil", Identifier},
		{"append", Identifier},
		{"cap", Identifier},
		{"close", Identifier},
		{"complex", Identifier},
		{"copy", Identifier},
		{"delete", Identifier},
		{"imag", Identifier},
		{"len", Identifier},
		{"make", Identifier},
		{"new", Identifier},
		{"panic", Identifier},
		{"print", Identifier},
		{"println", Identifier},
		{"real", Identifier},
		{"recover", Identifier},
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
		{"'\\z'", Error},
		{`'\"'`, Error},
		{`'\xZ'`, Error},
		{`'\x0Z'`, Error},
		{`'\uZ'`, Error},
		{`'\u0Z'`, Error},
		{`'\u00Z'`, Error},
		{`'\u000Z'`, Error},
		{`'\UZ'`, Error},
		{`'\U0Z'`, Error},
		{`'\U00Z'`, Error},
		{`'\U000Z'`, Error},
		{`'\U0000Z'`, Error},
		{`'\U00000Z'`, Error},
		{`'\U000000Z'`, Error},
		{`'\U0000000Z'`, Error},
		{`'\Z'`, Error},
		{`'\0Z'`, Error},
		{`'\00Z'`, Error},

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
		{`"\""`, StringLiteral},

		{"\"\x0A\"", Error},
		{`"\z"`, Error},
		{`"\'"`, Error},
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
		{"\nident", [][2]int{{1, 2}, {2, 2}, {2, 2}}},
		{"\nα", [][2]int{{1, 2}, {2, 2}, {2, 2}}},
		{"\nident\n&&", [][2]int{{1, 2}, {2, 2}, {2, 2}, {2, 3}, {3, 3}}},
		{"\x60\x0A\x0A\x60", [][2]int{{1, 3}, {1, 3}}},
	}
	tests.run(t, func(start, end Location) [2]int {
		return [2]int{start.Line, end.Line}
	})
}

func TestRuneNumbers(t *testing.T) {
	tests := locTests{
		{"\n", [][2]int{{1, 2}}},
		{"hello", [][2]int{{1, 6}, {1, 6}}},
		{"\t\t\t\t\t", [][2]int{{1, 6}}},
		{"α", [][2]int{{1, 2}, {1, 2}}},
		{"αβ", [][2]int{{1, 3}, {1, 3}}},
		{"α β", [][2]int{{1, 2}, {2, 3}, {3, 4}, {3, 4}}},
		{"α\nβ", [][2]int{{1, 2}, {1, 2}, {2, 3}, {3, 4}, {3, 4}}},
	}
	tests.run(t, func(start, end Location) [2]int {
		return [2]int{start.Rune, end.Rune}
	})
}
