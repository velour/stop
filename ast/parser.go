package ast

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/velour/stop/token"
)

// A SyntaxError is an error that describes a parse failure: something
// unexpected in the syntax of the Go source code.
type SyntaxError struct {
	// Wanted describes the value that was expected.
	Wanted string
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
	switch e.Got {
	case token.Semicolon:
		e.Text = ";"
	case token.EOF:
		e.Text = "EOF"
	}
	return fmt.Sprintf("%s: expected %s, got %s", e.Start, e.Wanted, e.Text)
}

// A MalformedLiteral is an error that describes a literal for which the
// value was malformed.  For example, an integer literal for which the
// text is not parsable as an integer.
//
// BUG(eaburns): The parser should not have to deal with this class
// of error.  The lexer should return an Error token if it scans a literal
// that is lexicographically incorrect.
type MalformedLiteral struct {
	// The name for the literal type.
	Type string
	// The malformed text.
	Text string
	// Start and End give the location of the error.
	Start, End token.Location
}

func (e *MalformedLiteral) Error() string {
	return fmt.Sprintf("%s: malformed %s: [%s]", e.Start, e.Type, e.Text)
}

// A Parser maintains the state needed to parse a stream of tokens
// representing Go source code.
type Parser struct {
	lex *token.Lexer
	tok token.Token

	// ExprLevel is the nesting level of expressions.  When the level
	// is greater or equal to zero, composite literals are allowed.
	// It is negative when parsing the initialization statement of an
	// if, switch, or for statement.
	exprLevel int

	// Cmnts is a slice of all comments that are preceeding the
	// current token without an intervening blank line.
	cmnts []string
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

// Comments returns a new slice containing the comments
// preceeding the current token without an intervening blank line.
// If there are no comments the nil is returned.
func (p *Parser) comments() []string {
	if len(p.cmnts) == 0 {
		return nil
	}
	c := make([]string, len(p.cmnts))
	copy(c, p.cmnts)
	return c
}

// Advances to the next non-whitespace, non-comment token.
func (p *Parser) next() {
	p.tok = p.lex.Next()
	cline := -1
	for {
		if cline < p.lex.Start.Line-1 {
			p.cmnts = p.cmnts[:0]
		}
		if p.tok != token.Whitespace && p.tok != token.Comment {
			break
		}
		if p.tok == token.Comment {
			text := p.text()
			cline = p.lex.End.Line
			if strings.HasPrefix(text, "//") {
				// Line comments end on the line following their
				// text, so subtract one.
				cline--
				// Strip the newline.
				text = text[:len(text)-1]
			}
			p.cmnts = append(p.cmnts, text)
		}
		p.tok = p.lex.Next()
	}
}

// Expect panics with an unexpected token error if the current parser
// token is not tok.
func (p *Parser) expect(tok token.Token) {
	if p.tok != tok {
		panic(p.err(tok.String()))
	}
}

type stringStringer string

func (s stringStringer) String() string {
	return string(s)
}

// Error returns a syntax error.  The argument must be either a string
// or a fmt.Stringer.  The syntax error states that the parse wanted
// the string value of the argument, but got the current token instead.
func (p *Parser) err(want interface{}, orWant ...interface{}) error {
	err := &SyntaxError{
		Wanted: fmt.Sprintf("%s", want),
		Got:    p.tok,
		Text:   p.text(),
		Start:  p.lex.Start,
		End:    p.lex.End,
	}
	for _, w := range orWant {
		err.Wanted += fmt.Sprintf(" or %s", w)
	}
	return err
}
