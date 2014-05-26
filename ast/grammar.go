package ast

import (
	"math/big"
	"strconv"
	"strings"

	"github.com/velour/stop/token"
)

// Parse returns the root of an abstract syntax tree for the Go language
// or an error if one is encountered.
//
// BUG(eaburns): This is currently just for testing since it doesn't
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
	return parseSourceFile(p), nil
}

func parseSourceFile(p *Parser) Node {
	p.expect(token.Package)
	s := &SourceFile{comments: p.comments(), startLoc: p.lex.Start}
	p.next()
	s.PackageName = *parseIdentifier(p)
	p.expect(token.Semicolon)
	p.next()

	for p.tok == token.Import {
		s.Imports = append(s.Imports, *parseImportDecl(p))
		if p.tok == token.EOF {
			break
		}
		p.expect(token.Semicolon)
		p.next()
	}
	for p.tok != token.EOF {
		decls := parseTopLevelDecl(p)
		s.Declarations = append(s.Declarations, decls...)
		if p.tok == token.EOF {
			break
		}
		p.expect(token.Semicolon)
		p.next()
	}
	s.endLoc = p.lex.Start
	return s
}

func parseStatement(p *Parser) Statement {
	switch p.tok {
	case token.Type, token.Const, token.Var:
		return &DeclarationStmt{
			comments:     p.comments(),
			Declarations: parseDeclarations(p),
		}

	case token.Go:
		return parseGo(p)

	case token.Return:
		return parseReturn(p)

	case token.Break:
		return parseBreak(p)

	case token.Continue:
		return parseContinue(p)

	case token.Goto:
		return parseGoto(p)

	case token.Fallthrough:
		c, s, e := p.comments(), p.lex.Start, p.lex.End
		p.next()
		return &FallthroughStmt{
			comments: c,
			startLoc: s,
			endLoc:   e,
		}

	case token.OpenBrace:
		return parseBlock(p)

	case token.If:
		return parseIf(p)

	case token.Switch:
		return parseSwitch(p)

	case token.Select:
		return parseSelect(p)

	case token.For:
		return parseFor(p)

	case token.Defer:
		return parseDefer(p)
	}
	return parseSimpleStmt(p, labelOK)
}

func parseSelect(p *Parser) Statement {
	p.expect(token.Select)
	s := &Select{
		comments: p.comments(),
		startLoc: p.lex.Start,
	}
	p.next()
	p.expect(token.OpenBrace)
	p.next()
	for p.tok == token.Case || p.tok == token.Default {
		s.Cases = append(s.Cases, parseCommCase(p))
	}
	p.expect(token.CloseBrace)
	s.endLoc = p.lex.Start
	p.next()
	return s
}

func parseCommCase(p *Parser) CommCase {
	var c CommCase
	if p.tok == token.Default {
		p.next()
	} else {
		p.expect(token.Case)
		p.next()

		s := parseSendOrRecvStmt(p)
		if recv, ok := s.(*RecvStmt); ok {
			c.Receive = recv
		} else if send, ok := s.(*SendStmt); ok {
			c.Send = send
		} else {
			panic("malformed communication case")
		}
	}
	p.expect(token.Colon)
	p.next()
	c.Statements = parseCaseStatements(p)
	return c
}

func parseSendOrRecvStmt(p *Parser) Statement {
	cmnts := p.comments()
	expr := parseExpr(p)
	switch p.tok {
	case token.Comma:
		p.next()
		exprs := append([]Expression{expr}, parseExpressionList(p)...)
		return parseRecvStmtTail(p, cmnts, exprs)

	case token.Equal, token.ColonEqual:
		return parseRecvStmtTail(p, cmnts, []Expression{expr})

	case token.LessMinus:
		p.next()
		return &SendStmt{
			comments:   cmnts,
			Channel:    expr,
			Expression: parseExpr(p),
		}
	}
	panic(p.err("send or receive statement"))
}

func parseRecvStmtTail(p *Parser, cmnts comments, left []Expression) Statement {
	ids := true
	for _, e := range left {
		if _, ok := e.(*Identifier); !ok {
			ids = false
			break
		}
	}
	recv := &RecvStmt{
		comments: cmnts,
		Op:       p.tok,
		Left:     left,
	}
	if ids && p.tok == token.ColonEqual {
		p.next()
	} else {
		p.expect(token.Equal)
		p.next()
	}
	p.expect(token.LessMinus)
	recv.Right = *parseUnaryExpr(p, false).(*UnaryOp)
	return recv
}

func parseSwitch(p *Parser) Statement {
	p.expect(token.Switch)
	loc := p.lex.Start
	cmnts := p.comments()
	p.next()

	prevLevel := p.exprLevel
	p.exprLevel = -1
	stmt := parseSimpleStmt(p, typeSwitchOK)
	p.exprLevel = prevLevel

	if expr, id := guardStatement(stmt); expr != nil {
		return parseTypeSwitchBlock(p, loc, cmnts, nil, expr, id)
	}
	if expr, ok := stmt.(*ExpressionStmt); ok && p.tok == token.OpenBrace {
		return parseExprSwitchBlock(p, loc, cmnts, nil, expr.Expression)
	}
	if p.tok == token.OpenBrace {
		return parseExprSwitchBlock(p, loc, cmnts, nil, nil)
	}

	p.expect(token.Semicolon)
	p.next()

	prevLevel, p.exprLevel = p.exprLevel, -1
	expr := parseExpression(p, true)
	p.exprLevel = prevLevel

	if id, ok := expr.(*Identifier); ok && p.tok == token.ColonEqual {
		// Must be a type switch with a declaration.
		p.next()
		ta, ok := parsePrimaryExpr(p, true).(*TypeAssertion)
		if !ok || ta.Type != nil {
			panic(p.err("type switch guard"))
		}
		return parseTypeSwitchBlock(p, loc, cmnts, stmt, ta.Expression, id)
	}
	if ta, ok := expr.(*TypeAssertion); ok && ta.Type == nil {
		return parseTypeSwitchBlock(p, loc, cmnts, stmt, ta.Expression, nil)
	}

	return parseExprSwitchBlock(p, loc, cmnts, stmt, expr)
}

// Returns the expression for a type switch guard or nil if the statement is not
// a type switch guard.  If the statement is a type switch guard that includes
// a short variable declaration, the identifier from the declaration is returned
// as the second value.
func guardStatement(stmt Statement) (Expression, *Identifier) {
	switch s := stmt.(type) {
	case *ShortVarDecl:
		if len(s.Right) != 1 {
			break
		}
		if ta, ok := s.Right[0].(*TypeAssertion); ok && ta.Type == nil {
			if len(s.Left) != 1 {
				panic("too many identifiers in a type switch guard")
			}
			return ta.Expression, &s.Left[0]
		}
		break

	case *ExpressionStmt:
		if ta, ok := s.Expression.(*TypeAssertion); ok && ta.Type == nil {
			return ta.Expression, nil
		}
	}
	return nil, nil
}

func parseExprSwitchBlock(p *Parser, loc token.Location, cmnts comments,
	init Statement, expr Expression) *ExprSwitch {
	p.expect(token.OpenBrace)
	p.next()

	sw := &ExprSwitch{
		comments:       cmnts,
		startLoc:       loc,
		Initialization: init,
		Expression:     expr,
	}
	for p.tok == token.Case || p.tok == token.Default {
		sw.Cases = append(sw.Cases, parseExprCase(p))
	}
	p.expect(token.CloseBrace)
	sw.endLoc = p.lex.Start
	p.next()
	return sw
}

func parseExprCase(p *Parser) (c ExprCase) {
	if p.tok != token.Default && p.tok != token.Case {
		panic(p.err(token.Default, token.Case))
	}
	def := p.tok == token.Default
	p.next()
	if !def {
		c.Expressions = parseExpressionList(p)
	}
	p.expect(token.Colon)
	p.next()
	c.Statements = parseCaseStatements(p)
	return c
}

func parseTypeSwitchBlock(p *Parser, loc token.Location, cmnts comments,
	init Statement, expr Expression, id *Identifier) *TypeSwitch {
	p.expect(token.OpenBrace)
	p.next()

	sw := &TypeSwitch{
		comments:       cmnts,
		startLoc:       loc,
		Initialization: init,
		Declaration:    id,
		Expression:     expr,
	}
	for p.tok == token.Case || p.tok == token.Default {
		sw.Cases = append(sw.Cases, parseTypeCase(p))
	}
	p.expect(token.CloseBrace)
	sw.endLoc = p.lex.Start
	p.next()
	return sw
}

func parseTypeCase(p *Parser) (c TypeCase) {
	if p.tok != token.Default && p.tok != token.Case {
		panic(p.err(token.Default, token.Case))
	}
	def := p.tok == token.Default
	p.next()
	if !def {
		c.Types = parseTypeList(p)
	}
	p.expect(token.Colon)
	p.next()
	c.Statements = parseCaseStatements(p)
	return c
}

func parseTypeList(p *Parser) []Type {
	types := []Type{parseType(p)}
	for p.tok == token.Comma {
		p.next()
		types = append(types, parseType(p))
	}
	return types
}

func parseCaseStatements(p *Parser) []Statement {
	var stmts []Statement
	for p.tok != token.Case && p.tok != token.Default && p.tok != token.CloseBrace {
		stmts = append(stmts, parseStatement(p))
		if p.tok == token.Case || p.tok == token.Default || p.tok == token.CloseBrace {
			break
		}
		p.expect(token.Semicolon)
		p.next()
	}
	return stmts
}

// A rangeClause is either an Assignment or a ShortVarDecl statement
// repressenting a range clause in a range-style for loop.
type rangeClause struct {
	Statement
}

func parseFor(p *Parser) Statement {
	f := &ForStmt{comments: p.comments(), startLoc: p.lex.Start}
	p.expect(token.For)
	p.next()

	if p.tok == token.OpenBrace {
		f.Block = *parseBlock(p)
		return f
	}

	prevLevel := p.exprLevel
	p.exprLevel = -1
	stmt := parseSimpleStmt(p, rangeOK)
	p.exprLevel = prevLevel

	if r, ok := stmt.(rangeClause); ok {
		f.Range = r.Statement
	} else if ex, ok := stmt.(*ExpressionStmt); ok && p.tok == token.OpenBrace {
		f.Condition = ex.Expression
	} else {
		f.Initialization = stmt
		p.expect(token.Semicolon)
		p.next()
		if p.tok != token.Semicolon {
			prevLevel, p.exprLevel = p.exprLevel, -1
			f.Condition = parseExpr(p)
			p.exprLevel = prevLevel
		}
		p.expect(token.Semicolon)
		p.next()

		prevLevel, p.exprLevel = p.exprLevel, -1
		f.Post = parseSimpleStmt(p, none)
		p.exprLevel = prevLevel

	}
	f.Block = *parseBlock(p)
	return f
}

func parseIf(p *Parser) Statement {
	ifst := &IfStmt{comments: p.comments(), startLoc: p.lex.Start}
	p.expect(token.If)
	p.next()

	prevLevel := p.exprLevel
	p.exprLevel = -1
	stmt := parseSimpleStmt(p, none)
	p.exprLevel = prevLevel

	if expr, ok := stmt.(*ExpressionStmt); ok && p.tok == token.OpenBrace {
		ifst.Condition = expr.Expression
		ifst.Block = *parseBlock(p)
	} else {
		p.expect(token.Semicolon)
		p.next()
		ifst.Statement = stmt
		prevLevel, p.exprLevel = p.exprLevel, -1
		ifst.Condition = parseExpr(p)
		p.exprLevel = prevLevel
		ifst.Block = *parseBlock(p)
	}
	if p.tok != token.Else {
		return ifst
	}
	p.next()
	if p.tok == token.If {
		ifst.Else = parseIf(p)
		return ifst
	}
	ifst.Else = parseBlock(p)
	return ifst
}

func parseBlock(p *Parser) *BlockStmt {
	p.expect(token.OpenBrace)
	c, s := p.comments(), p.lex.Start
	p.next()
	var stmts []Statement
	for p.tok != token.CloseBrace {
		stmt := parseStatement(p)
		if stmt != nil {
			stmts = append(stmts, stmt)
		}
		if p.tok == token.CloseBrace {
			break
		}
		p.expect(token.Semicolon)
		p.next()
	}
	p.expect(token.CloseBrace)
	e := p.lex.End
	p.next()
	return &BlockStmt{
		comments:   c,
		startLoc:   s,
		endLoc:     e,
		Statements: stmts,
	}
}

func parseGo(p *Parser) Statement {
	p.expect(token.Go)
	c, s := p.comments(), p.lex.Start
	p.next()
	return &GoStmt{
		comments:   c,
		startLoc:   s,
		Expression: parseExpr(p),
	}
}

func parseDefer(p *Parser) Statement {
	p.expect(token.Defer)
	c, s := p.comments(), p.lex.Start
	p.next()
	return &DeferStmt{
		comments:   c,
		startLoc:   s,
		Expression: parseExpr(p),
	}
}

func parseReturn(p *Parser) Statement {
	p.expect(token.Return)
	c, s, e := p.comments(), p.lex.Start, p.lex.End
	p.next()
	var exprs []Expression
	if expressionFirst[p.tok] {
		exprs = parseExpressionList(p)
		e = exprs[len(exprs)-1].End()
	}
	return &ReturnStmt{
		comments:    c,
		startLoc:    s,
		endLoc:      e,
		Expressions: exprs,
	}
}

func parseGoto(p *Parser) Statement {
	p.expect(token.Goto)
	c, s := p.comments(), p.lex.Start
	p.next()
	return &GotoStmt{
		comments: c,
		startLoc: s,
		Label:    *parseIdentifier(p),
	}
}

func parseContinue(p *Parser) Statement {
	p.expect(token.Continue)
	c, s := p.comments(), p.lex.Start
	p.next()
	var l *Identifier
	if p.tok == token.Identifier {
		l = parseIdentifier(p)
	}
	return &ContinueStmt{
		comments: c,
		startLoc: s,
		Label:    l,
	}
}

func parseBreak(p *Parser) Statement {
	p.expect(token.Break)
	c, s := p.comments(), p.lex.Start
	p.next()
	var l *Identifier
	if p.tok == token.Identifier {
		l = parseIdentifier(p)
	}
	return &BreakStmt{
		comments: c,
		startLoc: s,
		Label:    l,
	}
}

// AssignOps is a slice of all assignment operatiors.
var assignOps = []token.Token{
	token.Equal,
	token.PlusEqual,
	token.MinusEqual,
	token.OrEqual,
	token.CarrotEqual,
	token.StarEqual,
	token.DivideEqual,
	token.PercentEqual,
	token.LessLessEqual,
	token.GreaterGreaterEqual,
	token.AndEqual,
	token.AndCarrotEqual,
}

// AssignOp is the set of assignment operators.
var assignOp = func() map[token.Token]bool {
	ops := make(map[token.Token]bool)
	for _, op := range assignOps {
		ops[op] = true
	}
	return ops
}()

// Returns the current token if it is an assignment operator, otherwise
// panics with a syntax error.
func expectAssign(p *Parser) token.Token {
	if assignOp[p.tok] {
		return p.tok
	}
	ops := make([]interface{}, len(assignOps)-1)
	for i, op := range assignOps[1:] {
		ops[i] = op
	}
	panic(p.err(assignOps[0], ops...))
}

// SimpOptions are some options that allow parseSimpleStmt to return
// non-simple statements.
type options int

const (
	none options = iota
	// LabelOK allows parseSimpleStmt to return label statements.
	labelOK
	// RangeOK allows parseSimpleStmt to return RangeClauses
	// for either assingment or short variable declarations.
	rangeOK
	// TypeSwitchOK allows for a type switch guard in short variable
	// declarations.
	typeSwitchOK
)

func parseSimpleStmt(p *Parser, opts options) (st Statement) {
	cmnts := p.comments()
	if !expressionFirst[p.tok] {
		// Empty statement
		return nil
	}
	expr := parseExpression(p, opts == typeSwitchOK)
	id, isID := expr.(*Identifier)
	switch {
	case p.tok == token.LessMinus:
		p.next()
		return &SendStmt{
			comments:   cmnts,
			Channel:    expr,
			Expression: parseExpr(p),
		}

	case p.tok == token.MinusMinus || p.tok == token.PlusPlus:
		op, opEnd := p.tok, p.lex.End
		p.next()
		return &IncDecStmt{
			comments:   cmnts,
			Expression: expr,
			Op:         op,
			opEnd:      opEnd,
		}

	case assignOp[p.tok]:
		exprs := []Expression{expr}
		return parseAssignmentTail(p, cmnts, exprs, opts == rangeOK)

	case p.tok == token.Comma:
		p.next()
		exprs := append([]Expression{expr}, parseExpressionList(p)...)

		var ids []Identifier
		for _, e := range exprs {
			if id, ok := e.(*Identifier); ok {
				ids = append(ids, *id)
			} else {
				break
			}
		}
		// If all the expressions were identifiers then we could have
		// a short variable declaration.  Otherwise, it's an assignment.
		if len(ids) == len(exprs) && p.tok == token.ColonEqual {
			// A type switch guard is only allowed if there is a single
			// identifier on the left hand side.
			if opts == typeSwitchOK {
				opts = none
			}
			return parseShortVarDeclTail(p, cmnts, ids, opts)
		}
		return parseAssignmentTail(p, cmnts, exprs, opts == rangeOK)

	case isID && p.tok == token.ColonEqual:
		ids := []Identifier{*id}
		return parseShortVarDeclTail(p, cmnts, ids, opts)

	case opts == labelOK && isID && p.tok == token.Colon:
		p.next()
		return &LabeledStmt{
			comments:  cmnts,
			Label:     *id,
			Statement: parseStatement(p),
		}

	default:
		return &ExpressionStmt{
			comments:   cmnts,
			Expression: expr,
		}
	}
}

// Parses a short variable declaration beginning with the := operator.  If
// allowRange is true, then a rangeClause is returned if the range
// keyword appears after the := operator.
func parseShortVarDeclTail(p *Parser, cmnts comments, ids []Identifier, opts options) Statement {
	p.expect(token.ColonEqual)
	p.next()
	if opts == rangeOK && p.tok == token.Range {
		p.next()
		return rangeClause{&ShortVarDecl{
			comments: cmnts,
			Left:     ids,
			Right:    []Expression{parseExpr(p)},
		}}
	}
	var right []Expression
	if opts == typeSwitchOK {
		right = parseExpressionListOrTypeGuard(p)
	} else {
		right = parseExpressionList(p)
	}
	return &ShortVarDecl{
		comments: cmnts,
		Left:     ids,
		Right:    right,
	}
}

// Parses an assignment statement beginning with the assignment
// operator.  If allowRange is true, then a rangeClause is returned
// if the range keyword appears after the assignment operator.
func parseAssignmentTail(p *Parser, cmnts comments, exprs []Expression, rangeOK bool) Statement {
	op := expectAssign(p)
	p.next()
	if rangeOK && op == token.Equal && p.tok == token.Range {
		p.next()
		return rangeClause{&Assignment{
			comments: cmnts,
			Op:       op,
			Left:     exprs,
			Right:    []Expression{parseExpr(p)},
		}}
	}
	return &Assignment{
		comments: cmnts,
		Op:       op,
		Left:     exprs,
		Right:    parseExpressionList(p),
	}
}

func parseImportDecl(p *Parser) *ImportDecl {
	p.expect(token.Import)
	d := &ImportDecl{comments: p.comments(), startLoc: p.lex.Start}
	p.next()
	if p.tok == token.OpenParen {
		p.next()
		for p.tok != token.CloseParen {
			imp := parseImportSpec(p)
			d.Imports = append(d.Imports, imp)
			if p.tok == token.CloseParen {
				break
			}
			p.expect(token.Semicolon)
			p.next()
		}
		p.expect(token.CloseParen)
		d.endLoc = p.lex.Start
		p.next()
		return d
	}
	imp := parseImportSpec(p)
	d.Imports = []ImportSpec{imp}
	d.endLoc = imp.Path.End()
	return d
}

func parseImportSpec(p *Parser) ImportSpec {
	var s ImportSpec
	if p.tok == token.Dot {
		p.next()
		s.Dot = true
		s.Path = *parseStringLiteral(p)
		return s
	}
	if p.tok != token.StringLiteral {
		s.Name = parseIdentifier(p)
	}
	s.Path = *parseStringLiteral(p)
	return s
}

func parseTopLevelDecl(p *Parser) Declarations {
	if p.tok == token.Func {
		return Declarations{parseFunctionOrMethodDecl(p)}
	}
	return parseDeclarations(p)
}

func parseFunctionOrMethodDecl(p *Parser) Declaration {
	p.expect(token.Func)
	cmnts := p.comments()
	l := p.lex.Start
	p.next()
	if p.tok == token.OpenParen {
		p.next()
		m := &MethodDecl{comments: cmnts, startLoc: l}
		m.Receiver = *parseIdentifier(p)
		if p.tok == token.Star {
			p.next()
			m.Pointer = true
		}
		m.BaseTypeName = *parseIdentifier(p)
		p.expect(token.CloseParen)
		p.next()
		m.Name = *parseIdentifier(p)
		m.Signature = parseSignature(p)
		m.Body = *parseBlock(p)
		return m
	}
	return &FunctionDecl{
		comments:  cmnts,
		startLoc:  l,
		Name:      *parseIdentifier(p),
		Signature: parseSignature(p),
		Body:      *parseBlock(p),
	}
}

func parseDeclarations(p *Parser) Declarations {
	switch p.tok {
	case token.Type:
		return parseTypeDecl(p)
	case token.Const:
		return parseConstDecl(p)
	case token.Var:
		return parseVarDecl(p)
	}
	panic(p.err("type", "const", "var"))
}

func parseVarDecl(p *Parser) Declarations {
	var decls Declarations
	p.expect(token.Var)
	cmnts := p.comments()
	p.next()

	if p.tok != token.OpenParen {
		cs := parseVarSpec(p)
		cs.comments = cmnts
		return append(decls, cs)
	}
	p.next()

	for p.tok != token.CloseParen {
		decls = append(decls, parseVarSpec(p))
		if p.tok == token.CloseParen {
			break
		}
		p.expect(token.Semicolon)
		p.next()
	}
	p.next()
	return decls
}

func parseVarSpec(p *Parser) *VarSpec {
	vs := &VarSpec{
		comments: p.comments(),
		Names:    parseIdentifierList(p),
	}
	if typeFirst[p.tok] {
		vs.Type = parseType(p)
		if p.tok != token.Equal {
			return vs
		}
	}
	p.expect(token.Equal)
	p.next()
	vs.Values = parseExpressionList(p)
	return vs
}

func parseConstDecl(p *Parser) Declarations {
	var decls Declarations
	p.expect(token.Const)
	cmnts := p.comments()
	p.next()

	if p.tok != token.OpenParen {
		cs := parseConstSpec(p)
		cs.comments = cmnts
		return append(decls, cs)
	}
	p.next()

	for p.tok != token.CloseParen {
		decls = append(decls, parseConstSpec(p))
		if p.tok == token.CloseParen {
			break
		}
		p.expect(token.Semicolon)
		p.next()
	}
	p.expect(token.CloseParen)
	p.next()
	return decls
}

func parseConstSpec(p *Parser) *ConstSpec {
	cs := &ConstSpec{
		comments: p.comments(),
		Names:    parseIdentifierList(p),
	}
	if typeFirst[p.tok] {
		cs.Type = parseType(p)
	}
	if p.tok == token.Equal {
		p.next()
		cs.Values = parseExpressionList(p)
	}
	return cs
}

func parseIdentifierList(p *Parser) []Identifier {
	var ids []Identifier
	for {
		ids = append(ids, *parseIdentifier(p))
		if p.tok != token.Comma {
			break
		}
		p.next()
	}
	return ids
}

func parseTypeDecl(p *Parser) Declarations {
	var decls Declarations
	p.expect(token.Type)
	cmnts := p.comments()
	p.next()

	if p.tok != token.OpenParen {
		ts := parseTypeSpec(p)
		ts.comments = cmnts
		return append(decls, ts)
	}
	p.next()

	for p.tok != token.CloseParen {
		decls = append(decls, parseTypeSpec(p))
		if p.tok == token.CloseParen {
			break
		}
		p.expect(token.Semicolon)
		p.next()
	}
	p.expect(token.CloseParen)
	p.next()
	return decls
}

func parseTypeSpec(p *Parser) *TypeSpec {
	return &TypeSpec{
		comments: p.comments(),
		Name:     *parseIdentifier(p),
		Type:     parseType(p),
	}
}

// TypeFirst is the set of tokens that can start a type.
var typeFirst = map[token.Token]bool{
	token.Identifier:  true,
	token.Star:        true,
	token.OpenBracket: true,
	token.Struct:      true,
	token.Func:        true,
	token.Interface:   true,
	token.Map:         true,
	token.Chan:        true,
	token.LessMinus:   true,
	token.OpenParen:   true,
}

func parseType(p *Parser) Type {
	switch p.tok {
	case token.Identifier:
		return parseTypeName(p)

	case token.Star:
		starLoc := p.lex.Start
		p.next()
		return &Star{Target: parseType(p), starLoc: starLoc}

	case token.OpenBracket:
		return parseArrayOrSliceType(p, false)

	case token.Struct:
		return parseStructType(p)

	case token.Func:
		l := p.lex.Start
		p.next()
		return &FunctionType{
			Signature: parseSignature(p),
			funcLoc:   l,
		}

	case token.Interface:
		return parseInterfaceType(p)

	case token.Map:
		return parseMapType(p)

	case token.Chan:
		fallthrough
	case token.LessMinus:
		return parseChannelType(p)

	case token.OpenParen:
		p.next()
		t := parseType(p)
		p.expect(token.CloseParen)
		p.next()
		return t
	}

	panic(p.err("type"))
}

func parseStructType(p *Parser) *StructType {
	p.expect(token.Struct)
	st := &StructType{keywordLoc: p.lex.Start}
	p.next()
	p.expect(token.OpenBrace)
	p.next()

	for p.tok != token.CloseBrace {
		field := parseFieldDecl(p)
		st.Fields = append(st.Fields, field)
		if p.tok == token.CloseBrace {
			break
		}
		p.expect(token.Semicolon)
		p.next()
	}

	p.expect(token.CloseBrace)
	st.closeLoc = p.lex.Start
	p.next()
	return st
}

func parseFieldDecl(p *Parser) FieldDecl {
	var id *Identifier

	d := FieldDecl{}
	if p.tok == token.Star {
		l := p.lex.Start
		p.next()
		d.Type = &Star{Target: parseTypeName(p), starLoc: l}
		goto tag
	}

	id = parseIdentifier(p)
	switch p.tok {
	case token.Dot:
		l := p.lex.Start
		p.next()
		d.Type = &Selector{
			Parent: id,
			Name:   parseIdentifier(p),
			dotLoc: l,
		}
		goto tag

	case token.StringLiteral:
		d.Tag = parseStringLiteral(p)
		fallthrough
	case token.Semicolon:
		fallthrough
	case token.CloseBrace:
		d.Type = id
		return d
	}

	d.Identifiers = []Identifier{*id}
	for p.tok == token.Comma {
		p.next()
		id := parseIdentifier(p)
		d.Identifiers = append(d.Identifiers, *id)
	}
	d.Type = parseType(p)

tag:
	if p.tok == token.StringLiteral {
		d.Tag = parseStringLiteral(p)
	}
	return d
}

func parseInterfaceType(p *Parser) *InterfaceType {
	p.expect(token.Interface)
	it := &InterfaceType{keywordLoc: p.lex.Start}
	p.next()
	p.expect(token.OpenBrace)
	p.next()

	for p.tok != token.CloseBrace {
		id := parseIdentifier(p)
		switch p.tok {
		case token.OpenParen:
			it.Methods = append(it.Methods, &Method{
				Name:      *id,
				Signature: parseSignature(p),
			})

		case token.Dot:
			l := p.lex.Start
			p.next()
			it.Methods = append(it.Methods, &Selector{
				Parent: id,
				Name:   parseIdentifier(p),
				dotLoc: l,
			})

		default:
			it.Methods = append(it.Methods, id)
		}
		if p.tok == token.CloseBrace {
			break
		}
		p.expect(token.Semicolon)
		p.next()
	}

	p.expect(token.CloseBrace)
	it.closeLoc = p.lex.Start
	p.next()
	return it
}

func parseSignature(p *Parser) Signature {
	s := Signature{Parameters: parseParameterList(p)}
	if p.tok == token.OpenParen {
		s.Result = parseParameterList(p)
	} else if typeFirst[p.tok] {
		t := parseType(p)
		s.Result = ParameterList{
			Parameters: []ParameterDecl{{Type: t}},
			openLoc:    t.Start(),
			closeLoc:   t.End(),
		}
	}
	return s
}

// Parsing a parameter list is a bit complex.  The grammar productions
// in the spec are more permissive than the language actually allows.
// The text of the spec restricts parameter lists to be either a series of
// parameter declarations without identifiers and only types, or a
// series of declarations that all have one or more identifiers.  Instead,
// we use the grammar below, which only allows one type of list or the
// other, but not both.
//
// ParameterList = "(" ParameterListTail
// ParameterListTail =
// 	| “)”
// 	| Identifier “,” ParameterListTail
// 	| Identifier “.” Identifier TypeParameterList
// 	| Identifier Type DeclParameterList
// 	| NonTypeNameType TypeParameterList
// 	| “...” Type ")"
// TypeParameterList =
// 	| “)”
// 	| "," ")"
// 	| “,” Type TypeParameterList
// 	| “,” “...” Type ")"
// DeclParameterList =
// 	| ")"
// 	| "," ")"
// 	| "," IdentifierList Type DeclParameterList
// 	| "," IdentifierList "..." Type ")"
// IdentifierList =
// 	| Identifier “,” IdentifierList
// 	| Identifier
func parseParameterList(p *Parser) ParameterList {
	p.expect(token.OpenParen)
	pl := ParameterList{openLoc: p.lex.Start}
	p.next()
	parseParameterListTail(p, &pl, nil)
	p.expect(token.CloseParen)
	pl.closeLoc = p.lex.Start
	p.next()
	return pl
}

func parseParameterListTail(p *Parser, pl *ParameterList, idents []Identifier) {
	switch {
	case p.tok == token.CloseParen:
		pl.closeLoc = p.lex.Start
		pl.Parameters = typeNameDecls(idents)
		return

	case p.tok == token.Identifier:
		id := parseIdentifier(p)
		switch {
		case p.tok == token.Comma:
			p.next()
			fallthrough
		case p.tok == token.CloseParen:
			parseParameterListTail(p, pl, append(idents, *id))
			return

		case p.tok == token.Dot:
			l := p.lex.Start
			p.next()
			t := &Selector{
				Parent: id,
				Name:   parseIdentifier(p),
				dotLoc: l,
			}
			d := ParameterDecl{Type: t}
			pl.Parameters = append(typeNameDecls(idents), d)
			parseTypeParameterList(p, pl)
			return

		default:
			idents = append(idents, *id)
			d := ParameterDecl{Identifiers: idents}
			if p.tok == token.DotDotDot {
				d.DotDotDot = true
				p.next()
			}
			d.Type = parseType(p)
			pl.Parameters = []ParameterDecl{d}
			if !d.DotDotDot {
				parseDeclParameterList(p, pl)
			}
			return
		}

	case p.tok == token.DotDotDot:
		p.next()
		d := ParameterDecl{Type: parseType(p), DotDotDot: true}
		pl.Parameters = append(typeNameDecls(idents), d)
		return

	case typeFirst[p.tok]:
		d := ParameterDecl{Type: parseType(p)}
		pl.Parameters = append(typeNameDecls(idents), d)
		parseTypeParameterList(p, pl)
		return
	}

	panic(p.err(")", "...", "identifier", "type"))
}

func parseTypeParameterList(p *Parser, pl *ParameterList) {
	if p.tok == token.CloseParen {
		return
	}
	p.expect(token.Comma)
	p.next()

	// Allow trailing comma.
	if p.tok == token.CloseParen {
		return
	}

	d := ParameterDecl{}
	if p.tok == token.DotDotDot {
		d.DotDotDot = true
		p.next()
	}
	d.Type = parseType(p)
	pl.Parameters = append(pl.Parameters, d)
	if !d.DotDotDot {
		parseTypeParameterList(p, pl)
	}
	return
}

func parseDeclParameterList(p *Parser, pl *ParameterList) {
	if p.tok == token.CloseParen {
		return
	}

	p.expect(token.Comma)
	p.next()

	// Allow trailing comma.
	if p.tok == token.CloseParen {
		return
	}

	d := ParameterDecl{}
	for {
		id := parseIdentifier(p)
		d.Identifiers = append(d.Identifiers, *id)

		if p.tok != token.Comma {
			break
		}
		p.next()
	}
	if p.tok == token.DotDotDot {
		d.DotDotDot = true
		p.next()
	}
	d.Type = parseType(p)
	pl.Parameters = append(pl.Parameters, d)
	parseDeclParameterList(p, pl)
	return
}

func typeNameDecls(idents []Identifier) []ParameterDecl {
	decls := make([]ParameterDecl, len(idents))
	for i := range idents {
		decls[i].Type = &idents[i]
	}
	return decls
}

func parseChannelType(p *Parser) Type {
	ch := &ChannelType{Send: true, Receive: true, startLoc: p.lex.Start}
	if p.tok == token.LessMinus {
		ch.Send = false
		p.next()
	}
	p.expect(token.Chan)
	p.next()
	if ch.Send && p.tok == token.LessMinus {
		ch.Receive = false
		p.next()
	}
	ch.Type = parseType(p)
	return ch
}

func parseMapType(p *Parser) Type {
	p.expect(token.Map)
	m := &MapType{mapLoc: p.lex.Start}
	p.next()
	p.expect(token.OpenBracket)
	p.next()
	m.Key = parseType(p)
	p.expect(token.CloseBracket)
	p.next()
	m.Type = parseType(p)
	return m
}

// Parses an array or slice type.  If dotDotDot is true then it will accept an
// array with a size specified a "..." token, otherwise it will require a size.
func parseArrayOrSliceType(p *Parser, dotDotDot bool) Type {
	p.expect(token.OpenBracket)
	openLoc := p.lex.Start
	p.next()

	if p.tok == token.CloseBracket {
		p.next()
		sl := &SliceType{Type: parseType(p), openLoc: openLoc}
		return sl
	}
	ar := &ArrayType{openLoc: openLoc}
	if dotDotDot && p.tok == token.DotDotDot {
		p.next()
	} else {
		ar.Size = parseExpr(p)
	}
	p.expect(token.CloseBracket)
	p.next()
	ar.Type = parseType(p)
	return ar
}

func parseTypeName(p *Parser) Type {
	p.expect(token.Identifier)
	n := parseIdentifier(p)
	if p.tok == token.Dot {
		l := p.lex.Start
		p.next()
		return &Selector{
			Parent: n,
			Name:   parseIdentifier(p),
			dotLoc: l,
		}
	}
	return n
}

var (
	// ExpressionFirst is the set of tokens that may begin an expression.
	expressionFirst = map[token.Token]bool{
		// Unary Op
		token.Plus:      true,
		token.Minus:     true,
		token.Bang:      true,
		token.Carrot:    true,
		token.Star:      true,
		token.And:       true,
		token.LessMinus: true,

		// Type First
		token.Identifier: true,
		//	token.Star:        true,
		token.OpenBracket: true,
		token.Struct:      true,
		token.Func:        true,
		token.Interface:   true,
		token.Map:         true,
		token.Chan:        true,
		//	token.LessMinus:   true,
		token.OpenParen: true,

		// Literals
		token.IntegerLiteral:   true,
		token.FloatLiteral:     true,
		token.ImaginaryLiteral: true,
		token.RuneLiteral:      true,
		token.StringLiteral:    true,
	}

	// Binary op precedence for precedence climbing algorithm.
	// http://www.engr.mun.ca/~theo/Misc/exp_parsing.htm
	//
	// BUG(eaburns): Define tokens.NTokens and change
	// map[token.Token]Whatever to [nTokens]Whatever.
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

func parseExpr(p *Parser) Expression {
	return parseExpression(p, false)
}

func parseExpression(p *Parser, typeSwitch bool) Expression {
	return parseBinaryExpr(p, 1, typeSwitch)
}

func parseBinaryExpr(p *Parser, prec int, typeSwitch bool) Expression {
	left := parseUnaryExpr(p, typeSwitch)
	if ta, ok := left.(*TypeAssertion); ok && ta.Type == nil {
		if !typeSwitch {
			panic("parsed a disallowed type switch guard")
		}
		// This is a type guard, it cannot be the left operand of a
		// binary expression.
		return left
	}
	for {
		pr, ok := precedence[p.tok]
		if !ok || pr < prec {
			return left
		}
		op, opLoc := p.tok, p.lex.Start
		p.next()
		right := parseBinaryExpr(p, pr+1, false)
		left = &BinaryOp{
			Op:    op,
			opLoc: opLoc,
			Left:  left,
			Right: right,
		}
	}
}

func parseUnaryExpr(p *Parser, typeSwitch bool) Expression {
	if unary[p.tok] {
		op, opLoc := p.tok, p.lex.Start
		p.next()
		operand := parseUnaryExpr(p, false)
		if op == token.Star {
			return &Star{Target: operand, starLoc: opLoc}
		}
		return &UnaryOp{
			Op:      op,
			opLoc:   opLoc,
			Operand: operand,
		}
	}
	return parsePrimaryExpr(p, typeSwitch)
}

func parsePrimaryExpr(p *Parser, typeSwitch bool) Expression {
	left := parseOperand(p, typeSwitch)
	if t, ok := left.(Type); ok && p.exprLevel >= 0 && p.tok == token.OpenBrace {
		lit := parseLiteralValue(p)
		lit.Type = t
		return lit
	}
	for {
		switch p.tok {
		case token.OpenBracket:
			p.exprLevel++
			left = parseSliceOrIndex(p, left)
			p.exprLevel--
		case token.OpenParen:
			p.exprLevel++
			left = parseCall(p, left)
			p.exprLevel--
		case token.Dot:
			left = parseSelectorOrTypeAssertion(p, left, typeSwitch)
		default:
			return left
		}
	}
}

func parseSliceOrIndex(p *Parser, left Expression) Expression {
	p.expect(token.OpenBracket)
	openLoc := p.lex.Start
	p.next()

	if p.tok == token.Colon {
		sl := parseSliceHighMax(p, left, nil)
		sl.openLoc = openLoc
		return sl
	}

	e := parseExpr(p)

	switch p.tok {
	case token.CloseBracket:
		index := &Index{Expression: left, Index: e, openLoc: openLoc}
		p.expect(token.CloseBracket)
		index.closeLoc = p.lex.End
		p.next()
		return index

	case token.Colon:
		sl := parseSliceHighMax(p, left, e)
		sl.openLoc = openLoc
		return sl
	}

	panic(p.err(token.CloseBracket, token.Colon))
}

// Parses the remainder of a slice expression, beginning from the
// colon after the low term of the expression.  The returned Slice
// node does not have its openLoc field set; it must be set by the
// caller.
func parseSliceHighMax(p *Parser, left, low Expression) *Slice {
	p.expect(token.Colon)
	p.next()

	sl := &Slice{Expression: left, Low: low}
	if p.tok != token.CloseBracket {
		sl.High = parseExpr(p)
		if p.tok == token.Colon {
			p.next()
			sl.Max = parseExpr(p)
		}
	}
	p.expect(token.CloseBracket)
	sl.closeLoc = p.lex.Start
	p.next()
	return sl
}

func parseCall(p *Parser, left Expression) Expression {
	p.expect(token.OpenParen)
	c := &Call{Function: left, openLoc: p.lex.Start}
	p.next()
	if p.tok != token.CloseParen {
		c.Arguments = parseExpressionList(p)
		if p.tok == token.DotDotDot {
			c.DotDotDot = true
			p.next()
		}
	}
	p.expect(token.CloseParen)
	c.closeLoc = p.lex.End
	p.next()
	return c
}

func parseExpressionList(p *Parser) []Expression {
	var exprs []Expression
	for {
		exprs = append(exprs, parseExpr(p))
		if p.tok != token.Comma {
			break
		}
		p.next()
	}
	return exprs
}

func parseExpressionListOrTypeGuard(p *Parser) []Expression {
	var exprs []Expression
	for {
		expr := parseExpression(p, len(exprs) == 0)
		exprs = append(exprs, expr)
		if ta, ok := expr.(*TypeAssertion); ok && ta.Type == nil {
			if len(exprs) > 1 {
				panic("parsed disallowed type switch guard")
			}
			// Type switch guard.  It must be first and nothing can follow it.
			break
		}
		if p.tok != token.Comma {
			break
		}
		p.next()
	}
	return exprs
}

func parseOperand(p *Parser, typeSwitch bool) Expression {
	switch p.tok {
	case token.Identifier:
		id := parseIdentifier(p)
		if p.tok == token.Dot {
			return parseSelectorOrTypeAssertion(p, id, typeSwitch)
		}
		return id

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

	case token.Func:
		return parseFunctionLiteral(p)

	case token.OpenParen:
		p.next()
		p.exprLevel++
		e := parseExpr(p)
		p.exprLevel--
		p.expect(token.CloseParen)
		p.next()
		return e

	case token.OpenBracket:
		return parseArrayOrSliceType(p, true)
	}

	if typeFirst[p.tok] {
		return parseType(p)
	}
	panic(p.err("operand"))
}

func parseFunctionLiteral(p *Parser) *FunctionLiteral {
	p.expect(token.Func)
	f := &FunctionLiteral{startLoc: p.lex.Start}
	p.next()
	f.Signature = parseSignature(p)
	f.Body = *parseBlock(p)
	return f
}

func parseLiteralValue(p *Parser) *CompositeLiteral {
	p.expect(token.OpenBrace)
	v := &CompositeLiteral{openLoc: p.lex.Start}
	p.next()

	for p.tok != token.CloseBrace {
		e := parseElement(p)
		v.Elements = append(v.Elements, e)
		if p.tok != token.Comma {
			break
		}
		p.next()
	}

	p.expect(token.CloseBrace)
	v.closeLoc = p.lex.Start
	p.next()
	return v
}

func parseElement(p *Parser) Element {
	if p.tok == token.OpenBrace {
		return Element{Value: parseLiteralValue(p)}
	}

	expr := parseExpr(p)
	if p.tok != token.Colon {
		return Element{Value: expr}
	}

	p.next()
	elm := Element{Key: expr}
	if p.tok == token.OpenBrace {
		elm.Value = parseLiteralValue(p)
	} else {
		elm.Value = parseExpr(p)
	}
	return elm
}

func parseSelectorOrTypeAssertion(p *Parser, left Expression, typeSwitch bool) Expression {
	p.expect(token.Dot)
	dotLoc := p.lex.Start
	p.next()

	switch p.tok {
	case token.OpenParen:
		p.next()
		t := &TypeAssertion{Expression: left, dotLoc: dotLoc}
		if typeSwitch && p.tok == token.Type {
			p.next()
		} else {
			t.Type = parseType(p)
		}
		p.expect(token.CloseParen)
		t.closeLoc = p.lex.Start
		p.next()
		return t

	case token.Identifier:
		left = &Selector{
			Parent: left,
			Name:   parseIdentifier(p),
			dotLoc: dotLoc,
		}
		if p.tok == token.Dot {
			return parseSelectorOrTypeAssertion(p, left, typeSwitch)
		}
		return left
	}

	panic(p.err(token.OpenParen, token.Identifier))
}

func parseIntegerLiteral(p *Parser) Expression {
	l := &IntegerLiteral{Value: new(big.Int), span: p.span()}
	if _, ok := l.Value.SetString(p.lex.Text(), 0); ok {
		p.next()
		return l
	}
	// This check is needed to catch malformed octal literals that
	// are currently allowed by the lexer.  For example, 08.
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
	// I seem to recall that there was some case where the lexer
	// may return a malformed float, but I can't remember the
	// specifics.
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

	// I seem to recall that there was some case where the lexer
	// may return a malformed float, but I can't remember the
	// specifics.
	panic(&MalformedLiteral{
		Type:  "imaginary literal",
		Text:  p.text(),
		Start: p.lex.Start,
		End:   p.lex.End,
	})
}

func parseStringLiteral(p *Parser) *StringLiteral {
	text := p.lex.Text()
	if len(text) < 2 {
		panic("bad string literal: " + text)
	}
	l := &StringLiteral{span: p.span()}
	if text[0] == '`' {
		l.Value = strings.Replace(text[1:len(text)-1], "\r", "", -1)
	} else {
		var err error
		if l.Value, err = strconv.Unquote(text); err != nil {
			panic("bad string literal: " + p.lex.Text())
		}
	}
	p.next()
	return l
}

func parseIdentifier(p *Parser) *Identifier {
	p.expect(token.Identifier)
	id := &Identifier{Name: p.text(), span: p.span()}
	p.next()
	return id
}

func parseRuneLiteral(p *Parser) Expression {
	text := p.lex.Text()
	if len(text) < 3 {
		panic("bad rune literal: " + text)
	}
	r, _, _, err := strconv.UnquoteChar(text[1:], '\'')
	if err != nil {
		// The lexer may allow bad rune literals (>0x0010FFFF and
		// surrogate halves—whatever they are).
		panic(&MalformedLiteral{
			Type:  "rune literal",
			Text:  p.text(),
			Start: p.lex.Start,
			End:   p.lex.End,
		})
	}
	l := &RuneLiteral{Value: r, span: p.span()}
	p.next()
	return l
}
