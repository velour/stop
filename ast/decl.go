package ast

import (
	"github.com/velour/stop/token"
)

var (
	uint8Decl predeclared
	int32Decl predeclared

	// UnivScope is the universal scope, containing all predeclared identifiers.
	univScope = scope{
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

// A scope is the main element of the symbol table. It contains a mapping
// from all identifiers in a given scope to their declaration. The declarations
// are unique.  Each identifier declared in a VarSpec or a ConstSpec is
// mapped to a unique view of the declaring spec.
type scope struct {
	Up    *scope
	Decls map[string]Declaration
}

// Find returns the declaration bound to the given identifier, or nil if the identifier
// is not found.
func (s *scope) Find(id string) Declaration {
	if s == nil {
		return nil
	}
	if d, ok := s.Decls[id]; ok {
		return d
	}
	return s.Up.Find(id)
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

// A predeclared is a declaration node representing a predeclared identifier.
type predeclared struct{}

func (n *predeclared) Comments() []string    { return nil }
func (n *predeclared) Start() token.Location { return token.Location{} }
func (n *predeclared) End() token.Location   { return token.Location{} }

// PkgScope returns a scope, mapping from package-scoped identifiers
// to their corresponding declarations. Each identifier is mapped to a
// unique declaration. Each identifier declared in a VarSpec or a
// ConstSpec is mapped to a unique view of the declaring spec.
func pkgScope(files []*SourceFile) *scope {
	ds := make(map[string]Declaration)
	for _, f := range files {
		for _, d := range f.Declarations {
			switch d := d.(type) {
			case *MethodDecl:
				ds[d.Identifier.Name] = d
			case *FunctionDecl:
				ds[d.Identifier.Name] = d
			case *TypeSpec:
				ds[d.Identifier.Name] = d
			case *ConstSpec:
				for i := range d.Identifiers {
					n := d.Identifiers[i].Name
					ds[n] = &constSpecView{Index: i, ConstSpec: d}
				}
			case *VarSpec:
				for i := range d.Identifiers {
					n := d.Identifiers[i].Name
					ds[n] = &varSpecView{Index: i, VarSpec: d}
				}
			default:
				panic("invalid top-level declaration")
			}
		}
	}
	return &scope{Up: &univScope, Decls: ds}
}
