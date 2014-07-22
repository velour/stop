package ast

import (
	"fmt"
	"runtime"
	"strings"

	"github.com/velour/stop/token"
)

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

func (p *Parser) insertedSemicolon() bool {
	return p.tok == token.Semicolon && p.lex.Text() != ";"
}

func (p *Parser) start() token.Location {
	if p.insertedSemicolon() {
		// Inserted semicolon starts at the end of the previous token.
		return p.lex.End
	}
	return p.lex.Start
}

func (p *Parser) end() token.Location { return p.lex.End }

type span struct {
	start, end token.Location
}

func (s *span) Start() token.Location { return s.start }
func (s *span) Loc() token.Location   { return s.start }
func (s *span) End() token.Location   { return s.end }

// Span returns a span for the start and end location of the current token.
func (p *Parser) span() span {
	return span{start: p.start(), end: p.end()}
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
		if cline < p.start().Line-1 {
			p.cmnts = p.cmnts[:0]
		}
		if p.tok != token.Whitespace && p.tok != token.Comment {
			break
		}
		if p.tok == token.Comment {
			text := p.text()
			cline = p.end().Line
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

// Error returns a syntax error.  The argument must be either a string
// or a fmt.Stringer.  The syntax error states that the parse wanted
// the string value of the argument, but got the current token instead.
func (p *Parser) err(want interface{}, orWant ...interface{}) error {
	stack := make([]byte, 1024)
	n := runtime.Stack(stack, false)
	err := &SyntaxError{
		Wanted: fmt.Sprintf("%s", want),
		Got:    p.tok,
		Text:   p.text(),
		Start:  p.start(),
		End:    p.end(),
		Stack:  string(stack[:n]),
	}
	for _, w := range orWant {
		err.Wanted += fmt.Sprintf(" or %s", w)
	}
	return err
}
