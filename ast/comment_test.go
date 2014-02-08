package ast

import (
	"reflect"
	"testing"

	"github.com/velour/stop/token"
)

type commentTests []struct {
	src string
	// Cmnts are the comments that should appear before each
	// subsequent identifier in the source.
	cmnts [][]string
}

func (tests commentTests) run(t *testing.T) {
	for i, test := range tests {
		p := NewParser(token.NewLexer("", test.src))
		n := 0
		for p.tok != token.EOF {
			if p.tok != token.Identifier {
				goto next
			}
			n++
			if n > len(test.cmnts) {
				break
			}
			// Test lengths, because nil != []string{}.
			if len(p.cmnts) == 0 && len(test.cmnts[n-1]) == 0 {
				goto next
			}
			if reflect.DeepEqual(p.cmnts, test.cmnts[n-1]) {
				goto next
			}
			t.Errorf("test %d: expected comments %v, got %v", i, test.cmnts[n-1], p.cmnts)
		next:
			p.next()
		}
		if n != len(test.cmnts) {
			t.Fatalf("test %d: expected %d identifiers, got %d", i, len(test.cmnts), n)
		}
	}
}

func TestComments(t *testing.T) {
	tests := commentTests{
		{`a`, [][]string{{}}},
		{"// a\na", [][]string{{"// a"}}},
		{"/* a */a", [][]string{{"/* a */"}}},
		{"/* a */\na", [][]string{{"/* a */"}}},
		{"// a\n// b\na", [][]string{{"// a", "// b"}}},
		{"// a\n/* b */\na", [][]string{{"// a", "/* b */"}}},
		{"// a\na\n// b\nb", [][]string{{"// a"}, {"// b"}}},
		{"a\n// b\nb", [][]string{{}, {"// b"}}},
		{"// a\na\nb", [][]string{{"// a"}, {}}},
		{"// a\na\n// b\n\nb", [][]string{{"// a"}, {}}},
		{"// a\na\n// b\n\n// c\nb", [][]string{{"// a"}, {"// c"}}},
	}
	tests.run(t)
}
