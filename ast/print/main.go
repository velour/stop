// Print parses a source file and pretty-prints the syntax tree.
package main

import (
	"bufio"
	"flag"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"

	"github.com/mccoyst/pipeline"
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
		if err = ast.Print(out, root); err != nil {
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

	pl, err := pipeline.New(
		exec.Command("tee", "out.dot"),
		exec.Command("dot", "-o"+ps, "-Tps"),
		exec.Command("gv", ps),
	)
	if err != nil {
		die(err)
	}
	out, err := pl.First().StdinPipe()
	if err != nil {
		die(err)
	}
	go func() {
		if err := ast.Dot(out, root); err != nil {
			die(err)
		}
		out.Close()
	}()
	pl.Start()
	es := pl.Wait()
	if len(es) > 0 {
		die(errs(es))
	}
}

type errs []error

func (es errs) Error() string {
	s := ""
	for _, e := range es {
		s += e.Error() + "\n"
	}
	return strings.TrimSpace(s)
}

func die(err error) {
	str := err.Error()
	if se, ok := err.(*ast.SyntaxError); *v && ok {
		str += "\n" + se.Stack
	}
	os.Stdout.WriteString(str + "\n")
	os.Exit(1)
}
