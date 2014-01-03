package lexer

import (
	"bufio"
	"io"
	"os"
	"unicode"

	"bitbucket.org/eaburns/stop/loc"
)

// A TokenType identifies the type of a token in the input file.
type TokenType int

const (
	EOF   TokenType = -1
	Error TokenType = iota
	Identifier
	IntegerLiteral
	FloatLiteral
	ImaginaryLiteral
	RuneLiteral
	StringLiteral
	Newline
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
	enParen
	CloseParen
	Minus
	Or
	MinusEqual
	OrEqual
	OrOr
	Less
	LessEqual
	enBracket
	CloseBracket
	Star
	Carrot
	StarEqual
	CarrotEqual
	LessMinus
	Greater
	GreaterEqual
	enBrace
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
	Newline:             "Newline",
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
	enParen:             "enParen",
	CloseParen:          "CloseParen",
	Minus:               "Minus",
	Or:                  "Or",
	MinusEqual:          "MinusEqual",
	OrEqual:             "OrEqual",
	OrOr:                "OrOr",
	Less:                "Less",
	LessEqual:           "LessEqual",
	enBracket:           "enBracket",
	CloseBracket:        "CloseBracket",
	Star:                "Star",
	Carrot:              "Carrot",
	StarEqual:           "StarEqual",
	CarrotEqual:         "CarrotEqual",
	LessMinus:           "LessMinus",
	Greater:             "Greater",
	GreaterEqual:        "GreaterEqual",
	enBrace:             "enBrace",
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
	"(":   enParen,
	")":   CloseParen,
	"-":   Minus,
	"|":   Or,
	"-=":  MinusEqual,
	"|=":  OrEqual,
	"||":  OrOr,
	"<":   Less,
	"<=":  LessEqual,
	"[":   enBracket,
	"]":   CloseBracket,
	"*":   Star,
	"^":   Carrot,
	"*=":  StarEqual,
	"^=":  CarrotEqual,
	"<-":  LessMinus,
	">":   Greater,
	">=":  GreaterEqual,
	"{":   enBrace,
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
func (t Token) String() string {
	return "Token{Type:" + t.Type.String() + ", Text:`" + t.Text + "`, Span:" + t.Span.String() + "}"
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
	case r == '\n':
		return l.token(Newline)
	case isWhitespace(r):
		return whitespace(l)
	case isIdentStart(r):
		return identifier(l)
	case isDecimalDigit(r) || isOctalDigit(r) || isHexDigit(r):
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
	}
	return l.unexpected(r)
}

// Next returns the next logical token from the token stream.
// Semicolons are inserted, and if there is no newline before the end of
// the file, on is added.
func (l *Lexer) Next() *Token {
	tok := l.next
	l.next = nil
	if tok == nil {
		tok = l.nextLiteralToken()
	}
	if tok.Type == EOF && (l.prev == nil || l.prev.Type != Newline) {
		l.next = tok
		tok = &Token{
			Type: Newline,
			Span: tok.Span,
		}
	}
	if tok.Type == Newline && l.prev != nil && needsSemicolon(l.prev) {
		l.next = tok
		tok = &Token{
			Type: Semicolon,
			Span: loc.Span{0: l.next.Span[0], 1: l.next.Span[0]},
		}
	}
	l.prev = tok
	return tok
}

func needsSemicolon(tok *Token) bool {
	return tok.Type == Identifier ||
		tok.Type == IntegerLiteral ||
		tok.Type == FloatLiteral ||
		tok.Type == ImaginaryLiteral ||
		tok.Type == RuneLiteral ||
		tok.Type == StringLiteral ||
		tok.Type == Break ||
		tok.Type == Continue ||
		tok.Type == Fallthrough ||
		tok.Type == Return ||
		tok.Type == PlusPlus ||
		tok.Type == MinusMinus ||
		tok.Type == CloseParen ||
		tok.Type == CloseBracket ||
		tok.Type == CloseBrace
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
	return r == ' ' || r == '\t' || r == '\v' || r == '\r'
}

func comment(l *Lexer, closing []rune) *Token {
	i := 0
	typ := Whitespace
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
		if r == '\n' {
			typ = Newline
		}
	}
	return l.token(typ)
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

func isDecimalDigit(r rune) bool {
	return r >= '0' && r <= '9'
}

func isOctalDigit(r rune) bool {
	return r >= '0' && r <= '7'
}

func isHexDigit(r rune) bool {
	return (r >= '0' && r <= '9') || (r >= 'A' && r <= 'F') || (r >= 'a' && r <= 'f')
}
