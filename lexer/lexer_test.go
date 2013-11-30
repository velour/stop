package lexer

import (
	"strings"
	"testing"
	"unicode/utf8"

	"gogo/loc"
)

type test struct {
	text string
	want []Token
}

func (test test) run(num int, t *testing.T) {
	in := strings.NewReader(test.text)
	lex := New(in)
	var got []*Token
	for len(got) == 0 || got[len(got)-1].Type != TokenEOF {
		got = append(got, lex.Next())
	}
	for i, want := range test.want {
		if tokenEq(got[i], &want) {
			continue
		}
		t.Fatalf("test %d:\n%s:\ntoken %d: got %s, wanted %s",
			num, test.text, i, got, test.want)
	}
}

func tokenEq(a, b *Token) bool {
	return a.Type == b.Type &&
		a.Text == b.Text &&
		a.Span[0].Rune == b.Span[0].Rune &&
		a.Span[1].Rune == b.Span[1].Rune &&
		a.Keyword == b.Keyword &&
		a.Operator == b.Operator
}

func span(a, b int) loc.Span {
	return loc.Span{{Rune: a}, {Rune: b}}
}

func TestEOF(t *testing.T) {
	tests := [...]test{
		{"", []Token{
			{
				Type: TokenNewline,
				Span: span(1, 1),
			},
			{
				Type: TokenEOF,
				Span: span(1, 1),
			},
		}},
		{"\n", []Token{
			{
				Type: TokenNewline,
				Text: "\n",
				Span: span(1, 2),
			},
			{
				Type: TokenEOF,
				Span: span(2, 2),
			},
		}},
	}

	for i, test := range tests {
		test.run(i, t)
	}

}

func TestLineComment(t *testing.T) {
	tests := [...]test{
		{"//\n", []Token{
			{
				Type: TokenNewline,
				Text: "//\n",
				Span: span(1, 4),
			},
			{
				Type: TokenEOF,
				Span: span(4, 4),
			},
		}},
		{"// hello\n", []Token{
			{
				Type: TokenNewline,
				Text: "// hello\n",
				Span: span(1, 10),
			},
			{
				Type: TokenEOF,
				Span: span(10, 10),
			},
		}},
		{"//", []Token{
			{
				Type: TokenNewline,
				Text: "//",
				Span: span(1, 3),
			},
			{
				Type: TokenEOF,
				Span: span(3, 3),
			},
		}},
		{"\n//", []Token{
			{
				Type: TokenNewline,
				Text: "\n",
				Span: span(1, 2),
			},
			{
				Type: TokenNewline,
				Text: "//",
				Span: span(2, 4),
			},
			{
				Type: TokenEOF,
				Span: span(4, 4),
			},
		}},
	}

	for i, test := range tests {
		test.run(i, t)
	}
}

func TestGeneralComment(t *testing.T) {
	tests := [...]test{
		{"/* comment */", []Token{
			{
				Type: TokenWhitespace,
				Text: "/* comment */",
				Span: span(1, 14),
			},
			{
				Type: TokenNewline,
				Span: span(14, 14),
			},
			{
				Type: TokenEOF,
				Span: span(14, 14),
			},
		}},
		{"/*\ncomment */", []Token{
			{
				Type: TokenNewline,
				Text: "/*\ncomment */",
				Span: span(1, 14),
			},
			{
				Type: TokenEOF,
				Span: span(14, 14),
			},
		}},
		{"/*\ncomment\n*/", []Token{
			{
				Type: TokenNewline,
				Text: "/*\ncomment\n*/",
				Span: span(1, 14),
			},
			{
				Type: TokenEOF,
				Span: span(14, 14),
			},
		}},
		{"/* // comment */", []Token{
			{
				Type: TokenWhitespace,
				Text: "/* // comment */",
				Span: span(1, 17),
			},
			{
				Type: TokenNewline,
				Span: span(17, 17),
			},
			{
				Type: TokenEOF,
				Span: span(17, 17),
			},
		}},
		{"/* // /*comment*/ */", []Token{
			{
				Type: TokenWhitespace,
				Text: "/* // /*comment*/",
				Span: span(1, 18),
			},
			{
				Type: TokenWhitespace,
				Text: " ",
				Span: span(18, 19),
			},
			{
				Type:     TokenOperator,
				Text:     "*",
				Span:     span(19, 20),
				Operator: OpStar,
			},
			{
				Type:     TokenOperator,
				Text:     "/",
				Span:     span(20, 21),
				Operator: OpDivide,
			},
			{
				Type: TokenNewline,
				Text: "",
				Span: span(21, 21),
			},
			{
				Type: TokenEOF,
				Text: "",
				Span: span(21, 21),
			},
		}},
	}

	for i, test := range tests {
		test.run(i, t)
	}
}

func TestSemicolonInsertion(t *testing.T) {
	tests := [...]test{
		{"identifier\n", []Token{
			{
				Type: TokenIdentifier,
				Text: "identifier",
				Span: span(1, 11),
			},
			{
				Type:     TokenOperator,
				Span:     span(11, 11),
				Operator: OpSemicolon,
			},
			{
				Type: TokenNewline,
				Text: "\n",
				Span: span(11, 12),
			},
			{
				Type: TokenEOF,
				Span: span(12, 12),
			},
		}},
		{"identifier", []Token{
			{
				Type: TokenIdentifier,
				Text: "identifier",
				Span: span(1, 11),
			},
			{
				Type:     TokenOperator,
				Span:     span(11, 11),
				Operator: OpSemicolon,
			},
			{
				Type: TokenNewline,
				Span: span(11, 11),
			},
			{
				Type: TokenEOF,
				Span: span(11, 11),
			},
		}},
		{"++", []Token{
			{
				Type:     TokenOperator,
				Text:     "++",
				Span:     span(1, 3),
				Operator: OpPlusPlus,
			},
			{
				Type:     TokenOperator,
				Span:     span(3, 3),
				Operator: OpSemicolon,
			},
			{
				Type: TokenNewline,
				Span: span(3, 3),
			},
			{
				Type: TokenEOF,
				Span: span(3, 3),
			},
		}},
		{"++//hi", []Token{
			{
				Type:     TokenOperator,
				Text:     "++",
				Span:     span(1, 3),
				Operator: OpPlusPlus,
			},
			{
				Type:     TokenOperator,
				Span:     span(3, 3),
				Operator: OpSemicolon,
			},
			{
				Type: TokenNewline,
				Text: "//hi",
				Span: span(3, 7),
			},
			{
				Type: TokenEOF,
				Span: span(7, 7),
			},
		}},
	}

	for i, test := range tests {
		test.run(i, t)
	}
}

func TestKeywords(t *testing.T) {
	var tests []test
	for text, keyword := range keywords {
		if keyword == KeywordNone {
			continue
		}
		end := len(text) + 1
		toks := []Token{
			{
				Type:    TokenKeyword,
				Text:    text,
				Span:    span(1, end),
				Keyword: keyword,
			},
			{
				Type: TokenNewline,
				Span: span(end, end),
			},
			{
				Type: TokenEOF,
				Span: span(end, end),
			},
		}
		switch keyword {
		case KeywordBreak, KeywordContinue, KeywordFallthrough, KeywordReturn:
			semi := Token{
				Type:     TokenOperator,
				Span:     span(end, end),
				Operator: OpSemicolon,
			}
			toks = []Token{toks[0], semi, toks[1], toks[2]}
		}
		tests = append(tests, test{
			text: text,
			want: toks,
		})
	}

	for i, test := range tests {
		test.run(i, t)
	}
}

func TestOperators(t *testing.T) {
	var tests []test
	for text, op := range operators {
		if op == OpNone {
			continue
		}
		end := len(text) + 1
		toks := []Token{
			{
				Type:     TokenOperator,
				Text:     text,
				Span:     span(1, end),
				Operator: op,
			},
			{
				Type: TokenNewline,
				Span: span(end, end),
			},
			{
				Type: TokenEOF,
				Span: span(end, end),
			},
		}
		switch op {
		case OpPlusPlus, OpMinusMinus, OpCloseParen, OpCloseBracket, OpCloseBrace:
			semi := Token{
				Type:     TokenOperator,
				Span:     span(end, end),
				Operator: OpSemicolon,
			}
			toks = []Token{toks[0], semi, toks[1], toks[2]}
		}
		tests = append(tests, test{
			text: text,
			want: toks,
		})
	}

	for i, test := range tests {
		test.run(i, t)
	}
}

func TestIdentifier(t *testing.T) {
	texts := []string{
		"a",
		"_x9",
		"ThisVariableIsExported",
		"αβ",
	}
	literalThenSemicolon(t, texts, TokenIdentifier)
}

func TestIntegerLiteral(t *testing.T) {
	texts := []string{
		"42",
		"0600",
		"0xBadFace",
		"170141183460469231731687303715884105727",
		"0",
		"01",
		"0777",
		"0xF",
		"0XF",
		"9832",
	}
	literalThenSemicolon(t, texts, TokenIntegerLiteral)
}

func TestFloatLiteral(t *testing.T) {
	texts := []string{
		"0.",
		"72.40",
		"072.40",
		"2.71828",
		"1.e+0",
		"6.67428e-11",
		"1E6",
		".25",
		".12345E+5",
	}
	literalThenSemicolon(t, texts, TokenFloatLiteral)
}

func TestImaginaryLiteral(t *testing.T) {
	texts := []string{
		"0i",
		"011i",
		"0.i",
		"2.71828i",
		"1.e+0i",
		"6.67428e-11i",
		"1E6i",
		".25i",
		".12345E+5i",
	}
	literalThenSemicolon(t, texts, TokenImaginaryLiteral)
}

func literalThenSemicolon(t *testing.T, texts []string, kind TokenType) {
	var tests []test
	for _, text := range texts {
		end := utf8.RuneCountInString(text) + 1
		tests = append(tests, test{
			text: text,
			want: []Token{
				{
					Type: kind,
					Text: text,
					Span: span(1, end),
				},
				{
					Type:     TokenOperator,
					Span:     span(end, end),
					Operator: OpSemicolon,
				},
				{
					Type: TokenNewline,
					Span: span(end, end),
				},
				{
					Type: TokenEOF,
					Span: span(end, end),
				},
			},
		})
	}

	for i, test := range tests {
		test.run(i, t)
	}
}
