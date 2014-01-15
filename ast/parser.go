package ast

import (
	"fmt"
	"strconv"

	"bitbucket.org/eaburns/stop/token"
)

// A SyntaxError is an error that describes a parse failure: something
// unexpected in the syntax of the Go source code.
type SyntaxError struct {
	// Wanted describes the value that was expected.
	Wanted fmt.Stringer
	// Got describes the unexpected token.
	Got fmt.Stringer
	// Text is the text of the unexpected token.
	Text string
	// Start and End give the location of the error.
	Start, End token.Location
}

func (e *SyntaxError) Error() string {
	if e.Got == token.Error {
		text := strconv.QuoteToASCII(e.Text)
		return fmt.Sprintf("%s: unexpected rune in input [%s]", e.Start, text[1:len(text)-1])
	}
	if e.Got == token.Semicolon {
		// Text from inserted ; is the text of the previous token.
		e.Text = ";"
	}
	return fmt.Sprintf("%s: expected %s, got %s", e.Start, e.Wanted, e.Text)
}

// A Parser maintains the state needed to parse a stream of tokens
// representing Go source code.
type Parser struct {
	lex *token.Lexer
	tok token.Token
}

// NewParser returns a new parser that parses from the given token.Lexer.
func NewParser(lex *token.Lexer) *Parser {
	p := &Parser{lex: lex}
	p.next()
	return p
}

type span struct {
	start, end token.Location
}

func (s *span) Start() token.Location { return s.start }
func (s *span) Loc() token.Location   { return s.start }
func (s *span) End() token.Location   { return s.end }

// Span returns a span for the start and end location of the current token.
func (p *Parser) span() span {
	return span{start: p.lex.Start, end: p.lex.End}
}

// Text returns the text of the current token.  The returned value is
// a copy of the text so that it does not reference the lexer's source
// string.
func (p *Parser) text() string {
	return string([]byte(p.lex.Text()))
}

// Advances to the next non-whitespace, non-comment token.
func (p *Parser) next() {
	p.tok = p.lex.Next()
	for p.tok == token.Whitespace || p.tok == token.Comment {
		p.tok = p.lex.Next()
	}
}

// Expect panicks with an unexpected token error if the current parser
// token is not tok.
func (p *Parser) expect(tok token.Token) {
	if p.tok != tok {
		panic(p.err(tok))
	}
}

type stringStringer string

func (s stringStringer) String() string {
	return string(s)
}

// Error returns a syntax error.  The argument must be either a string
// or a fmt.Stringer.  The syntax error states that the parse wanted
// the string value of the argument, but got the current token instead.
func (p *Parser) err(wanted interface{}) error {
	w, ok := wanted.(fmt.Stringer)
	if !ok {
		str, ok := wanted.(string)
		if !ok {
			panic("Parser.error needs a string or fmt.Stringer")
		}
		w = stringStringer(str)
	}
	return &SyntaxError{
		Wanted: w,
		Got:    p.tok,
		Text:   p.text(),
		Start:  p.lex.Start,
		End:    p.lex.End,
	}
}
