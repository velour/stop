package ast

import (
	"math/big"
	"strconv"
	"unicode/utf8"

	"bitbucket.org/eaburns/stop/token"
)

var (
	// TODO(eaburns): Define tokens.nTokens and change
	// map[token.Token]Whatever to [nTokens]Whatever.

	// Binary op precedence for precedence climbing algorithm.
	// http://www.engr.mun.ca/~theo/Misc/exp_parsing.htm
	precedence = map[token.Token]int{
		token.OrOr:           1,
		token.AndAnd:         2,
		token.EqualEqual:     3,
		token.BangEqual:      3,
		token.Less:           3,
		token.LessEqual:      3,
		token.Greater:        3,
		token.GreaterEqual:   3,
		token.Plus:           4,
		token.Minus:          4,
		token.Or:             4,
		token.Carrot:         4,
		token.Star:           5,
		token.Divide:         5,
		token.Percent:        5,
		token.LessLess:       5,
		token.GreaterGreater: 5,
		token.And:            5,
		token.AndCarrot:      5,
	}

	// Set of unary operators.
	unary = map[token.Token]bool{
		token.Plus:      true,
		token.Minus:     true,
		token.Bang:      true,
		token.Carrot:    true,
		token.Star:      true,
		token.And:       true,
		token.LessMinus: true,
	}

	// UnEsc maps escape characters to their unescaped runes.
	unEsc = [...]rune{
		'a':  '\a',
		'b':  '\b',
		'f':  '\f',
		'n':  '\n',
		'r':  '\r',
		't':  '\t',
		'v':  '\v',
		'\'': '\'',
		'"':  '"',
	}
)

// Parse returns the root of an abstract syntax tree for the Go language
// or an error if one is encountered.
//
// TODO(eaburns): This is currently just for testing since it doesn't
// parse the top-level production.
func Parse(p *Parser) (root Node, err error) {
	defer func() {
		r := recover()
		if r == nil {
			return
		}
		switch e := r.(type) {
		case *SyntaxError:
			err = e
		case *MalformedLiteral:
			err = e
		default:
			panic(r)
		}

	}()
	root = parseExpression(p)
	return
}

func parseExpression(p *Parser) Expression {
	return parseBinaryExpr(p, 1)
}

func parseBinaryExpr(p *Parser, prec int) Expression {
	left := parseUnaryExpr(p)
	for {
		pr, ok := precedence[p.tok]
		if !ok || pr < prec {
			return left
		}
		op, opLoc := p.tok, p.lex.Start
		p.next()
		right := parseBinaryExpr(p, pr+1)
		left = &BinaryOp{
			Op:    op,
			opLoc: opLoc,
			Left:  left,
			Right: right,
		}
	}
}

func parseUnaryExpr(p *Parser) Expression {
	if unary[p.tok] {
		op, opLoc := p.tok, p.lex.Start
		p.next()
		operand := parseUnaryExpr(p)
		return &UnaryOp{
			Op:      op,
			opLoc:   opLoc,
			Operand: operand,
		}
	}
	return parseOperand(p)
}

func parseOperand(p *Parser) Expression {
	switch p.tok {
	case token.Identifier:
		return parseOperandName(p)

	case token.IntegerLiteral:
		return parseIntegerLiteral(p)

	case token.FloatLiteral:
		return parseFloatLiteral(p)

	case token.ImaginaryLiteral:
		return parseImaginaryLiteral(p)

	case token.RuneLiteral:
		return parseRuneLiteral(p)

	case token.StringLiteral:
		return parseStringLiteral(p)

	// TODO(eaburns): Composite literal
	// TODO(eaburns): Function literal
	// TODO(eaburns): MethodExpr—needs to be merged with the OpenParen case

	case token.OpenParen:
		p.next()
		e := parseExpression(p)
		p.expect(token.CloseParen)
		p.next()
		return e

	default:
		panic(p.err("operand"))
	}
}

func parseIntegerLiteral(p *Parser) Expression {
	l := &IntegerLiteral{Value: new(big.Int), span: p.span()}
	if _, ok := l.Value.SetString(p.lex.Text(), 0); ok {
		p.next()
		return l
	}
	panic(&MalformedLiteral{
		Type:  "integer literal",
		Text:  p.text(),
		Start: p.lex.Start,
		End:   p.lex.End,
	})
}

func parseFloatLiteral(p *Parser) Expression {
	l := &FloatLiteral{Value: new(big.Rat), span: p.span()}
	if _, ok := l.Value.SetString(p.lex.Text()); ok {
		p.next()
		return l
	}
	panic(&MalformedLiteral{
		Type:  "float literal",
		Text:  p.text(),
		Start: p.lex.Start,
		End:   p.lex.End,
	})
}

func parseImaginaryLiteral(p *Parser) Expression {
	text := p.lex.Text()
	if len(text) < 1 || text[len(text)-1] != 'i' {
		panic("bad imaginary literal: " + text)
	}
	text = text[:len(text)-1]
	l := &ImaginaryLiteral{Value: new(big.Rat), span: p.span()}
	if _, ok := l.Value.SetString(text); ok {
		p.next()
		return l
	}
	panic(&MalformedLiteral{
		Type:  "imaginary literal",
		Text:  p.text(),
		Start: p.lex.Start,
		End:   p.lex.End,
	})
}

func parseStringLiteral(p *Parser) Expression {
	text := p.lex.Text()
	if len(text) < 2 {
		panic("bad string literal: " + text)
	}
	raw := text[0] == '`'
	text = text[1 : len(text)-1]
	esc := false
	runes := make([]rune, 0, len(text))
	for _, r := range text {
		if raw {
			if r != '\r' {
				runes = append(runes, r)
			}
			continue
		}
		if !esc && r == '\\' {
			esc = true
			continue
		} else if esc {
			esc = false
			r = unEsc[r]
			if r == 0 {
				panic("bad escape sequence: " + p.lex.Text())
			}
		}
		runes = append(runes, r)
	}
	l := &StringLiteral{Value: string(runes), span: p.span()}
	p.next()
	return l
}

func parseRuneLiteral(p *Parser) Expression {
	l := &RuneLiteral{span: p.span()}

	text := p.lex.Text()
	if len(text) < 3 {
		panic("bad rune literal: " + p.lex.Text())
	}

	if text[1] != '\\' {
		l.Value, _ = utf8.DecodeRuneInString(text[1:])
		p.next()
		return l
	}

	if len(text) < 4 {
		panic("bad rune literal: " + p.lex.Text())
	}
	kind := text[2]
	text = text[3 : len(text)-1]
	switch kind {
	case 'U':
		if len(text) < 8 {
			panic("bad rune literal: " + p.lex.Text())
		}
		fallthrough
	case 'u':
		if len(text) < 4 {
			panic("bad rune literal: " + p.lex.Text())
		}
		fallthrough
	case 'x':
		if len(text) < 2 {
			panic("bad rune literal: " + p.lex.Text())
		}
		v, err := strconv.ParseUint(text, 16, 32)
		// TODO(eaburns): Figure out what surrogate halves are and
		// check for them here (they are malformed).
		if err != nil || v > 0x10FFFF {
			goto malformed
		}
		l.Value = rune(v)

	case 'a', 'b', 'f', 'n', 'r', 't', 'v', '\\', '\'':
		l.Value = unEsc[kind]
		if l.Value == 0 {
			panic("bad escape sequence: " + p.lex.Text())
		}

	default:
		if len(text) != 3 {
			panic("bad rune literal: " + p.lex.Text())
		}
		v, err := strconv.ParseUint(text, 8, 32)
		if err != nil {
			goto malformed
		}
		l.Value = rune(v)
	}
	p.next()
	return l

malformed:
	panic(&MalformedLiteral{
		Type:  "rune literal",
		Text:  p.text(),
		Start: p.lex.Start,
		End:   p.lex.End,
	})
}

func parseOperandName(p *Parser) Expression {
	name := &Identifier{Name: p.lex.Text(), span: p.span()}
	p.next()
	if p.tok == token.Dot {
		p.next()
		p.expect(token.Identifier)
		name.Package = name.Name
		name.Name = p.lex.Text()
		name.span.end = p.lex.End
		p.next()
	}
	return name
}
