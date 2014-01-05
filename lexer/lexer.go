// Package lexer provides a lexer for the Go language.
//
// Quirks:
//
// Mal-formed octal literals that begin with 0 followed by any
// number of decmial digits are returned as valid integer literals.
//
// Unicode code points are not validated, so characters with
// invalid code points (such as '\U00110000' and '\uDFFF') are
// returned as valid character literals.
package lexer

import (
	"bufio"
	"io"
	"os"
	"strings"
	"unicode"

	"bitbucket.org/eaburns/stop/loc"
)

// A TokenType identifies the type of a token in the input file.
type TokenType int

// The set of constants defining the types of tokens.
const (
	EOF   TokenType = -1
	Error TokenType = iota
	Identifier
	IntegerLiteral
	FloatLiteral
	ImaginaryLiteral
	RuneLiteral
	StringLiteral
	Comment
	Whitespace

	// Keywords

	Break
	Default
	Func
	Interface
	Select
	Case
	Defer
	Go
	Map
	Struct
	Chan
	Else
	Goto
	Package
	Switch
	Const
	Fallthrough
	If
	Range
	Type
	Continue
	For
	Import
	Return
	Var

	// Operators

	Plus
	And
	PlusEqual
	AndEqual
	AndAnd
	EqualEqual
	BangEqual
	OpenParen
	CloseParen
	Minus
	Or
	MinusEqual
	OrEqual
	OrOr
	Less
	LessEqual
	OpenBracket
	CloseBracket
	Star
	Carrot
	StarEqual
	CarrotEqual
	LessMinus
	Greater
	GreaterEqual
	OpenBrace
	CloseBrace
	Divide
	LessLess
	DivideEqual
	LessLessEqual
	PlusPlus
	Equal
	ColonEqual
	Comma
	Semicolon
	Percent
	GreaterGreater
	PercentEqual
	GreaterGreaterEqual
	MinusMinus
	Bang
	DotDotDot
	Dot
	Colon
	AndCarrot
	AndCarrotEqual
)

var tokenTypeNames = map[TokenType]string{
	EOF:                 "EOF",
	Error:               "Error",
	Identifier:          "Identifier",
	IntegerLiteral:      "IntegerLiteral",
	FloatLiteral:        "FloatLiteral",
	ImaginaryLiteral:    "ImaginaryLiteral",
	RuneLiteral:         "RuneLiteral",
	StringLiteral:       "StringLiteral",
	Comment:             "Comment",
	Whitespace:          "Whitespace",
	Break:               "Break",
	Default:             "Default",
	Func:                "Func",
	Interface:           "Interface",
	Select:              "Select",
	Case:                "Case",
	Defer:               "Defer",
	Go:                  "Go",
	Map:                 "Map",
	Struct:              "Struct",
	Chan:                "Chan",
	Else:                "Else",
	Goto:                "Goto",
	Package:             "Package",
	Switch:              "Switch",
	Const:               "Const",
	Fallthrough:         "Fallthrough",
	If:                  "If",
	Range:               "Range",
	Type:                "Type",
	Continue:            "Continue",
	For:                 "For",
	Import:              "Import",
	Return:              "Return",
	Var:                 "Var",
	Plus:                "Plus",
	And:                 "And",
	PlusEqual:           "PlusEqual",
	AndEqual:            "AndEqual",
	AndAnd:              "AndAnd",
	EqualEqual:          "EqualEqual",
	BangEqual:           "BangEqual",
	OpenParen:           "OpenParen",
	CloseParen:          "CloseParen",
	Minus:               "Minus",
	Or:                  "Or",
	MinusEqual:          "MinusEqual",
	OrEqual:             "OrEqual",
	OrOr:                "OrOr",
	Less:                "Less",
	LessEqual:           "LessEqual",
	OpenBracket:         "OpenBracket",
	CloseBracket:        "CloseBracket",
	Star:                "Star",
	Carrot:              "Carrot",
	StarEqual:           "StarEqual",
	CarrotEqual:         "CarrotEqual",
	LessMinus:           "LessMinus",
	Greater:             "Greater",
	GreaterEqual:        "GreaterEqual",
	OpenBrace:           "OpenBrace",
	CloseBrace:          "CloseBrace",
	Divide:              "Divide",
	LessLess:            "LessLess",
	DivideEqual:         "DivideEqual",
	LessLessEqual:       "LessLessEqual",
	PlusPlus:            "PlusPlus",
	Equal:               "Equal",
	ColonEqual:          "ColonEqual",
	Comma:               "Comma",
	Semicolon:           "Semicolon",
	Percent:             "Percent",
	GreaterGreater:      "GreaterGreater",
	PercentEqual:        "PercentEqual",
	GreaterGreaterEqual: "GreaterGreaterEqual",
	MinusMinus:          "MinusMinus",
	Bang:                "Bang",
	DotDotDot:           "DotDotDot",
	Dot:                 "Dot",
	Colon:               "Colon",
	AndCarrot:           "AndCarrot",
	AndCarrotEqual:      "AndCarrotEqual",
}

// String returns the string represenation of the token type.
func (tt TokenType) String() string {
	return tokenTypeNames[tt]
}

var keywords = map[string]TokenType{
	"break":       Break,
	"default":     Default,
	"func":        Func,
	"interface":   Interface,
	"select":      Select,
	"case":        Case,
	"defer":       Defer,
	"go":          Go,
	"map":         Map,
	"struct":      Struct,
	"chan":        Chan,
	"else":        Else,
	"goto":        Goto,
	"package":     Package,
	"switch":      Switch,
	"const":       Const,
	"fallthrough": Fallthrough,
	"if":          If,
	"range":       Range,
	"type":        Type,
	"continue":    Continue,
	"for":         For,
	"import":      Import,
	"return":      Return,
	"var":         Var,
}

var operators = map[string]TokenType{
	"+":   Plus,
	"&":   And,
	"+=":  PlusEqual,
	"&=":  AndEqual,
	"&&":  AndAnd,
	"==":  EqualEqual,
	"!=":  BangEqual,
	"(":   OpenParen,
	")":   CloseParen,
	"-":   Minus,
	"|":   Or,
	"-=":  MinusEqual,
	"|=":  OrEqual,
	"||":  OrOr,
	"<":   Less,
	"<=":  LessEqual,
	"[":   OpenBracket,
	"]":   CloseBracket,
	"*":   Star,
	"^":   Carrot,
	"*=":  StarEqual,
	"^=":  CarrotEqual,
	"<-":  LessMinus,
	">":   Greater,
	">=":  GreaterEqual,
	"{":   OpenBrace,
	"}":   CloseBrace,
	"/":   Divide,
	"<<":  LessLess,
	"/=":  DivideEqual,
	"<<=": LessLessEqual,
	"++":  PlusPlus,
	"=":   Equal,
	":=":  ColonEqual,
	",":   Comma,
	";":   Semicolon,
	"%":   Percent,
	">>":  GreaterGreater,
	"%=":  PercentEqual,
	">>=": GreaterGreaterEqual,
	"--":  MinusMinus,
	"!":   Bang,
	"...": DotDotDot,
	".":   Dot,
	":":   Colon,
	"&^":  AndCarrot,
	"&^=": AndCarrotEqual,
}

// Token is a the atomic unit of the vocabulary of a Go program.
type Token struct {
	Type TokenType
	Text string
	Span loc.Span
}

// String returns a human-readable string representation of the token.
func (t *Token) String() string {
	return "Token{Type:" + t.Type.String() + ", Text:`" + t.Text + "`, Span:" + t.Span.String() + "}"
}

// IsNewline returns true if the token acts as a newline.
func (t *Token) IsNewline() bool {
	if t.Type != Whitespace && t.Type != Comment {
		return false
	}
	// The first clause is a special check for line comments,
	// because they may not end with a \n if they occur
	// immediately before EOF.
	return t.Type == Comment && strings.HasPrefix(t.Text, "//") || strings.ContainsRune(t.Text, '\n')
}

// A Lexer lexes Go tokens from an input stream.
type Lexer struct {
	in            *bufio.Reader
	text          []rune
	span          loc.Span
	prevLineStart int

	prev *Token
	next *Token
}

// New returns a new Lexer that reads tokens from an input file.
// If the underlying type of the reader is an os.File, then all Locations
// produced by the lexer use the file name as their path.
func New(in io.Reader) *Lexer {
	path := ""
	if file, ok := in.(*os.File); ok {
		path = file.Name()
	}
	return &Lexer{
		in:            bufio.NewReader(in),
		span:          loc.Span{0: loc.Zero(path), 1: loc.Zero(path)},
		prevLineStart: -1,
	}
}

// Rune reads and returns the next rune from the input stream and updates the
// span and text of the current token.  If an error is encountered then it panicks.
// If the end of the input is reached -1 is returned.
func (l *Lexer) rune() rune {
	if l.in == nil {
		return -1
	}
	r, _, err := l.in.ReadRune()
	if err != nil && err != io.EOF {
		panic(err)
	}
	loc := &l.span[1]
	if err == io.EOF {
		l.in = nil
		return -1
	}
	loc.Rune++
	if r == '\n' {
		loc.Line++
		l.prevLineStart = loc.LineStart
		loc.LineStart = loc.Rune
	}
	l.text = append(l.text, r)
	return r
}

// Replace replaces the most-recently-read rune into the input stream.  If an
// error is encountered then it panicks.
func (l *Lexer) replace() {
	if len(l.text) == 0 {
		panic("nothing to replace")
	}
	if l.in == nil {
		return
	}
	l.text = l.text[:len(l.text)-1]
	if err := l.in.UnreadRune(); err != nil {
		panic(err)
	}

	loc := &l.span[1]
	loc.Rune--
	if loc.Rune < loc.LineStart {
		loc.Line--
		loc.LineStart = l.prevLineStart
		l.prevLineStart = -1
	}
}

func (l *Lexer) token(typ TokenType) *Token {
	t := &Token{
		Text: string(l.text),
		Type: typ,
		Span: l.span,
	}
	l.span[0] = l.span[1]
	l.text = l.text[:0]
	return t
}

func (l *Lexer) unexpected(r rune) *Token {
	l.span[0] = l.span[1]
	l0 := &l.span[0]
	l0.Rune--
	if l0.Rune < l0.LineStart {
		l0.Line--
		l0.LineStart = l.prevLineStart
	}
	l.in = nil
	return &Token{
		Text: "unexpected input rune: " + string([]rune{r}),
		Type: Error,
		Span: l.span,
	}
}

func (l *Lexer) nextLiteralToken() *Token {
	r := l.rune()
	switch {
	case r < 0:
		return l.token(EOF)
	case isWhitespace(r):
		return whitespace(l)
	case isIdentStart(r):
		return identifier(l)
	case isDecimalDigit(r) || isHexDigit(r):
		return number(r, l)
	case r == '^':
		if l.rune() != '=' {
			l.replace()
		}
		return operator(l)
	case r == '<':
		if r = l.rune(); r == '<' {
			if r = l.rune(); r != '=' {
				l.replace()
			}
		} else if r != '=' && r != '-' {
			l.replace()
		}
		return operator(l)
	case r == '=' || r == ':' || r == '!' || r == '*' || r == '%':
		if r = l.rune(); r != '=' {
			l.replace()
		}
		return operator(l)
	case r == '>':
		if r = l.rune(); r == '>' {
			if r = l.rune(); r != '=' {
				l.replace()
			}
		} else if r != '=' {
			l.replace()
		}
		return operator(l)
	case r == '|':
		if r = l.rune(); r != '=' && r != '|' {
			l.replace()
		}
		return operator(l)
	case r == '-':
		if r = l.rune(); r != '-' && r != '=' {
			l.replace()
		}
		return operator(l)
	case r == ',' || r == ';' || r == '(' || r == ')' || r == '[' || r == ']' || r == '{' || r == '}':
		return operator(l)
	case r == '/':
		switch r := l.rune(); {
		case r == '/':
			return comment(l, []rune{'\n'})
		case r == '*':
			return comment(l, []rune{'*', '/'})
		case r != '=':
			l.replace()
		}
		return operator(l)
	case r == '.':
		if r = l.rune(); r == '.' {
			if r = l.rune(); r != '.' {
				return l.unexpected(r)
			}
		} else if isDecimalDigit(r) {
			return fraction(l)
		} else {
			l.replace()
		}
		return operator(l)
	case r == '&':
		if r = l.rune(); r == '^' {
			if r = l.rune(); r != '=' {
				l.replace()
			}
		} else if r != '=' && r != '&' {
			l.replace()
		}
		return operator(l)
	case r == '+':
		if r = l.rune(); r != '+' && r != '=' {
			l.replace()
		}
		return operator(l)
	case r == '\'':
		return runeLiteral(l)
	case r == '"':
		return interpertedStringLiteral(l)
	case r == '`':
		return rawStringLiteral(l)
	}
	return l.unexpected(r)
}

// Next returns the next logical token from the token stream.
// Once an Error or EOF token is returned, all subsequent tokens will
// be of type EOF.
func (l *Lexer) Next() *Token {
	tok := l.next
	l.next = nil
	if tok == nil {
		tok = l.nextLiteralToken()
	}
	if tok.IsNewline() && l.prev != nil &&
		(l.prev.Type == Identifier ||
			l.prev.Type == IntegerLiteral ||
			l.prev.Type == FloatLiteral ||
			l.prev.Type == ImaginaryLiteral ||
			l.prev.Type == RuneLiteral ||
			l.prev.Type == StringLiteral ||
			l.prev.Type == Break ||
			l.prev.Type == Continue ||
			l.prev.Type == Fallthrough ||
			l.prev.Type == Return ||
			l.prev.Type == PlusPlus ||
			l.prev.Type == MinusMinus ||
			l.prev.Type == CloseParen ||
			l.prev.Type == CloseBracket ||
			l.prev.Type == CloseBrace) {
		l.next = tok
		tok = &Token{
			Type: Semicolon,
			Span: loc.Span{0: l.next.Span[0], 1: l.next.Span[0]},
		}
	}
	l.prev = tok
	return tok
}

func operator(l *Lexer) *Token {
	tok := l.token(Error)
	oper, ok := operators[tok.Text]
	if !ok {
		panic("bad operator: \"" + tok.Text + "\"")
	}
	tok.Type = oper
	return tok
}

func identifier(l *Lexer) *Token {
	r := l.rune()
	for isIdent(r) {
		r = l.rune()
	}
	l.replace()
	tok := l.token(Identifier)
	if keyword, ok := keywords[tok.Text]; ok {
		tok.Type = keyword
	}
	return tok
}

func isIdentStart(r rune) bool {
	return unicode.IsLetter(r) || r == '_'
}

func isIdent(r rune) bool {
	return isIdentStart(r) || unicode.IsDigit(r)
}

func whitespace(l *Lexer) *Token {
	r := l.rune()
	for isWhitespace(r) {
		r = l.rune()
	}
	l.replace()
	return l.token(Whitespace)
}

func isWhitespace(r rune) bool {
	return r == ' ' || r == '\t' || r == '\v' || r == '\r' || r == '\n'
}

func comment(l *Lexer, closing []rune) *Token {
	i := 0
	for i < len(closing) {
		r := l.rune()
		switch {
		case r < 0 && closing[i] == '\n':
			r = '\n'
			i++
		case r < 0:
			panic("unexpected EOF")
		case r != closing[i]:
			i = 0
		default:
			i++
		}
	}
	return l.token(Comment)
}

func number(r0 rune, l *Lexer) *Token {
	r := l.rune()
	if r0 == '0' && (r == 'x' || r == 'X') {
		return hex(l)
	}
	for {
		switch {
		case r == 'e' || r == 'E':
			return mantissa(l)
		case r == '.':
			return fraction(l)
		case r == 'i':
			return l.token(ImaginaryLiteral)
		case !isDecimalDigit(r):
			l.replace()
			return l.token(IntegerLiteral)
		}
		r = l.rune()
	}
}

func mantissa(l *Lexer) *Token {
	r := l.rune()
	if r == '+' || r == '-' {
		r = l.rune()
	}
	for isDecimalDigit(r) {
		r = l.rune()
	}
	if r == 'i' {
		return l.token(ImaginaryLiteral)
	}
	l.replace()
	return l.token(FloatLiteral)
}

func fraction(l *Lexer) *Token {
	r := l.rune()
	for isDecimalDigit(r) {
		r = l.rune()
	}
	switch {
	case r == 'i':
		return l.token(ImaginaryLiteral)
	case r == 'e' || r == 'E':
		return mantissa(l)
	}
	l.replace()
	return l.token(FloatLiteral)
}

func hex(l *Lexer) *Token {
	r := l.rune()
	for isHexDigit(r) {
		r = l.rune()
	}
	l.replace()
	return l.token(IntegerLiteral)
}

func runeLiteral(l *Lexer) *Token {
	if l.rune() == '\\' {
		if err := unicodeValue(l); err != nil {
			return err
		}
	}
	if r := l.rune(); r != '\'' {
		return l.unexpected(r)
	}
	return l.token(RuneLiteral)
}

func interpertedStringLiteral(l *Lexer) *Token {
	for {
		r := l.rune()
		switch {
		case r == '"':
			return l.token(StringLiteral)
		case r == '\\':
			if err := unicodeValue(l); err != nil {
				return err
			}
		case r < 0:
			return l.unexpected(r)
		}
	}
}

func rawStringLiteral(l *Lexer) *Token {
	for {
		r := l.rune()
		switch {
		case r == '`':
			return l.token(StringLiteral)
		case r < 0:
			return l.unexpected(r)
		}
	}
}

// Parses a unicode value (assuming that the leading '\' has
// already been consumed), and returns nil on success or
// an error token if the parse failed.
func unicodeValue(l *Lexer) *Token {
	r := l.rune()
	switch {
	case r == 'U':
		if s := l.rune(); !isHexDigit(s) {
			return l.unexpected(s)
		}
		if s := l.rune(); !isHexDigit(s) {
			return l.unexpected(s)
		}
		if s := l.rune(); !isHexDigit(s) {
			return l.unexpected(s)
		}
		if s := l.rune(); !isHexDigit(s) {
			return l.unexpected(s)
		}
		fallthrough

	case r == 'u':
		if s := l.rune(); !isHexDigit(s) {
			return l.unexpected(s)
		}
		if s := l.rune(); !isHexDigit(s) {
			return l.unexpected(s)
		}
		fallthrough

	case r == 'x':
		if s := l.rune(); !isHexDigit(s) {
			return l.unexpected(s)
		}
		if s := l.rune(); !isHexDigit(s) {
			return l.unexpected(s)
		}

	case isOctalDigit(r):
		if s := l.rune(); !isOctalDigit(s) {
			return l.unexpected(s)
		}
		if s := l.rune(); !isOctalDigit(s) {
			return l.unexpected(s)
		}

	case r != 'a' && r != 'b' && r != 'f' && r != 'n' && r != 'r' && r != 't' && r != 'v' && r != '\\' && r != '\'' && r != '"':
		l.replace()
	}
	return nil
}

func isDecimalDigit(r rune) bool {
	return r >= '0' && r <= '9'
}

func isOctalDigit(r rune) bool {
	return r >= '0' && r <= '7'
}

func isHexDigit(r rune) bool {
	return (r >= '0' && r <= '9') || (r >= 'A' && r <= 'F') || (r >= 'a' && r <= 'f')
}
