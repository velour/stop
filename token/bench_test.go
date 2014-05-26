package token

import (
	"go/scanner"
	"go/token"
	"testing"

	"github.com/velour/stop/test"
)

func BenchmarkLexer(b *testing.B) {
	for i := 0; i < b.N; i++ {
		lex := NewLexer("", test.Prog)
		tok := Semicolon
		for tok != EOF && tok != Error {
			tok = lex.Next()
		}
		if tok == Error {
			b.Fatalf("lex error")
		}
	}
}

// Benchmarks the lexer from the standard library for comparison.
func BenchmarkStandardLibraryLexer(b *testing.B) {
	var lex scanner.Scanner
	src := []byte(test.Prog)
	fileSet := token.NewFileSet()
	file := fileSet.AddFile("", fileSet.Base(), len(src))
	for i := 0; i < b.N; i++ {
		lex.Init(file, src, nil, scanner.ScanComments)
		tok := token.ILLEGAL
		for tok != token.EOF {
			_, tok, _ = lex.Scan()
		}
	}
}
