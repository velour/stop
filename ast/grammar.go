package ast

import (
	"bitbucket.org/eaburns/stop/token"
)

var (
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
)

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
		l := &IntegerLiteral{StringValue: p.text(), span: p.span()}
		p.next()
		return l

	case token.FloatLiteral:
		l := &FloatLiteral{StringValue: p.text(), span: p.span()}
		p.next()
		return l

	case token.ImaginaryLiteral:
		l := &ImaginaryLiteral{StringValue: p.text(), span: p.span()}
		p.next()
		return l

	case token.RuneLiteral:
		l := &RuneLiteral{StringValue: p.text(), span: p.span()}
		p.next()
		return l

	case token.StringLiteral:
		l := &StringLiteral{Value: p.text(), span: p.span()}
		p.next()
		return l

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
