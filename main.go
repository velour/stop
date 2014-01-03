package main

import (
	"fmt"
	"os"

	"bitbucket.org/eaburns/stop/lexer"
)

func main() {
	f, err := os.Open(os.Args[1])
	if err != nil {
		panic(err)
	}
	l := lexer.New(f)
	for {
		tok := l.Next()
		fmt.Printf("%+v\n", tok)
		if tok.Type == lexer.TokenEOF {
			break
		}
	}
}
