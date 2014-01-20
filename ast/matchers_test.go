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

func opName(pkg, name string) matcher {
	return func(n Node, err error) bool {
		s, ok := n.(*OperandName)
		return err == nil && ok && s.Package == pkg && s.Name == name
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
