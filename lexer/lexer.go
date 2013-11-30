package lexer

import (
	"bufio"
	"io"
	"os"
	"unicode"

	"gogo/loc"
)

// Token is a the atomic unit of the vocabulary of a Go program.
type Token struct {
	Type TokenType
	Text string
	Span loc.Span
	// Keyword holds the keyword if Type is TokenKeyword.
	Keyword Keyword
	// Operator holds the operator if Type is TokenOperator.
	Operator Operator
}

// String returns a human-readable string representation of the token.
func (t Token) String() string {
	str := "Token{Type:" + t.Type.String() + ", Text:`" + t.Text + "`, Span:" + t.Span.String()
	switch t.Type {
	case TokenKeyword:
		str += ", Keyword:" + t.Keyword.String()
	case TokenOperator:
		str += ", Operator:" + t.Operator.String()
	}
	return str + "}"
}

// A TokenType identifies the type of a token in the input file.
type TokenType int

const (
	TokenIdentifier TokenType = iota
	TokenKeyword
	TokenOperator
	TokenIntegerLiteral
	TokenFloatLiteral
	TokenImaginaryLiteral
	TokenRuneLiteral
	TokenStringLiteral
	TokenNewline
	TokenWhitespace
	TokenEOF
	TokenError
)

var tokenTypeNames = [...]string{
	TokenIdentifier:       "TokenIdentifier",
	TokenKeyword:          "TokenKeyword",
	TokenOperator:         "TokenOperator",
	TokenIntegerLiteral:   "TokenIntegerLiteral",
	TokenFloatLiteral:     "TokenFloatLiteral",
	TokenImaginaryLiteral: "TokenImaginaryLiteral",
	TokenRuneLiteral:      "TokenRuneLiteral",
	TokenStringLiteral:    "TokenStringLiteral",
	TokenNewline:          "TokenNewline",
	TokenWhitespace:       "TokenWhitespace",
	TokenEOF:              "TokenEOF",
	TokenError:            "TokenError",
}

// String returns the string representation of the token type.
func (t TokenType) String() string {
	return tokenTypeNames[t]
}

// Keyword is a reserved word that may not be used as an identifier.
type Keyword int

const (
	KeywordNone Keyword = iota
	KeywordBreak
	KeywordDefault
	KeywordFunc
	KeywordInterface
	KeywordSelect
	KeywordCase
	KeywordDefer
	KeywordGo
	KeywordMap
	KeywordStruct
	KeywordChan
	KeywordElse
	KeywordGoto
	KeywordPackage
	KeywordSwitch
	KeywordConst
	KeywordFallthrough
	KeywordIf
	KeywordRange
	KeywordType
	KeywordContinue
	KeywordFor
	KeywordImport
	KeywordReturn
	KeywordVar
)

var keywords = map[string]Keyword{
	"none":        KeywordNone,
	"break":       KeywordBreak,
	"default":     KeywordDefault,
	"func":        KeywordFunc,
	"interface":   KeywordInterface,
	"select":      KeywordSelect,
	"case":        KeywordCase,
	"defer":       KeywordDefer,
	"go":          KeywordGo,
	"map":         KeywordMap,
	"struct":      KeywordStruct,
	"chan":        KeywordChan,
	"else":        KeywordElse,
	"goto":        KeywordGoto,
	"package":     KeywordPackage,
	"switch":      KeywordSwitch,
	"const":       KeywordConst,
	"fallthrough": KeywordFallthrough,
	"if":          KeywordIf,
	"range":       KeywordRange,
	"type":        KeywordType,
	"continue":    KeywordContinue,
	"for":         KeywordFor,
	"import":      KeywordImport,
	"return":      KeywordReturn,
	"var":         KeywordVar,
}

var keywordText = func() []string {
	text := make([]string, len(keywords))
	for t, k := range keywords {
		text[k] = t
	}
	return text
}()

var keywordNames = [...]string{
	KeywordNone:        "KeywordNone",
	KeywordBreak:       "KeywordBreak",
	KeywordDefault:     "KeywordDefault",
	KeywordFunc:        "KeywordFunc",
	KeywordInterface:   "KeywordInterface",
	KeywordSelect:      "KeywordSelect",
	KeywordCase:        "KeywordCase",
	KeywordDefer:       "KeywordDefer",
	KeywordGo:          "KeywordGo",
	KeywordMap:         "KeywordMap",
	KeywordStruct:      "KeywordStruct",
	KeywordChan:        "KeywordChan",
	KeywordElse:        "KeywordElse",
	KeywordGoto:        "KeywordGoto",
	KeywordPackage:     "KeywordPackage",
	KeywordSwitch:      "KeywordSwitch",
	KeywordConst:       "KeywordConst",
	KeywordFallthrough: "KeywordFallthrough",
	KeywordIf:          "KeywordIf",
	KeywordRange:       "KeywordRange",
	KeywordType:        "KeywordType",
	KeywordContinue:    "KeywordContinue",
	KeywordFor:         "KeywordFor",
	KeywordImport:      "KeywordImport",
	KeywordReturn:      "KeywordReturn",
	KeywordVar:         "KeywordVar",
}

// String returns the string representation of a Keyword constant.
// For example, KeywordDefault.String() will return "KeywordDefault."
func (k Keyword) String() string {
	return keywordNames[k]
}

// Text returns the textual representation of the Keyword.
// For example, KeywordDefault.Text() will return "default."
func (k Keyword) Text() string {
	return keywordText[k]
}

// An Operator is an operator.
type Operator int

const (
	OpNone Operator = iota
	OpPlus
	OpAnd
	OpPlusEqual
	OpAndEqual
	OpAndAnd
	OpEqualEqual
	OpBangEqual
	OpOpenParen
	OpCloseParen
	OpMinus
	OpOr
	OpMinusEqual
	OpOrEqual
	OpOrOr
	OpLess
	OpLessEqual
	OpOpenBracket
	OpCloseBracket
	OpStar
	OpCarrot
	OpStarEqual
	OpCarrotEqual
	OpLessMinus
	OpGreater
	OpGreaterEqual
	OpOpenBrace
	OpCloseBrace
	OpDivide
	OpLessLess
	OpDivideEqual
	OpLessLessEqual
	OpPlusPlus
	OpEqual
	OpColonEqual
	OpComma
	OpSemicolon
	OpPercent
	OpGreaterGreater
	OpPercentEqual
	OpGreaterGreaterEqual
	OpMinusMinus
	OpBang
	OpDotDotDot
	OpDot
	OpColon
	OpAndCarrot
	OpAndCarrotEqual
)

var operators = map[string]Operator{
	"none": OpNone,
	"+":    OpPlus,
	"&":    OpAnd,
	"+=":   OpPlusEqual,
	"&=":   OpAndEqual,
	"&&":   OpAndAnd,
	"==":   OpEqualEqual,
	"!=":   OpBangEqual,
	"(":    OpOpenParen,
	")":    OpCloseParen,
	"-":    OpMinus,
	"|":    OpOr,
	"-=":   OpMinusEqual,
	"|=":   OpOrEqual,
	"||":   OpOrOr,
	"<":    OpLess,
	"<=":   OpLessEqual,
	"[":    OpOpenBracket,
	"]":    OpCloseBracket,
	"*":    OpStar,
	"^":    OpCarrot,
	"*=":   OpStarEqual,
	"^=":   OpCarrotEqual,
	"<-":   OpLessMinus,
	">":    OpGreater,
	">=":   OpGreaterEqual,
	"{":    OpOpenBrace,
	"}":    OpCloseBrace,
	"/":    OpDivide,
	"<<":   OpLessLess,
	"/=":   OpDivideEqual,
	"<<=":  OpLessLessEqual,
	"++":   OpPlusPlus,
	"=":    OpEqual,
	":=":   OpColonEqual,
	",":    OpComma,
	";":    OpSemicolon,
	"%":    OpPercent,
	">>":   OpGreaterGreater,
	"%=":   OpPercentEqual,
	">>=":  OpGreaterGreaterEqual,
	"--":   OpMinusMinus,
	"!":    OpBang,
	"...":  OpDotDotDot,
	".":    OpDot,
	":":    OpColon,
	"&^":   OpAndCarrot,
	"&^=":  OpAndCarrotEqual,
}

var operatorText = func() []string {
	text := make([]string, len(operators))
	for t, o := range operators {
		text[o] = t
	}
	return text
}()

var operatorNames = [...]string{
	OpNone:                "OpNone",
	OpPlus:                "OpPlus",
	OpAnd:                 "OpAnd",
	OpPlusEqual:           "OpPlusEqual",
	OpAndEqual:            "OpAndEqual",
	OpAndAnd:              "OpAndAnd",
	OpEqualEqual:          "OpEqualEqual",
	OpBangEqual:           "OpBangEqual",
	OpOpenParen:           "OpOpenParen",
	OpCloseParen:          "OpCloseParen",
	OpMinus:               "OpMinus",
	OpOr:                  "OpOr",
	OpMinusEqual:          "OpMinusEqual",
	OpOrEqual:             "OpOrEqual",
	OpOrOr:                "OpOrOr",
	OpLess:                "OpLess",
	OpLessEqual:           "OpLessEqual",
	OpOpenBracket:         "OpOpenBracket",
	OpCloseBracket:        "OpCloseBracket",
	OpStar:                "OpStar",
	OpCarrot:              "OpCarrot",
	OpStarEqual:           "OpStarEqual",
	OpCarrotEqual:         "OpCarrotEqual",
	OpLessMinus:           "OpLessMinus",
	OpGreater:             "OpGreater",
	OpGreaterEqual:        "OpGreaterEqual",
	OpOpenBrace:           "OpOpenBrace",
	OpCloseBrace:          "OpCloseBrace",
	OpDivide:              "OpDivide",
	OpLessLess:            "OpLessLess",
	OpDivideEqual:         "OpDivideEqual",
	OpLessLessEqual:       "OpLessLessEqual",
	OpPlusPlus:            "OpPlusPlus",
	OpEqual:               "OpEqual",
	OpColonEqual:          "OpColonEqual",
	OpComma:               "OpComma",
	OpSemicolon:           "OpSemicolon",
	OpPercent:             "OpPercent",
	OpGreaterGreater:      "OpGreaterGreater",
	OpPercentEqual:        "OpPercentEqual",
	OpGreaterGreaterEqual: "OpGreaterGreaterEqual",
	OpMinusMinus:          "OpMinusMinus",
	OpBang:                "OpBang",
	OpDotDotDot:           "OpDotDotDot",
	OpDot:                 "OpDot",
	OpColon:               "OpColon",
	OpAndCarrot:           "OpAndCarrot",
	OpAndCarrotEqual:      "OpAndCarrotEqual",
}

// String returns the string representation of the operator constant.
// For example, OpColon.String() will return "OpColon."
func (o Operator) String() string {
	return operatorNames[o]
}

// Text returns the textual representation of the operator.
// For example, OpColon.Text() will return ":."
func (o Operator) Text() string {
	return operatorText[o]
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
		span:          loc.Span{loc.Zero(path), loc.Zero(path)},
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
		Type: TokenError,
		Span: l.span,
	}
}

func (l *Lexer) nextLiteralToken() *Token {
	r := l.rune()
	switch {
	case r < 0:
		return l.token(TokenEOF)
	case r == '\n':
		return l.token(TokenNewline)
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
	if tok.Type == TokenEOF && (l.prev == nil || l.prev.Type != TokenNewline) {
		l.next = tok
		tok = &Token{
			Type: TokenNewline,
			Span: tok.Span,
		}
	}
	if tok.Type == TokenNewline && l.prev != nil && needsSemicolon(l.prev) {
		l.next = tok
		tok = &Token{
			Type:     TokenOperator,
			Span:     loc.Span{l.next.Span[0], l.next.Span[0]},
			Operator: OpSemicolon,
		}
	}
	l.prev = tok
	return tok
}

func needsSemicolon(tok *Token) bool {
	return tok.Type == TokenIdentifier ||
		tok.Type == TokenIntegerLiteral ||
		tok.Type == TokenFloatLiteral ||
		tok.Type == TokenImaginaryLiteral ||
		tok.Type == TokenRuneLiteral ||
		tok.Type == TokenStringLiteral ||
		(tok.Type == TokenKeyword &&
			(tok.Keyword == KeywordBreak ||
				tok.Keyword == KeywordContinue ||
				tok.Keyword == KeywordFallthrough ||
				tok.Keyword == KeywordReturn)) ||
		(tok.Type == TokenOperator &&
			(tok.Operator == OpPlusPlus ||
				tok.Operator == OpMinusMinus ||
				tok.Operator == OpCloseParen ||
				tok.Operator == OpCloseBracket ||
				tok.Operator == OpCloseBrace))
}

func operator(l *Lexer) *Token {
	tok := l.token(TokenOperator)
	oper, ok := operators[tok.Text]
	if !ok {
		panic("bad operator: \"" + tok.Text + "\"")
	}
	tok.Operator = oper
	return tok
}

func identifier(l *Lexer) *Token {
	r := l.rune()
	for isIdent(r) {
		r = l.rune()
	}
	l.replace()
	tok := l.token(TokenIdentifier)
	if keyword, ok := keywords[tok.Text]; ok {
		tok.Type = TokenKeyword
		tok.Keyword = keyword
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
	return l.token(TokenWhitespace)
}

func isWhitespace(r rune) bool {
	return r == ' ' || r == '\t' || r == '\v' || r == '\r'
}

func comment(l *Lexer, closing []rune) *Token {
	i := 0
	typ := TokenWhitespace
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
			typ = TokenNewline
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
			return l.token(TokenImaginaryLiteral)
		case !isDecimalDigit(r):
			l.replace()
			return l.token(TokenIntegerLiteral)
		}
		r = l.rune()
	}
	panic("unreachable")
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
		return l.token(TokenImaginaryLiteral)
	}
	l.replace()
	return l.token(TokenFloatLiteral)
}

func fraction(l *Lexer) *Token {
	r := l.rune()
	for isDecimalDigit(r) {
		r = l.rune()
	}
	switch {
	case r == 'i':
		return l.token(TokenImaginaryLiteral)
	case r == 'e' || r == 'E':
		return mantissa(l)
	}
	l.replace()
	return l.token(TokenFloatLiteral)
}

func hex(l *Lexer) *Token {
	r := l.rune()
	for isHexDigit(r) {
		r = l.rune()
	}
	l.replace()
	return l.token(TokenIntegerLiteral)
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
