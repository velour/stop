// Package token provides tokens and a lexer for the Go language.
package token

// A Token identifies the type of a token in the input file.
type Token int

// Constants defining the tokens.
const (
	Error Token = iota
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

	EOF Token = -1
)

var tokenNames = map[Token]string{
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
	Break:               "break",
	Default:             "default",
	Func:                "func",
	Interface:           "interface",
	Select:              "select",
	Case:                "case",
	Defer:               "defer",
	Go:                  "go",
	Map:                 "map",
	Struct:              "struct",
	Chan:                "chan",
	Else:                "else",
	Goto:                "goto",
	Package:             "package",
	Switch:              "switch",
	Const:               "const",
	Fallthrough:         "fallthrough",
	If:                  "if",
	Range:               "range",
	Type:                "type",
	Continue:            "continue",
	For:                 "for",
	Import:              "import",
	Return:              "return",
	Var:                 "var",
	Plus:                "+",
	And:                 "&",
	PlusEqual:           "+=",
	AndEqual:            "&=",
	AndAnd:              "&&",
	EqualEqual:          "==",
	BangEqual:           "!=",
	OpenParen:           "(",
	CloseParen:          ")",
	Minus:               "-",
	Or:                  "|",
	MinusEqual:          "-=",
	OrEqual:             "|=",
	OrOr:                "||",
	Less:                "<",
	LessEqual:           "<=",
	OpenBracket:         "[",
	CloseBracket:        "]",
	Star:                "*",
	Carrot:              "^",
	StarEqual:           "*=",
	CarrotEqual:         "^=",
	LessMinus:           "<-",
	Greater:             ">",
	GreaterEqual:        ">=",
	OpenBrace:           "{",
	CloseBrace:          "}",
	Divide:              "/",
	LessLess:            "<<",
	DivideEqual:         "/=",
	LessLessEqual:       "<<=",
	PlusPlus:            "++",
	Equal:               "=",
	ColonEqual:          ":=",
	Comma:               ",",
	Semicolon:           ";",
	Percent:             "%",
	GreaterGreater:      ">>",
	PercentEqual:        "%=",
	GreaterGreaterEqual: ">>=",
	MinusMinus:          "--",
	Bang:                "!",
	DotDotDot:           "...",
	Dot:                 ".",
	Colon:               ":",
	AndCarrot:           "&^",
	AndCarrotEqual:      "&^=",
}

// String returns the string represenation of a token.
func (tok Token) String() string { return tokenNames[tok] }

// PrettyPrint implements the pretty.PrettyPrinter interface.
func (tok Token) PrettyPrint() string { return tok.String() }

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
