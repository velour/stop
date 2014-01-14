package token

import "strconv"

// A Location is the address of a particular rune in an input stream.
type Location struct {
	// Path is the path to the file for this location or "".
	Path string
	// Rune is the rune offset within the input stream.
	Rune int
	// Line is the line number within the input stream.
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
	line := strconv.Itoa(l.Line)
	col := strconv.Itoa(l.Column())
	return l.Path + ":" + line + "-+#" + col
}
