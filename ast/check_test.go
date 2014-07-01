package ast

import (
	"reflect"
	"sort"
	"testing"

	"github.com/eaburns/pp"
	"github.com/velour/stop/token"
)

var (
	testSource = `
		package testpkg
		const (
			π = 3.1415926535
			name = "eaburns"
		)
		var mayChange int = f()
		var len = 1	// Shadows the predeclared len function.

		func f() int { return 0 }
	`

	testScope = pkgScope([]*SourceFile{
		parseSourceFile(NewParser(token.NewLexer("", testSource))),
	})
)

var pkgDeclsTests = []struct {
	srcFiles []string
	// The top-level identifiers.
	ids []string
}{
	{
		[]string{
			`package a; const a = 1`,
			`package a; const B = 1`,
		},
		[]string{"a", "B"},
	},
	{
		[]string{`package a; const a, B = B, 1`},
		[]string{"a", "B"},
	},
	{
		[]string{
			`package a
			const (
				a, B = B, 1
				c, d
			)`,
		},
		[]string{"a", "B", "c", "d"},
	},
	{
		[]string{
			`package a; var a = 1`,
			`package a; var B = 1`,
		},
		[]string{"a", "B"},
	},
	{
		[]string{`package a; var a, B = 1, 2`},
		[]string{"a", "B"},
	},
	{
		[]string{
			`package a
			var (
				a, B = 0, 1
				c, d = 3.0, 4.0
			)`,
		},
		[]string{"a", "B", "c", "d"},
	},
	{
		[]string{
			`package a
			type (
				a struct{}
				B struct { C }
			)`,
		},
		[]string{"a", "B"},
	},
	{
		[]string{
			`package a
			func a(){}
			func B() int { return 0 }`,
		},
		[]string{"a", "B"},
	},
	{
		[]string{
			`package a
			func a(){}
			func B() int { return 0 }`,
			`package a
			func c(e int){}
			func d() (f int) { return 0 }`,
		},
		[]string{"a", "B", "c", "d"},
	},
	{
		[]string{
			`package a
			func (z int) a(){}
			func (z float64) B() int { return 0 }`,
		},
		[]string{"a", "B"},
	},
	{
		[]string{
			`package a
			func (z T0) a(){}
			func (z T1) B() int { return 0 }`,
			`package a
			func (z T2) c(e int){}
			func (z T3) d() (f int) { return 0 }`,
		},
		[]string{"a", "B", "c", "d"},
	},
}

func TestPkgDecls(t *testing.T) {
	for _, test := range pkgDeclsTests {
		files := parseSrcFiles(t, test.srcFiles)
		if files == nil {
			continue
		}
		var ids []string
		seen := make(map[Declaration]bool)
		for id, d := range pkgScope(files).Decls {
			ids = append(ids, id)
			if seen[d] {
				t.Errorf("multiple idents map to declaration %s", pp.MustString(d))
			}
			seen[d] = true
		}
		sort.Strings(ids)
		sort.Strings(test.ids)
		if reflect.DeepEqual(ids, test.ids) {
			continue
		}
		t.Errorf("pkgDecls(%s)=%v, want %v", pp.MustString(test.srcFiles), ids, test.ids)
	}
}

func parseSrcFiles(t *testing.T, srcFiles []string) []*SourceFile {
	var files []*SourceFile
	for _, src := range srcFiles {
		p := NewParser(token.NewLexer("", src))
		file, err := Parse(p)
		if err == nil {
			files = append(files, file)
			continue
		}
		t.Fatalf("Parse(%s), got unexpected error %v", src, err)
		return nil
	}
	return files
}

func TestScopeFind(t *testing.T) {
	src := `
		package testpkg
		const (
			π = 3.1415926535
			name = "eaburns"
		)
		var mayChange int = f()
		var len = 1	// Shadows the predeclared len function.
		func f() int { return 0 }
	`
	p := NewParser(token.NewLexer("", src))
	scope := pkgScope([]*SourceFile{parseSourceFile(p)})

	tests := []struct {
		id       string
		declType reflect.Type
	}{
		{"π", reflect.TypeOf(&constSpecView{})},
		{"name", reflect.TypeOf(&constSpecView{})},
		{"mayChange", reflect.TypeOf(&varSpecView{})},
		{"len", reflect.TypeOf(&varSpecView{})},
		{"f", reflect.TypeOf(&FunctionDecl{})},
		{"notDefined", nil},
	}
	for _, test := range tests {
		d := scope.Find(test.id)
		typ := reflect.TypeOf(d)
		if (d == nil && test.declType != nil) || (d != nil && typ != test.declType) {
			t.Errorf("Find(%s)=%s, want %s", test.id, typ, test.declType)
		}
	}

	// Our testing source file shadowed len, but let's check that it's
	// still findable in the universal scope.
	d := univScope.Find("len")
	typ := reflect.TypeOf(d)
	if d == nil || typ != reflect.TypeOf(&predeclared{}) {
		t.Errorf("Find(len)=%s, want %s", typ, reflect.TypeOf(&predeclared{}))
	}

	// Rune and byte are aliases for int32 and uint8 respectively;
	// the have the same declaration.
	if univScope.Find("rune") != univScope.Find("int32") {
		t.Errorf("rune is not an alias for int32")
	}
	if univScope.Find("byte") != univScope.Find("uint8") {
		t.Errorf("byte is not an alias for uint8")
	}
}
