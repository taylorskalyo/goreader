package main

import (
	"bytes"
	"fmt"
	"io"

	"golang.org/x/net/html"
)

type parser struct {
	elStack   []string
	tokenizer *html.Tokenizer
	buf       bytes.Buffer
}

// parseText takes in html content via an io.Reader and returns a buffer
// containing only plain text.
func parseText(r io.Reader) (bytes.Buffer, error) {
	tokenizer := html.NewTokenizer(r)
	p := parser{tokenizer: tokenizer}
	err := p.parse(r)
	if err != nil {
		return p.buf, err
	}
	return p.buf, nil
}

// parse walks an html document and appends text elements to a buffer.
func (p *parser) parse(io.Reader) (err error) {
	for {
		switch p.tokenizer.Next() {
		case html.ErrorToken:
			err = p.tokenizer.Err()
		case html.StartTagToken:
			p.elStack = append(p.elStack, p.tokenizer.Token().Data) // push element
			fallthrough
		case html.SelfClosingTagToken:
			err = p.handleTags()
		case html.TextToken:
			err = p.handleText()
		case html.EndTagToken:
			p.elStack = p.elStack[:len(p.elStack)-1] // pop element
		}
		if err == io.EOF {
			return nil
		} else if err != nil {
			return err
		}
	}
}

// handleText appends text elements to the parser buffer. It filters elements
// that should not be displayed as text (e.g. style blocks).
func (p *parser) handleText() error {
	// Skip style tags
	if len(p.elStack) > 0 && p.elStack[len(p.elStack)-1] == "style" {
		return nil
	}
	_, err := p.buf.Write(p.tokenizer.Text())
	return err
}

// handleTags appends text representations of non-text elements (e.g. image alt
// tags) to the parser buffer.
func (p *parser) handleTags() error {
	token := p.tokenizer.Token()
	// Display alt text in place of image
	if token.Data == "img" {
		for _, a := range token.Attr {
			if a.Key == "alt" {
				_, err := p.buf.Write([]byte(fmt.Sprintf("\nImage alt text: %s\n", a.Val)))
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}
