// Package loc provides types for representing the location of items
// in an input stream.
package loc

import (
	"strconv"
)

// A Location is the address of a particular rune in an input stream.
type Location struct {
	// Path is the path to the file for this location, or "" if there is no file.
	Path string
	// Rune is the rune number of this location within the input stream.
	Rune int
	// Line is the line number of this location.
	Line int
	// LineStart is the rune number of the first rune on the line of this location.
	LineStart int
}

// RuneOnLine returns the rune offset into the line of a location.
func (l Location) RuneOnLine() int {
	return l.Rune - l.LineStart
}

// String returns the string representation of a location as an Acme address.
func (l Location) String() string {
	return l.Path + ":" + strconv.Itoa(l.Line) + "-+#" + strconv.Itoa(l.RuneOnLine())
}

// Zero returns the beginning, or zero, location for a path.
func Zero(path string) Location {
	return Location{Path: path, Line: 1, Rune: 1, LineStart: 1}
}
