package token

import (
	"unicode"
	"unicode/utf8"
)

// A Lexer scans and returns Go tokens from an input stream.
type Lexer struct {
	// Src is the source text being scanned.
	src string

	// N is the next un-scanned byte in src and w is the width of the
	// most-recently returned rune.  If n >= len(src) && w==0
	// then an end of file, -1, rune has been returned by rune()
	n, w int

	// PrevLineStart is the rune offset of the start of the previous
	// line.  It is needed to reset the LineStart field of the End
	// location when replace is called on a newline rune.
	prevLineStart int

	// Prev is the most-recent, non-comment, non-whitespace
	// token returned by Next.
	prev Token

	// Start and End are the locations of the start and end of the
	// last token scanned by Next.  They are left unmodified
	// when Next returns an inserted semicolon.
	Start, End Location
}

// NewLexer returns a new Lexer that reads tokens from an input file.
// If the underlying type of the reader is an os.File, then all Locations
// produced by the lexer use the file name as their path.
func NewLexer(path string, src string) *Lexer {
	l := &Lexer{
		src:           src,
		prevLineStart: -1,
		Start:         Location{Path: path, Line: 1, Rune: 1, LineStart: 1},
		End:           Location{Path: path, Line: 1, Rune: 1, LineStart: 1},
	}
	return l
}

// Text returns the text of the last token scanned by Next.  Inserted
// semicolons do not update the string retuned by Text.  If Next
// returns an inserted semicolon, then Text will return the text of the
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
	l.n, l.w = 0, 0
	l.Start = l.End
	t := l.scan()

	nl := (t == Comment || t == Whitespace) && l.End.Line > l.Start.Line
	if (nl || t == EOF) &&
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
		t = Semicolon
	}
	if t != Comment && t != Whitespace {
		l.prev = t
	}
	if t == Error {
		// Advance to the most-recent rune: the unexpected one.
		l.src = l.src[l.n-1:]
		l.n = 1
	}
	return t
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

// Rune reads and returns the next rune from the input stream and
// updates the locations and text of the current token.  If the end of
// the input is reached -1 is returned.
func (l *Lexer) rune() (r rune) {
	if l.n >= len(l.src) {
		l.w = 0
		return -1
	}
	r, l.w = utf8.DecodeRuneInString(l.src[l.n:])
	l.n += l.w
	l.End.Rune++
	if r == '\n' {
		l.End.Line++
		l.prevLineStart = l.End.LineStart
		l.End.LineStart = l.End.Rune
	}
	return r
}

// Replace replaces the most-recently-read rune into the input
// stream.  If no runes were read since the last call to Next then
// in panicks; nothing can be replaced.
func (l *Lexer) replace() {
	if l.n == 0 {
		panic("nothing to replace")
	}
	if l.n >= len(l.src) && l.w == 0 {
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
