package htmldoc

import (
	"fmt"
	"io"
	"regexp"
	"strings"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

// stringer is a struct that can be expressed as a string.
type stringer interface {
	toString() string
}

// container is a struct used to buffer unrendered text.
type container interface {
	add(stringer) error
	toString() string
	width() int
}

// styleBlock is a placeholder container that discards any content inside it.
type styleBlock struct{}

// add discards content.
func (block styleBlock) add(s stringer) error {
	return nil
}

// toString always returns an empty string.
func (block styleBlock) toString() string {
	return ""
}

// width always returns 0.
func (block styleBlock) width() int {
	return 0
}

// text represents word-wrapped string content.
type text struct {
	content string
	width   int
}

// toString renders a text's content as a word-wrapped string
func (t text) toString() string {
	//return wrap(block.content, block.width)
	return t.content
}

// hRule represents a horizontal rule.
type hRule struct {
	width int
}

// toString renders an hRule.
func (hr hRule) toString() string {
	return strings.Repeat("-", hr.width)
}

// A Document represents HTML rendered as text suitable for output within a
// terminal or other text-only environments. Optionally, minimal formatting can
// be applied using ANSI escape sequenes.
type Document struct {
	// RefMap is used to look-up elements with href attributes.
	RefMap map[string]string

	// Reader is the HTML source.
	Reader io.Reader

	// Writer is the destination for rendered text.
	Writer io.Writer

	// Width is the maximum width of a Document's rendered text.
	Width int

	// ANSIEnabled determines whether or not to format rendered text using ANSI
	// escape sequences.
	ANSIEnabled bool

	// containerStack is used to buffer unrendered text.
	containerStack []container
}

// add stores renderable content in the active container of a Document.
func (doc Document) add(s stringer) error {
	if c := doc.activeContainer(); c != nil {
		c.add(s)
		return nil
	}

	_, err := io.WriteString(doc.Writer, s.toString())
	return err
}

// activeContainer returns the top container on a Document's container stack.
func (doc Document) activeContainer() container {
	if last := len(doc.containerStack) - 1; last >= 0 {
		return doc.containerStack[last]
	}

	return nil
}

func (doc Document) activeContainerWidth() int {
	if c := doc.activeContainer(); c != nil {
		return c.width()
	}

	return doc.Width
}

// Render writes rendered text to a Document's Writer.
func (doc Document) Render() (err error) {
	tokenizer := html.NewTokenizer(doc.Reader)
	for {
		tokenType := tokenizer.Next()
		token := tokenizer.Token()
		switch tokenType {
		case html.ErrorToken:
			err = tokenizer.Err()
			if err == io.EOF {
				return nil
			} else if err != nil {
				return err
			}
		case html.StartTagToken:
			doc.handleStartTag(token)
			fallthrough
		case html.SelfClosingTagToken:
			doc.handleTag(token)
		case html.TextToken:
			if err = doc.handleText(token); err != nil {
				return err
			}
		case html.EndTagToken:
			doc.handleEndTag(token)
		}
	}
}

// stripFormatting replaces each occurrence of one or more whitespace
// characters with a single space.
func stripFormatting(s string) string {
	re := regexp.MustCompile(`\s+`)
	return strings.TrimSpace(re.ReplaceAllString(s, " "))
}

// handleText adds text elements to a Document.
func (doc Document) handleText(token html.Token) error {
	content := stripFormatting(token.Data)
	t := text{content: content, width: doc.activeContainerWidth()}
	return doc.add(t)
}

// handleStartTag adds a new container to a Document's container stack, if
// necessary.
func (doc *Document) handleStartTag(token html.Token) {
	switch token.DataAtom {
	case atom.Style:
		doc.containerStack = append(doc.containerStack, styleBlock{})
	}
}

// handleTag adds text representations of non-text elements to a Document.
func (doc Document) handleTag(token html.Token) {
	switch token.DataAtom {
	case atom.Img:
		for _, a := range token.Attr {
			switch atom.Lookup([]byte(a.Key)) {
			case atom.Alt:
				altText := fmt.Sprintf("Alt text: %s\n", a.Val)
				t := text{content: altText, width: doc.activeContainerWidth()}
				doc.add(t)
			}
		}
	case atom.Br:
		t := text{content: "\n", width: 0}
		doc.add(t)
	case atom.Hr:
		width := doc.activeContainerWidth()
		if width <= 0 {
			width = 5 // default to 5 if parent width isn't specified
		}
		doc.add(hRule{width: width})
	}
}

// handleEndTag pops a container from a Document's container stack, if
// necessary.
func (doc *Document) handleEndTag(token html.Token) {
	switch token.DataAtom {
	case atom.Style:
		doc.popContainer(atom.Style)
	}
}

// popContainer removes the top container from a Document's container stack.
func (doc *Document) popContainer(a atom.Atom) container {
	last := len(doc.containerStack) - 1
	if last < 0 {
		return nil
	}
	switch v := doc.containerStack[last].(type) {
	case styleBlock:
		if a != atom.Style {
			return nil
		}
		doc.containerStack = doc.containerStack[:last]
		return v
	}
	return nil
}
