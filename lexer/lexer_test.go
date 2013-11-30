package lexer

import (
	"strings"
	"testing"

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
