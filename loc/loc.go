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

// A Span specifies a range of runes in an input stream by a start and end location.
// The end location is exclusive.
type Span [2]Location

// String returns the string representation of a span as an Acme address (it's a bit ugly, but often convenient).
// Spans that cross file boundaries, will not result in valid addresses.
func (s Span) String() string {
	switch {
	case s[0].Path != s[1].Path:
		return s[0].String() + "-" + s[1].String()
	case s[0] != s[1]:
		l0 := strconv.Itoa(s[0].Line)
		r0 := strconv.Itoa(s[0].RuneOnLine())
		l1 := strconv.Itoa(s[1].Line)
		r1 := strconv.Itoa(s[1].RuneOnLine())
		return s[0].Path + ":" + l0 + "-+#" + r0 + "," + l1 + "-+#" + r1
	default:
		return s[0].String()
	}
}
