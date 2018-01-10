package htmldoc

import "strings"

// stringer is a struct that can be expressed as a string.
type stringer interface {
	toString() string
}

// str represents a raw string.
type str struct {
	content string
}

// toString returns a str's content.
func (s str) toString() string {
	return s.content
}

// hRule represents a horizontal rule.
type hRule struct {
	width int
}

// toString renders an hRule.
func (hr hRule) toString() string {
	return strings.Repeat("-", hr.width)
}
