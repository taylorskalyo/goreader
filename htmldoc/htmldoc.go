package htmldoc

import (
	"fmt"
	"io"
	"regexp"
	"strings"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

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
	return doc.add(str{stripFormatting(token.Data)})
}

// handleStartTag modifies a Document based on a start tag token.
func (doc *Document) handleStartTag(token html.Token) {
	switch token.DataAtom {
	case atom.Style:
		doc.containerStack = append(doc.containerStack, &styleBlock{})
	case atom.Html, atom.Body, atom.Head:
		t := textBlock{w: doc.activeContainerWidth()}
		doc.containerStack = append(doc.containerStack, &t)
	case atom.Div:
		// Separate blocks with content.
		if c := doc.activeContainer(); c != nil && c.hasContent() {
			doc.add(str{"\n"})
		}
		t := textBlock{w: doc.activeContainerWidth()}
		doc.containerStack = append(doc.containerStack, &t)
	case atom.P, atom.H1, atom.H2, atom.H3, atom.H4, atom.H5, atom.H6:
		// Separate blocks with content.
		if c := doc.activeContainer(); c != nil && c.hasContent() {
			doc.add(str{"\n\n"})
		}
		t := textBlock{w: doc.activeContainerWidth()}
		doc.containerStack = append(doc.containerStack, &t)
	}
}

// handleTag modifies a Document based on a tag token.
func (doc Document) handleTag(token html.Token) {
	switch token.DataAtom {
	case atom.Img:
		for _, a := range token.Attr {
			switch atom.Lookup([]byte(a.Key)) {
			case atom.Alt:
				altText := fmt.Sprintf("Alt text: %s\n", a.Val)
				doc.add(str{altText})
			}
		}
	case atom.Br:
		doc.add(str{"\n"})
	case atom.Hr:
		width := doc.activeContainerWidth()
		if width <= 0 {
			width = 5 // default to 5 if parent width isn't specified
		}
		doc.add(hRule{width: width})
	}
}

// handleEndTag modifies a Document based on an end tag token.
func (doc *Document) handleEndTag(token html.Token) {
	switch token.DataAtom {
	case atom.Style:
		doc.popContainer(token.DataAtom)
	case atom.Html, atom.Body, atom.Head, atom.Div, atom.P, atom.H1, atom.H2, atom.H3, atom.H4, atom.H5, atom.H6:
		t := doc.popContainer(token.DataAtom)
		if t != nil {
			doc.add(t)
		}
	}
}

// popContainer removes the top container from a Document's container stack.
func (doc *Document) popContainer(a atom.Atom) container {
	last := len(doc.containerStack) - 1
	if last < 0 {
		return nil
	}
	switch v := doc.containerStack[last].(type) {
	case *styleBlock:
		if a != atom.Style {
			return nil
		}
		doc.containerStack = doc.containerStack[:last]
		return v
	case *textBlock:
		switch a {
		case atom.Html, atom.Body, atom.Head, atom.Div, atom.P, atom.H1, atom.H2, atom.H3, atom.H4, atom.H5, atom.H6:
		default:
			return nil
		}
		doc.containerStack = doc.containerStack[:last]
		return v
	}
	return nil
}
