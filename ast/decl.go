package ast

import (
	"github.com/velour/stop/token"
)

var (
	uint8Decl predeclared
	int32Decl predeclared

	// UnivScope is the universal symtab, containing all predeclared identifiers.
	univScope = symtab{
		Decls: map[string]Declaration{
			// Predeclared types.
			"bool":       &predeclared{},
			"byte":       &uint8Decl,
			"complex64":  &predeclared{},
			"complex128": &predeclared{},
			"error":      &predeclared{},
			"float32":    &predeclared{},
			"float64":    &predeclared{},
			"int":        &predeclared{},
			"int8":       &predeclared{},
			"int16":      &predeclared{},
			"int32":      &int32Decl,
			"int64":      &predeclared{},
			"rune":       &int32Decl,
			"string":     &predeclared{},
			"uint":       &predeclared{},
			"uint8":      &uint8Decl,
			"uint16":     &predeclared{},
			"uint32":     &predeclared{},
			"uint64":     &predeclared{},
			"uintptr":    &predeclared{},

			// Predeclared constants.
			"true":  &predeclared{},
			"false": &predeclared{},
			"iota":  &predeclared{},

			// Predeclared zero value.
			"nil": &predeclared{},

			// Predeclared functions.
			"append":  &predeclared{},
			"cap":     &predeclared{},
			"close":   &predeclared{},
			"complex": &predeclared{},
			"copy":    &predeclared{},
			"delete":  &predeclared{},
			"imag":    &predeclared{},
			"len":     &predeclared{},
			"make":    &predeclared{},
			"new":     &predeclared{},
			"panic":   &predeclared{},
			"print":   &predeclared{},
			"println": &predeclared{},
			"real":    &predeclared{},
			"recover": &predeclared{},
		},
	}
)

// A predeclared is a declaration node representing a predeclared identifier.
type predeclared struct{}

func (n *predeclared) Comments() []string    { return nil }
func (n *predeclared) Start() token.Location { return token.Location{} }
func (n *predeclared) End() token.Location   { return token.Location{} }

// A symtab is the main element of the symbol table. It contains a mapping
// from all identifiers in a given scope to their declaration. The declarations
// are unique.  Each identifier declared in a VarSpec or a ConstSpec is
// mapped to a unique view of the declaring spec.
type symtab struct {
	Up    *symtab
	Decls map[string]Declaration
}

// Find returns the declaration bound to the given identifier, or nil if the identifier
// is not found.
func (s *symtab) Find(n string) Declaration {
	if s == nil {
		return nil
	}
	if d, ok := s.Decls[n]; ok {
		return d
	}
	return s.Up.Find(n)
}

// Bind binds a name to its declaration, returning an error if the name is already
// bound in this symtab. The blank identifier is not bound.
func (s *symtab) Bind(n string, decl Declaration) error {
	if n == "_" {
		return nil
	}
	if d, ok := s.Decls[n]; ok {
		return &Redeclaration{Name: n, First: d, Second: decl}
	}
	s.Decls[n] = decl
	return nil
}

// A constSpecView is a view of a ConstSpec that focuses on a single identifier
// at a given index.
type constSpecView struct {
	Index int
	*ConstSpec
}

// Value returns the bound expression or nil if there isn't one.
func (c constSpecView) Value() Expression {
	if c.Index < len(c.Values) {
		return c.Values[c.Index]
	}
	return nil
}

// A varSpecView is a view of a VarSpec that focuses on a single identifier
// at a given index.
type varSpecView struct {
	Index int
	*VarSpec
}

// PkgDecls returns a symtab, mapping from package-scoped identifiers
// to their corresponding declarations. Each identifier is mapped to a
// unique declaration. Each identifier declared in a VarSpec or a
// ConstSpec is mapped to a unique view of the declaring spec.
func pkgDecls(files []*SourceFile) (*symtab, error) {
	s := &symtab{Up: &univScope, Decls: make(map[string]Declaration)}
	var errs errors
	for _, f := range files {
		for _, d := range f.Declarations {
			switch d := d.(type) {
			case *MethodDecl:
				errs.Add(s.Bind(d.Identifier.Name, d))
			case *FunctionDecl:
				errs.Add(s.Bind(d.Identifier.Name, d))
			case *TypeSpec:
				errs.Add(s.Bind(d.Identifier.Name, d))
			case *ConstSpec:
				for i := range d.Identifiers {
					n := d.Identifiers[i].Name
					v := &constSpecView{Index: i, ConstSpec: d}
					errs.Add(s.Bind(n, v))
				}
			case *VarSpec:
				for i := range d.Identifiers {
					n := d.Identifiers[i].Name
					v := &varSpecView{Index: i, VarSpec: d}
					errs.Add(s.Bind(n, v))
				}
			default:
				panic("invalid top-level declaration")
			}
		}
	}
	return s, errs.ErrorOrNil()
}
