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
	return (n == nil && m == nil) || (n != nil && m != nil && m(n, nil))
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
		return parseExpression(p)
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

func pointer(typ matcher) matcher {
	return func(n Node, err error) bool {
		p, ok := n.(*PointerType)
		return err == nil && ok && typ(p.Type, nil)
	}
}

func typeName(pkg, name string) matcher {
	return func(n Node, err error) bool {
		t, ok := n.(*TypeName)
		return err == nil && ok && t.Package == pkg && t.Name == name
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
		return err == nil && ok && expr(t.Expression, nil) && typ(t.Type, nil)
	}
}

func sel(expr, sele matcher) matcher {
	return func(n Node, err error) bool {
		s, ok := n.(*Selector)
		return err == nil && ok && expr(s.Expression, nil) && sele(s.Selection, nil)
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
