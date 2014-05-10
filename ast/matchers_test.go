package ast

import (
	"bytes"
	"fmt"
	"math/big"
	"regexp"
	"strings"
	"testing"

	"github.com/velour/stop/token"
)

// NodeString returns a human-readable string representation of the AST node.
func nodeString(n Node) string {
	out := bytes.NewBuffer(nil)
	Print(out, n)
	return string(out.Bytes())
}

// A matcher function is used for testing the expected structure of ASTs.  The
// matcher returns true if the tree matches the expected structure of if there
// is an expected error.
type matcher func(Node, error) bool

func ms(ms ...matcher) []matcher {
	return ms
}

// Returns true if both the node and matcher are nil, or if they are
// both non-nil and the matcher matches the node.
func nilOr(n Node, m matcher) bool {
	return n == nil && m == nil || n != nil && m != nil && m(n, nil)
}

func parseErr(reStr string) matcher {
	re := regexp.MustCompile(reStr)
	return func(_ Node, err error) bool {
		return err != nil && re.MatchString(err.Error())
	}
}

type parserTests []struct {
	src   string
	match matcher
}

func parseTest(p *Parser, production func(*Parser) Node) (root Node, err error) {
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
	root = production(p)
	return
}

// Runs a test on the given grammar production parser rule.
func (tests parserTests) run(t *testing.T, production func(*Parser) Node) {
	for i, test := range tests {
		l := token.NewLexer("", test.src)
		p := NewParser(l)
		n, err := parseTest(p, production)
		if err != nil {
			if !test.match(nil, err) {
				t.Errorf("test %d: unexpected error parsing [%s]: %s", i, test.src, err)
			}
			continue
		}
		if !test.match(n, nil) {
			t.Errorf("test %d: failed to match [%s], got:\n%s", i, test.src, nodeString(n))
		}
	}
}

func (tests parserTests) runType(t *testing.T) {
	tests.run(t, func(p *Parser) Node {
		return parseType(p)
	})
}

func (tests parserTests) runExpr(t *testing.T) {
	tests.run(t, func(p *Parser) Node {
		return parseExpr(p)
	})
}

func (tests parserTests) runParamList(t *testing.T) {
	tests.run(t, func(p *Parser) Node {
		parms := parseParameterList(p)
		return &parms
	})
}

func (tests parserTests) runDeclarations(t *testing.T) {
	tests.run(t, func(p *Parser) Node {
		return parseDeclarations(p)
	})
}

func (tests parserTests) runStatements(t *testing.T) {
	tests.run(t, func(p *Parser) Node {
		return parseStatement(p)
	})
}

// Matches a select statement communication case.
type commMatcher struct {
	send  matcher
	recv  matcher
	stmts []matcher
}

func selectStmt(cases ...commMatcher) matcher {
	return func(n Node, err error) bool {
		s, ok := n.(*Select)
		if !ok || len(cases) != len(s.Cases) {
			return false
		}
		for i, c := range cases {
			if (c.recv == nil) != (s.Cases[i].Receive == nil) ||
				(c.recv != nil && !c.recv(s.Cases[i].Receive, nil)) ||
				(c.send == nil) != (s.Cases[i].Send == nil) ||
				(c.send != nil && !c.send(s.Cases[i].Send, nil)) ||
				len(c.stmts) != len(s.Cases[i].Statements) {
				return false
			}
			for j, m := range c.stmts {
				if !m(s.Cases[i].Statements[j], nil) {
					return false
				}
			}
		}
		return true
	}

}

func recv(left []matcher, op token.Token, recvExpr matcher) matcher {
	return func(n Node, err error) bool {
		s, ok := n.(*RecvStmt)
		if err != nil || !ok || len(left) != len(s.Left) || op != s.Op ||
			!unOp(token.LessMinus, recvExpr)(&s.Right, nil) {
			return false
		}
		for i, l := range left {
			if !l(s.Left[i], nil) {
				return false
			}
		}
		return true
	}
}

// Matches a type switch or expression switch case.
type caseMatcher struct {
	guards []matcher
	stmts  []matcher
}

func exprSwitch(init matcher, expr matcher, cases ...caseMatcher) matcher {
	return func(n Node, err error) bool {
		s, ok := n.(*ExprSwitch)
		if !ok || !nilOr(s.Initialization, init) || !nilOr(s.Expression, expr) || len(cases) != len(s.Cases) {
			return false
		}
		for i, c := range cases {
			if len(c.guards) != len(s.Cases[i].Expressions) ||
				len(c.stmts) != len(s.Cases[i].Statements) {
				return false
			}
			for j, m := range c.guards {
				if !m(s.Cases[i].Expressions[j], nil) {
					return false
				}
			}
			for j, m := range c.stmts {
				if !m(s.Cases[i].Statements[j], nil) {
					return false
				}
			}
		}
		return true
	}
}

func typeSwitch(init matcher, decl matcher, expr matcher, cases ...caseMatcher) matcher {
	return func(n Node, err error) bool {
		s, ok := n.(*TypeSwitch)
		if !ok || !nilOr(s.Initialization, init) || !expr(s.Expression, nil) || len(cases) != len(s.Cases) {
			return false
		}
		if (s.Declaration == nil) != (decl == nil) || (decl != nil && !decl(s.Declaration, nil)) {
			return false
		}
		for i, c := range cases {
			if len(c.guards) != len(s.Cases[i].Types) || len(c.stmts) != len(s.Cases[i].Statements) {
				return false
			}
			for j, m := range c.guards {
				if !m(s.Cases[i].Types[j], nil) {
					return false
				}
			}
			for j, m := range c.stmts {
				if !m(s.Cases[i].Statements[j], nil) {
					return false
				}
			}
		}
		return true
	}
}

func forRange(rng matcher, blk matcher) matcher {
	return func(n Node, err error) bool {
		s, ok := n.(*ForStmt)
		return err == nil && ok && s.Initialization == nil && s.Condition == nil && s.Post == nil && rng(s.Range, nil) && blk(&s.Block, nil)
	}
}

func forLoop(init matcher, cond matcher, post matcher, blk matcher) matcher {
	return func(n Node, err error) bool {
		s, ok := n.(*ForStmt)
		return err == nil && ok && s.Range == nil && nilOr(s.Initialization, init) && nilOr(s.Condition, cond) && nilOr(s.Post, post) && blk(&s.Block, nil)
	}
}

func ifStmt(st matcher, cond matcher, blk matcher, els matcher) matcher {
	return func(n Node, err error) bool {
		s, ok := n.(*IfStmt)
		return err == nil && ok && nilOr(s.Statement, st) && cond(s.Condition, nil) && blk(&s.Block, nil) && nilOr(s.Else, els)
	}
}

func labeled(l matcher, st matcher) matcher {
	return func(n Node, err error) bool {
		s, ok := n.(*LabeledStmt)
		return err == nil && ok && l(&s.Label, nil) && st(s.Statement, nil)
	}
}

func block(stmts ...matcher) matcher {
	return func(n Node, err error) bool {
		s, ok := n.(*BlockStmt)
		if err != nil || !ok || len(stmts) != len(s.Statements) {
			return false
		}
		for i, stmt := range stmts {
			if !stmt(s.Statements[i], nil) {
				return false
			}
		}
		return true
	}
}

func deferStmt(e matcher) matcher {
	return func(n Node, err error) bool {
		s, ok := n.(*DeferStmt)
		return err == nil && ok && e(s.Expression, nil)
	}
}

func goStmt(e matcher) matcher {
	return func(n Node, err error) bool {
		s, ok := n.(*GoStmt)
		return err == nil && ok && e(s.Expression, nil)
	}
}

func fallthroughStmt() matcher {
	return func(n Node, err error) bool {
		_, ok := n.(*FallthroughStmt)
		return err == nil && ok
	}
}

func returnStmt(es ...matcher) matcher {
	return func(n Node, err error) bool {
		s, ok := n.(*ReturnStmt)
		if err != nil || !ok || len(es) != len(s.Expressions) {
			return false
		}
		for i, e := range es {
			if !e(s.Expressions[i], nil) {
				return false
			}
		}
		return true
	}
}

func continueStmt(l matcher) matcher {
	return func(n Node, err error) bool {
		s, ok := n.(*ContinueStmt)
		return err == nil && ok && (s.Label == nil && l == nil || s.Label != nil && l != nil && l(s.Label, nil))
	}
}

func breakStmt(l matcher) matcher {
	return func(n Node, err error) bool {
		s, ok := n.(*BreakStmt)
		return err == nil && ok && (s.Label == nil && l == nil || s.Label != nil && l != nil && l(s.Label, nil))
	}
}

func gotoStmt(l matcher) matcher {
	return func(n Node, err error) bool {
		s, ok := n.(*GotoStmt)
		return err == nil && ok && l(&s.Label, nil)
	}
}

func decl(ds ...matcher) matcher {
	return func(n Node, err error) bool {
		s, ok := n.(*DeclarationStmt)
		if err != nil || !ok || len(ds) != len(s.Declarations) {
			return false
		}
		for i, d := range ds {
			if !d(s.Declarations[i], nil) {
				return false
			}
		}
		return true
	}
}

func shortDecl(left []matcher, right ...matcher) matcher {
	return func(n Node, err error) bool {
		s, ok := n.(*ShortVarDecl)
		if err != nil || !ok || len(left) != len(s.Left) || len(right) != len(s.Right) {
			return false
		}
		for i, l := range left {
			if !l(&s.Left[i], nil) {
				return false
			}
		}
		for i, r := range right {
			if !r(s.Right[i], nil) {
				return false
			}
		}
		return true
	}
}

func assign(op token.Token, left []matcher, right ...matcher) matcher {
	return func(n Node, err error) bool {
		s, ok := n.(*Assignment)
		if err != nil || !ok || len(left) != len(s.Left) || len(right) != len(s.Right) || op != s.Op {
			return false
		}
		for i, l := range left {
			if !l(s.Left[i], nil) {
				return false
			}
		}
		for i, r := range right {
			if !r(s.Right[i], nil) {
				return false
			}
		}
		return true
	}
}

func expr(ex matcher) matcher {
	return func(n Node, err error) bool {
		s, ok := n.(*ExpressionStmt)
		return err == nil && ok && ex(s.Expression, nil)
	}
}

func decr(ex matcher) matcher {
	return func(n Node, err error) bool {
		s, ok := n.(*IncDecStmt)
		return err == nil && ok && s.Op == token.MinusMinus && ex(s.Expression, nil)
	}
}

func incr(ex matcher) matcher {
	return func(n Node, err error) bool {
		s, ok := n.(*IncDecStmt)
		return err == nil && ok && s.Op == token.PlusPlus && ex(s.Expression, nil)
	}
}

func send(ch, ex matcher) matcher {
	return func(n Node, err error) bool {
		s, ok := n.(*SendStmt)
		return err == nil && ok && ch(s.Channel, nil) && ex(s.Expression, nil)
	}
}

func empty() matcher {
	return func(n Node, err error) bool {
		return err == nil && n == nil
	}
}

func decls(decls ...matcher) matcher {
	return func(n Node, err error) bool {
		ds, ok := n.(Declarations)
		if err != nil || !ok || len(ds) != len(decls) {
			return false
		}
		for i, d := range ds {
			if !decls[i](d, nil) {
				return false
			}
		}
		return true
	}
}

func vars(names []matcher, typ matcher, vals ...matcher) matcher {
	return func(n Node, err error) bool {
		vs, ok := n.(*VarSpec)
		if err != nil || !ok || !nilOr(vs.Type, typ) || len(names) != len(vs.Names) || len(vals) != len(vs.Values) {
			return false
		}
		for i, n := range vs.Names {
			if !names[i](&n, nil) {
				return false
			}
		}
		for i, v := range vs.Values {
			if !vals[i](v, nil) {
				return false
			}
		}
		return true
	}
}

func cnst(names []matcher, typ matcher, vals ...matcher) matcher {
	return func(n Node, err error) bool {
		cs, ok := n.(*ConstSpec)
		if err != nil || !ok || !nilOr(cs.Type, typ) || len(names) != len(cs.Names) || len(vals) != len(cs.Values) {
			return false
		}
		for i, n := range cs.Names {
			if !names[i](&n, nil) {
				return false
			}
		}
		for i, v := range cs.Values {
			if !vals[i](v, nil) {
				return false
			}
		}
		return true
	}
}

func typ(name matcher, typ matcher) matcher {
	return func(n Node, err error) bool {
		t, ok := n.(*TypeSpec)
		return err == nil && ok && name(&t.Name, nil) && typ(t.Type, nil)
	}
}

func structType(fields ...matcher) matcher {
	return func(n Node, err error) bool {
		s, ok := n.(*StructType)
		if err != nil || !ok || len(s.Fields) != len(fields) {
			return false
		}
		for i, f := range s.Fields {
			if !fields[i](&f, nil) {
				return false
			}
		}
		return true
	}
}

func fieldDecl(typ matcher, idents ...matcher) matcher {
	return func(n Node, err error) bool {
		f, ok := n.(*FieldDecl)
		if err != nil || !ok || !typ(f.Type, nil) || len(idents) != len(f.Identifiers) {
			return false
		}
		for i, id := range f.Identifiers {
			if !idents[i](&id, nil) {
				return false
			}
		}
		return true
	}
}

func tagFieldDecl(typ matcher, tag matcher, idents ...matcher) matcher {
	return func(n Node, err error) bool {
		if !fieldDecl(typ, idents...)(n, err) {
			return false
		}
		t := n.(*FieldDecl).Tag
		return t != nil && tag(t, nil)
	}
}

func ifaceType(methods ...matcher) matcher {
	return func(n Node, err error) bool {
		i, ok := n.(*InterfaceType)
		if err != nil || !ok || len(methods) != len(i.Methods) {
			return false
		}
		for i, meth := range i.Methods {
			if !methods[i](meth, nil) {
				return false
			}
		}
		return true
	}
}

func method(name matcher, parm matcher, res matcher) matcher {
	return func(n Node, err error) bool {
		m, ok := n.(*Method)
		return err == nil && ok && name(&m.Name, nil) && parm(&m.Parameters, nil) && res(&m.Result, nil)
	}
}

func funcType(parmList matcher, res matcher) matcher {
	return func(n Node, err error) bool {
		f, ok := n.(*FunctionType)
		if err != nil || !ok || !res(&f.Result, nil) {
			return false
		}
		return parmList(&f.Parameters, nil)
	}
}

func parmList(decls ...matcher) matcher {
	return func(n Node, err error) bool {
		l, ok := n.(*ParameterList)
		if err != nil || !ok || len(decls) != len(l.Parameters) {
			return false
		}
		for i, d := range l.Parameters {
			if !decls[i](&d, nil) {
				return false
			}
		}
		return true
	}
}

func parmDecl(typ matcher, dotDotDot bool, idents ...matcher) matcher {
	return func(n Node, err error) bool {
		d, ok := n.(*ParameterDecl)
		if err != nil || !ok || len(idents) != len(d.Identifiers) || dotDotDot != d.DotDotDot || !typ(d.Type, nil) {
			return false
		}
		for i, id := range d.Identifiers {
			if !idents[i](&id, nil) {
				return false
			}
		}
		return true
	}
}

func recvChan(typ matcher) matcher {
	return func(n Node, err error) bool {
		c, ok := n.(*ChannelType)
		return err == nil && ok && typ(c.Type, nil) && !c.Send && c.Receive
	}
}

func sendChan(typ matcher) matcher {
	return func(n Node, err error) bool {
		c, ok := n.(*ChannelType)
		return err == nil && ok && typ(c.Type, nil) && c.Send && !c.Receive
	}
}

func chanType(typ matcher) matcher {
	return func(n Node, err error) bool {
		c, ok := n.(*ChannelType)
		return err == nil && ok && typ(c.Type, nil) && c.Send && c.Receive
	}
}

func mapType(key, typ matcher) matcher {
	return func(n Node, err error) bool {
		m, ok := n.(*MapType)
		return err == nil && ok && key(m.Key, nil) && typ(m.Type, nil)
	}
}

func arrayType(size, typ matcher) matcher {
	return func(n Node, err error) bool {
		a, ok := n.(*ArrayType)
		return err == nil && ok && nilOr(a.Size, size) && typ(a.Type, nil)
	}
}

func sliceType(typ matcher) matcher {
	return func(n Node, err error) bool {
		s, ok := n.(*SliceType)
		return err == nil && ok && typ(s.Type, nil)
	}
}

func star(typ matcher) matcher {
	return func(n Node, err error) bool {
		p, ok := n.(*Star)
		return err == nil && ok && typ(p.Target, nil)
	}
}

func compLit(typ matcher, elms ...matcher) matcher {
	return func(n Node, err error) bool {
		c, ok := n.(*CompositeLiteral)
		if err != nil || !ok || !nilOr(c.Type, typ) || len(elms) != len(c.Elements) {
			return false
		}
		for i, e := range c.Elements {
			if !elms[i](&e, nil) {
				return false
			}
		}
		return true
	}
}

func litVal(elms ...matcher) matcher {
	return func(n Node, err error) bool {
		c, ok := n.(*CompositeLiteral)
		if err != nil || !ok || c.Type != nil || len(elms) != len(c.Elements) {
			return false
		}
		for i, e := range c.Elements {
			if !elms[i](&e, nil) {
				return false
			}
		}
		return true
	}
}

func elm(key matcher, val matcher) matcher {
	return func(n Node, err error) bool {
		e, ok := n.(*Element)
		return err == nil && ok && nilOr(e.Key, key) && val(e.Value, nil)
	}
}

func slice(expr, low, high, max matcher) matcher {
	return func(n Node, err error) bool {
		e, ok := n.(*Slice)
		return err == nil && ok && expr(e.Expression, nil) &&
			nilOr(e.Low, low) &&
			nilOr(e.High, high) &&
			nilOr(e.Max, max)
	}
}

func index(expr, idx matcher) matcher {
	return func(n Node, err error) bool {
		e, ok := n.(*Index)
		return err == nil && ok && expr(e.Expression, nil) && idx(e.Index, nil)
	}
}

func tAssert(expr, typ matcher) matcher {
	return func(n Node, err error) bool {
		t, ok := n.(*TypeAssertion)
		return err == nil && ok && expr(t.Expression, nil) && nilOr(t.Type, typ)
	}
}

func sel(expr, sele matcher) matcher {
	return func(n Node, err error) bool {
		s, ok := n.(*Selector)
		return err == nil && ok && expr(s.Parent, nil) && sele(s.Name, nil)
	}
}

func call(fun matcher, ddd bool, args ...matcher) matcher {
	return func(n Node, err error) bool {
		c, ok := n.(*Call)
		if err != nil || !ok || !fun(c.Function, nil) || len(args) != len(c.Arguments) || c.DotDotDot != ddd {
			return false
		}
		for i, a := range c.Arguments {
			if !args[i](a, nil) {
				return false
			}
		}
		return true
	}
}

func unOp(op token.Token, operand matcher) matcher {
	return func(n Node, err error) bool {
		u, ok := n.(*UnaryOp)
		return err == nil && ok && u.Op == op && operand(u.Operand, nil)
	}
}

func binOp(op token.Token, left, right matcher) matcher {
	return func(n Node, err error) bool {
		b, ok := n.(*BinaryOp)
		return err == nil && ok && b.Op == op && left(b.Left, nil) && right(b.Right, nil)
	}
}

func ident(name string) matcher {
	return func(n Node, err error) bool {
		id, ok := n.(*Identifier)
		return err == nil && ok && id.Name == name
	}
}

func intLit(str string) matcher {
	return func(n Node, err error) bool {
		i, ok := n.(*IntegerLiteral)
		return err == nil && ok && i.Value.String() == str
	}
}

func floatLit(str string) matcher {
	return func(n Node, err error) bool {
		f, ok := n.(*FloatLiteral)
		return err == nil && ok && f.printString() == str
	}
}

func imaginaryLit(str string) matcher {
	return func(n Node, err error) bool {
		i, ok := n.(*ImaginaryLiteral)
		return err == nil && ok && i.printString() == str
	}
}

func strLit(str string) matcher {
	return func(n Node, err error) bool {
		s, ok := n.(*StringLiteral)
		ok = err == nil && ok && s.Value == str
		if !ok {
			fmt.Printf("[%s] != [%s]\n", s.Value, str)
		}
		return ok
	}
}

func runeLit(ru rune) matcher {
	return func(n Node, err error) bool {
		r, ok := n.(*RuneLiteral)
		return err == nil && ok && r.Value == ru
	}
}

func (n *FloatLiteral) printString() string {
	return ratPrintString(n.Value)
}

func (n *ImaginaryLiteral) printString() string {
	return ratPrintString(n.Value) + "i"
}

// PrintFloatPrecision is the number of digits of precision to use when
// printing floating point values.  Trailing zeroes are trimmed, so
// using a large value won't necessarily lead to extremely large strings.
const printFloatPrecision = 20

// RatPrintString returns a printable string representation of a big.Rat.
func ratPrintString(rat *big.Rat) string {
	s := rat.FloatString(printFloatPrecision)
	s = strings.TrimRight(s, "0")
	if len(s) > 0 && s[len(s)-1] == '.' {
		s += "0"
	}
	return s
}
