package htmldoc

import (
	"strings"
)

// container is a struct used to buffer unrendered text.
type container interface {
	add(stringer) error
	hasContent() bool
	toString() string
	width() int
}

// styleBlock is a placeholder container that discards any content inside it.
type styleBlock struct{}

// add discards content.
func (block *styleBlock) add(s stringer) error {
	return nil
}

// hasContent always returns false.
func (block styleBlock) hasContent() bool {
	return false
}

// toString always returns an empty string.
func (block styleBlock) toString() string {
	return ""
}

// width always returns 0.
func (block styleBlock) width() int {
	return 0
}

// textBlock represents word-wrapped string content.
type textBlock struct {
	content string
	w       int
}

// add appends text to a textBlock's content.
func (block *textBlock) add(s stringer) error {
	block.content = block.content + s.toString()
	return nil
}

// hasContent always returns false.
func (block textBlock) hasContent() bool {
	return len(block.content) > 0
}

// toString renders a textBlock's content as a word-wrapped string.
func (block textBlock) toString() (s string) {
	s = strings.TrimSpace(wrap(block.content, block.w))
	return
}

// width returns the width of a textBlock.
func (block textBlock) width() int {
	return block.w
}
