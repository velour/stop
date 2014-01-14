// Package token provides tokens and a lexer for the Go language.
//
// Quirks:
//
// Mal-formed octal literals that begin with 0 followed by any
// number of decmial digits are returned as valid integer literals.
//
// Unicode code points are not validated, so characters with
// invalid code points (such as '\U00110000' and '\uDFFF') are
// returned as valid character literals.
package token

// A Token identifies the type of a token in the input file.
type Token int

// The set of constants defining the types of tokens.
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
func (tok Token) String() string {
	return tokenNames[tok]
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
