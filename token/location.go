package token

import "strconv"

// A Location is the address of a rune within the input text.
type Location struct {
	// Path is the path to the file containing the input.
	Path string
	// Rune is the rune offset.
	Rune int
	// Line is the line number.
	Line int
	// LineStart is the rune offset of the first rune on this line.
	LineStart int
}

// Column returns the rune offset into the line of a location.
func (l Location) Column() int {
	return l.Rune - l.LineStart
}

// String returns the string representation of a location as an
// Acme address.
func (l Location) String() string {
	return l.Path + ":" + strconv.Itoa(l.Line) + ":" + strconv.Itoa(l.Column())
}
