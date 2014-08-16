// Print parses a source file and pretty-prints the syntax tree.
package main

import (
	"bufio"
	"flag"
	"io/ioutil"
	"os"
	"os/exec"

	"github.com/eaburns/pp"
	"github.com/eaburns/pretty"
	"github.com/velour/stop/ast"
	"github.com/velour/stop/token"
)

var (
	gv = flag.Bool("gv", false, "displays the tree using graphviz+postscript+gv")
	v  = flag.Bool("v", false, "display verbose parse errors")
)

func main() {
	flag.Parse()

	in, err := os.Open(flag.Arg(0))
	if err != nil {
		die(err)
	}
	defer in.Close()

	src, err := ioutil.ReadAll(in)
	if err != nil {
		die(err)
	}

	l := token.NewLexer(in.Name(), string(src))
	p := ast.NewParser(l)
	root, err := ast.Parse(p)
	if err != nil {
		die(err)
	}

	if *gv {
		dot(root)
	} else {
		out := bufio.NewWriter(os.Stdout)
		if err = pretty.Fprint(out, root); err != nil {
			die(err)
		}
		out.Flush()
	}
}

func dot(root ast.Node) {
	tmp, err := ioutil.TempFile("", "stop-dot-")
	if err != nil {
		die(err)
	}
	ps := tmp.Name()
	tmp.Close()
	defer os.Remove(ps)

	dotCmd := exec.Command("dot", "-o"+ps, "-Tps")
	out, err := dotCmd.StdinPipe()
	if err != nil {
		die(err)
	}
	go func() {
		if err := pp.Dot(out, root); err != nil {
			die(err)
		}
		out.Close()
	}()
	if err := dotCmd.Run(); err != nil {
		die(err)
	}
	if err := exec.Command("gv", ps).Run(); err != nil {
		die(err)
	}
}

func die(err error) {
	str := err.Error()
	if se, ok := err.(*ast.SyntaxError); *v && ok {
		str += "\n" + se.Stack
	}
	os.Stdout.WriteString(str + "\n")
	os.Exit(1)
}
