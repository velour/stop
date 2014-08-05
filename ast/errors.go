package ast

import (
	"fmt"
	"strconv"

	"github.com/velour/stop/token"
)

// Errors defines a slice of errors. The Error method returns the concatination of the
// Error strings of all errors, separated by newlines.
type errors []error

func (es errors) Error() string {
	s := ""
	for _, e := range es {
		if s != "" {
			s += "\n"
		}
		s += e.Error()
	}
	return s
}

// Errs returns an errors from a sequence of errors.
func errs(es ...error) errors {
	return errors(es)
}

// ErrorOrNil returns nil if the errors is empty or it returns the errors as an error.
func (es errors) ErrorOrNil() error {
	if len(es) == 0 {
		return nil
	}
	return es
}

// All returns all errors, gathered recursively by calling All on any nested errors.
func (es errors) All() []error {
	var all []error
	for _, e := range es {
		if es, ok := e.(errors); ok {
			all = append(all, es.All()...)
		} else {
			all = append(all, e)
		}
	}
	return all
}

// A SyntaxError is an error that describes a parse failure: something
// unexpected in the syntax of the Go source code.
type SyntaxError struct {
	// Wanted describes the value that was expected.
	Wanted string
	// Got describes the unexpected token.
	Got fmt.Stringer
	// Text is the text of the unexpected token.
	Text string
	// Start and End give the location of the error.
	Start, End token.Location
	// Stack is a human-readable trace of the stack showing the
	// cause of the syntax error.
	Stack string
}

func (e *SyntaxError) Error() string {
	if e.Got == token.Error {
		text := strconv.QuoteToASCII(e.Text)
		return fmt.Sprintf("%s: unexpected rune in input [%s]", e.Start, text[1:len(text)-1])
	}
	switch e.Got {
	case token.Semicolon:
		e.Text = ";"
	case token.EOF:
		e.Text = "EOF"
	}
	return fmt.Sprintf("%s: expected %s, got %s", e.Start, e.Wanted, e.Text)
}

// A MalformedLiteral is an error that describes a literal for which the
// value was malformed.  For example, an integer literal for which the
// text is not parsable as an integer.
//
// BUG(eaburns): The parser should not have to deal with this class
// of error.  The lexer should return an Error token if it scans a literal
// that is lexicographically incorrect.
type MalformedLiteral struct {
	// The name for the literal type.
	Type string
	// The malformed text.
	Text string
	// Start and End give the location of the error.
	Start, End token.Location
}

func (e *MalformedLiteral) Error() string {
	return fmt.Sprintf("%s: malformed %s: [%s]", e.Start, e.Type, e.Text)
}

// Redeclaration is an error that denotes multiple definitions of the same
// variable within the same scope.
type Redeclaration struct {
	Name          string
	First, Second Declaration
}

func (e *Redeclaration) Error() string {
	return fmt.Sprintf("%s: %s redeclared, originally declared at %s", e.Second.Start(), e.Name, e.First.Start())
}

// An Undeclared is an error returned for undeclared identifiers.
type Undeclared struct{ *Identifier }

func (e Undeclared) Error() string {
	return fmt.Sprintf("%s: undeclared identifier %s", e.Start(), e.Name)
}

// A ConstantLoop is an error returned when there is a cycle in a constant definition.
type ConstantLoop struct{ *constSpecView }

func (e ConstantLoop) Error() string {
	return fmt.Sprintf("%s: constant definition loop", e.Start())
}

// A NotConstant is an error returned when a constant initializer is not constant.
type NotConstant struct{ Expression }

func (e NotConstant) Error() string {
	return fmt.Sprintf("%s: const initializer is not constant", e.Start())
}

// A BadConstAssign is an error returned when a constant operand has a value
// that is not representable by the type to which it is being assigned.
type BadConstAssign struct {
	Expression
	Type
}

func (e BadConstAssign) Error() string {
	return fmt.Sprintf("%s: bad constant assignment", e.Expression.Start())
}

// A BadAssign is an error returned when an expression is not assignable
// to the type to which it is being assigned.
type BadAssign struct {
	Expression
	Type
}

func (e BadAssign) Error() string {
	return fmt.Sprintf("%s: bad assignment", e.Expression.Start())
}

// A AssignCountMismatch is an error returned when a variable or contstant
// assignment has differing numbers of identifiers as it has expressions
// being assigned.
type AssignCountMismatch struct {
	Declaration
}

func (e AssignCountMismatch) Error() string {
	return fmt.Sprintf("%s: assignment count mismatch", e.Start())
}
