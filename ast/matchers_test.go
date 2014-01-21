package ast

import (
	"bytes"
	"fmt"
	"regexp"
	"testing"

	"bitbucket.org/eaburns/stop/token"
)

// NodeString returns a human-readable string representation of the AST node.
func nodeString(n Node) string {
	out := bytes.NewBuffer(nil)
	n.print(0, out)
	return string(out.Bytes())
}

// A matcher function is used for testing the expected structure of ASTs.  The
// matcher returns true if the tree matches the expected structure of if there
// is an expected error.
type matcher func(Node, error) bool

type parserTests []struct {
	src   string
	match matcher
}

func (tests parserTests) run(t *testing.T) {
	for i, test := range tests {
		l := token.NewLexer("", test.src)
		p := NewParser(l)
		n, err := Parse(p)
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

func parseErr(reStr string) matcher {
	re := regexp.MustCompile(reStr)
	return func(_ Node, err error) bool {
		return err != nil && re.MatchString(err.Error())
	}
}

func slice(expr, low, high, max matcher) matcher {
	return func(n Node, err error) bool {
		e, ok := n.(*Slice)
		return err == nil && ok && expr(e.Expression, nil) &&
			(low == nil && e.Low == nil || low(e.Low, nil)) &&
			(high == nil && e.High == nil || high(e.High, nil)) &&
			(max == nil && e.Max == nil || max(e.Max, nil))
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
