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
	"strconv"
	"unicode"
	"unicode/utf8"
)

// A Location is the address of a particular rune in an input stream.
type Location struct {
	// Path is the path to the file for this location, or "" if there is no file.
	Path string
	// Rune is the rune number of this location within the input stream.
	Rune int
	// Line is the line number of this location.
	Line int
	// LineStart is the rune number of the first rune on the line of this location.
	LineStart int
}

// RuneOnLine returns the rune offset into the line of a location.
func (l Location) RuneOnLine() int {
	return l.Rune - l.LineStart
}

// String returns the string representation of a location as an Acme address.
func (l Location) String() string {
	return l.Path + ":" + strconv.Itoa(l.Line) + "-+#" + strconv.Itoa(l.RuneOnLine())
}

// A Token identifies the type of a token in the input file.
type Token int

// The set of constants defining the types of tokens.
const (
	EOF   Token = -1
	Error Token = iota - 1 // 0, since EOF came first
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

var tokenTypeNames = map[Token]string{
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
func (tt Token) String() string {
	return tokenTypeNames[tt]
}

var keywords = map[string]Token{
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

var operators = map[string]Token{
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

// A Lexer scans and returns Go tokens from an input stream.
type Lexer struct {
	src           string
	n, w          int
	eof           bool
	prevLineStart int
	// Start and End are the locations of the start and end of the
	// last token scanned by Next.  They are left unmodified
	// when Next returns an inserted semicolon.
	Start, End Location

	// The most-recent, non-comment, non-whitespace token.
	prev Token
}

// New returns a new Lexer that reads tokens from an input file.
// If the underlying type of the reader is an os.File, then all Locations
// produced by the lexer use the file name as their path.
func New(path string, src string) *Lexer {
	l := &Lexer{
		src:           src,
		prevLineStart: -1,
		Start:         Location{Path: path, Line: 1, Rune: 1, LineStart: 1},
		End:           Location{Path: path, Line: 1, Rune: 1, LineStart: 1},
	}
	return l
}

// Text returns the text of the last token scanned by Next.  Inserted
// semicolons do not update the string retuned by Text, so if Next
// returns an inserted semicolon then Text will return the text of the
// token preceeding the semicolon.
func (l *Lexer) Text() string {
	return l.src[:l.n]
}

// Next returns the next token, inserting semicolons where required.
// Once an Error or EOF token is returned, all subsequent tokens will
// be of type EOF.
func (l *Lexer) Next() Token {
	// Saved, to be reset if a semicolon is inserted.
	src, n, start, end := l.src, l.n, l.Start, l.End

	l.src = l.src[l.n:]
	l.n = 0
	l.w = 0
	l.Start = l.End
	tok := l.scan()

	if tok == Error {
		// The most-recently read rune was unexpected.
		l.src = l.src[l.n-1:]
		l.n = 1
	}

	newline := (tok == Comment || tok == Whitespace) && l.End.Line > l.Start.Line
	if (newline || tok == EOF) &&
		(l.prev == Identifier ||
			l.prev == IntegerLiteral ||
			l.prev == FloatLiteral ||
			l.prev == ImaginaryLiteral ||
			l.prev == RuneLiteral ||
			l.prev == StringLiteral ||
			l.prev == Break ||
			l.prev == Continue ||
			l.prev == Fallthrough ||
			l.prev == Return ||
			l.prev == PlusPlus ||
			l.prev == MinusMinus ||
			l.prev == CloseParen ||
			l.prev == CloseBracket ||
			l.prev == CloseBrace) {
		l.src, l.n, l.Start, l.End = src, n, start, end
		tok = Semicolon
	}
	if tok != Comment && tok != Whitespace {
		l.prev = tok
		return tok
	}
	return tok
}

// Scan scans the next token and returns it.
func (l *Lexer) scan() Token {
	r := l.rune()
	switch {
	case r < 0:
		return EOF
	case isWhitespace(r):
		return whitespace(l)
	case isLetter(r):
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
				return Error
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
	return Error
}

// Rune reads and returns the next rune from the input stream and updates the
// span and text of the current token.  If an error is encountered then it panicks.
// If the end of the input is reached -1 is returned.
func (l *Lexer) rune() rune {
	if l.n >= len(l.src) {
		l.eof = true
		return -1
	}
	r, w := utf8.DecodeRuneInString(l.src[l.n:])
	l.n += w
	l.w = w
	l.End.Rune++
	if r == '\n' {
		l.End.Line++
		l.prevLineStart = l.End.LineStart
		l.End.LineStart = l.End.Rune
	}
	return r
}

// Replace replaces the most-recently-read rune into the input stream.  If an
// error is encountered then it panicks.
func (l *Lexer) replace() {
	if l.n == 0 {
		panic("nothing to replace")
	}
	if l.eof {
		return
	}
	l.n -= l.w
	l.w = 0

	l.End.Rune--
	if l.End.Rune < l.End.LineStart {
		l.End.Line--
		l.End.LineStart = l.prevLineStart
		l.prevLineStart = -1
	}
}

func operator(l *Lexer) Token {
	text := l.src[:l.n]
	oper, ok := operators[text]
	if !ok {
		panic("bad operator: \"" + string(text) + "\"")
	}
	return oper
}

func identifier(l *Lexer) Token {
	r := l.rune()
	for isIdent(r) {
		r = l.rune()
	}
	l.replace()
	text := l.src[:l.n]
	if keyword, ok := keywords[text]; ok {
		return keyword
	}
	return Identifier
}

func whitespace(l *Lexer) Token {
	r := l.rune()
	for isWhitespace(r) {
		r = l.rune()
	}
	l.replace()
	return Whitespace
}

func comment(l *Lexer, closing []rune) Token {
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
	return Comment
}

func number(r0 rune, l *Lexer) Token {
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
			return ImaginaryLiteral
		case !isDecimalDigit(r):
			l.replace()
			return IntegerLiteral
		}
		r = l.rune()
	}
}

func mantissa(l *Lexer) Token {
	r := l.rune()
	if r == '+' || r == '-' {
		r = l.rune()
	}
	for isDecimalDigit(r) {
		r = l.rune()
	}
	if r == 'i' {
		return ImaginaryLiteral
	}
	l.replace()
	return FloatLiteral
}

func fraction(l *Lexer) Token {
	r := l.rune()
	for isDecimalDigit(r) {
		r = l.rune()
	}
	switch {
	case r == 'i':
		return ImaginaryLiteral
	case r == 'e' || r == 'E':
		return mantissa(l)
	}
	l.replace()
	return FloatLiteral
}

func hex(l *Lexer) Token {
	r := l.rune()
	for isHexDigit(r) {
		r = l.rune()
	}
	l.replace()
	return IntegerLiteral
}

func runeLiteral(l *Lexer) Token {
	if l.rune() == '\\' {
		if !unicodeValue(l, '\'') {
			return Error
		}
	}
	if r := l.rune(); r != '\'' {
		return Error
	}
	return RuneLiteral
}

func interpertedStringLiteral(l *Lexer) Token {
	for {
		r := l.rune()
		switch {
		case r == '"':
			return StringLiteral
		case r == '\n':
			return Error
		case r == '\\':
			if !unicodeValue(l, '"') {
				return Error
			}
		case r < 0:
			return Error
		}
	}
}

func rawStringLiteral(l *Lexer) Token {
	for {
		r := l.rune()
		switch {
		case r == '`':
			return StringLiteral
		case r < 0:
			return Error
		}
	}
}

// Parses a unicode value (assuming that the leading '\' has
// already been consumed), and returns true on success or
// false if the last rune read was unexpected.
func unicodeValue(l *Lexer, quote rune) bool {
	r := l.rune()
	switch {
	case r == 'U':
		if s := l.rune(); !isHexDigit(s) {
			return false
		}
		if s := l.rune(); !isHexDigit(s) {
			return false
		}
		if s := l.rune(); !isHexDigit(s) {
			return false
		}
		if s := l.rune(); !isHexDigit(s) {
			return false
		}
		fallthrough

	case r == 'u':
		if s := l.rune(); !isHexDigit(s) {
			return false
		}
		if s := l.rune(); !isHexDigit(s) {
			return false
		}
		fallthrough

	case r == 'x':
		if s := l.rune(); !isHexDigit(s) {
			return false
		}
		if s := l.rune(); !isHexDigit(s) {
			return false
		}

	case isOctalDigit(r):
		if s := l.rune(); !isOctalDigit(s) {
			return false
		}
		if s := l.rune(); !isOctalDigit(s) {
			return false
		}

	case r != 'a' && r != 'b' && r != 'f' && r != 'n' && r != 'r' && r != 't' && r != 'v' && r != '\\' && r != quote:
		return false
	}
	return true
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

func isWhitespace(r rune) bool {
	return r == ' ' || r == '\t' || r == '\v' || r == '\r' || r == '\n'
}

func isLetter(r rune) bool {
	return unicode.IsLetter(r) || r == '_'
}

func isIdent(r rune) bool {
	return isLetter(r) || unicode.IsDigit(r)
}
