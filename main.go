package main

import (
	"fmt"
	"os"

	"bitbucket.org/eaburns/stop/token"
)

func main() {
	l := lexer.New(os.Stdin)
	for {
		tok := token.NewScanner()
		fmt.Printf("%+v\n", tok)
		if tok.Type == token.EOF {
			break
		}
	}
}
