package main

import (
	"bytes"
	"fmt"
	"io"

	"golang.org/x/net/html"
)

type renderer struct {
	elStack []string
	t       *html.Tokenizer
	b       bytes.Buffer
}

// render takes in html content via an io.Reader and returns a buffer
// containing only plain text.
func render(r io.Reader) (bytes.Buffer, error) {
	t := html.NewTokenizer(r)
	re := renderer{t: t}
	err := re.parse(r)
	if err != nil {
		return re.b, err
	}
	return re.b, nil
}

// parse walks an html document and appends text elements to a buffer.
func (re *renderer) parse(io.Reader) (err error) {
	for {
		switch re.t.Next() {
		case html.ErrorToken:
			err = re.t.Err()
		case html.StartTagToken:
			re.elStack = append(re.elStack, re.t.Token().Data) // push element
			fallthrough
		case html.SelfClosingTagToken:
			err = re.handleTags()
		case html.TextToken:
			err = re.handleText()
		case html.EndTagToken:
			re.elStack = re.elStack[:len(re.elStack)-1] // pop element
		}
		if err == io.EOF {
			return nil
		} else if err != nil {
			return err
		}
	}
}

// handleText appends text elements to the renderer buffer. It filters elements
// that should not be displayed as text (e.g. style blocks).
func (re *renderer) handleText() error {
	// Skip style tags
	if len(re.elStack) > 0 && re.elStack[len(re.elStack)-1] == "style" {
		return nil
	}
	_, err := re.b.Write(re.t.Text())
	return err
}

// handleTags appends text representations of non-text elements (e.g. image alt
// tags) to the renderer buffer.
func (re *renderer) handleTags() error {
	token := re.t.Token()
	// Display alt text in place of image
	if token.Data == "img" {
		for _, a := range token.Attr {
			if a.Key == "alt" {
				_, err := re.b.Write([]byte(fmt.Sprintf("\nImage alt text: %s\n", a.Val)))
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}
