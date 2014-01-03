package main

import (
	"fmt"
	"os"

	"bitbucket.org/eaburns/stop/lexer"
)

func main() {
	l := lexer.New(os.Stdin)
	for {
		tok := l.Next()
		fmt.Printf("%+v\n", tok)
		if tok.Type == lexer.EOF {
			break
		}
	}
}
