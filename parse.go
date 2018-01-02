package main

import (
	"bufio"
	"fmt"
	"io"
	"strings"
	"unicode/utf8"

	termbox "github.com/nsf/termbox-go"

	"golang.org/x/net/html"
)

type parser struct {
	elStack   []string
	tokenizer *html.Tokenizer
	doc       cellbuf
}

type cellbuf struct {
	cells   []termbox.Cell
	width   int
	lmargin int
	col     int
	row     int
}

// setCell changes a cell's attributes in the cell buffer document at the given
// position.
func (c *cellbuf) setCell(x, y int, ch rune, fg, bg termbox.Attribute) {
	// Grow in steps of 1024 when out of space.
	for y*c.width+x >= len(c.cells) {
		c.cells = append(c.cells, make([]termbox.Cell, 1024)...)
	}
	c.cells[y*c.width+x] = termbox.Cell{Ch: ch, Fg: fg, Bg: bg}
}

// scanWords is a split function for a Scanner that returns space-separated
// words. Unlike bufio.ScanWords(), scanWords only splits on spaces (i.e. not
// newlines, tabs, or other whitespace).
func scanWords(data []byte, atEOF bool) (advance int, token []byte, err error) {
	// Skip leading spaces.
	start := 0
	for width := 0; start < len(data); start += width {
		var r rune
		r, width = utf8.DecodeRune(data[start:])
		if r != ' ' {
			break
		}
	}

	// Scan until space, marking end of word.
	for width, i := 0, start; i < len(data); i += width {
		var r rune
		r, width = utf8.DecodeRune(data[i:])
		if r == ' ' {
			return i + width, data[start:i], nil
		}
	}

	// If we're at EOF, we have a final, non-empty, non-terminated word. Return
	// it.
	if atEOF && len(data) > start {
		return len(data), data[start:], nil
	}

	// Request more data.
	return start, nil, nil
}

// appendStyledText appends styled text to the cell buffer document.
func (c *cellbuf) appendStyledText(str string, fg, bg termbox.Attribute) {
	if c.col < c.lmargin {
		c.col = c.lmargin
	}
	scanner := bufio.NewScanner(strings.NewReader(str))
	scanner.Split(scanWords)
	for scanner.Scan() {
		word := []rune(scanner.Text())
		if len(word) > c.width-c.col {
			c.row++
			c.col = c.lmargin
		}
		for _, r := range word {
			if r == '\n' {
				c.row++
				c.col = c.lmargin
				continue
			}
			c.setCell(c.col, c.row, r, fg, bg)
			c.col++
		}
		if c.col < c.width {
			c.col++
		}
	}
}

// appendText appends unstyled text to the cell buffer document.
func (c *cellbuf) appendText(str string) {
	c.appendStyledText(str, termbox.ColorDefault, termbox.ColorDefault)
}

// parseText takes in html content via an io.Reader and returns a buffer
// containing only plain text.
func parseText(r io.Reader) (cellbuf, error) {
	tokenizer := html.NewTokenizer(r)
	doc := cellbuf{width: 80}
	p := parser{tokenizer: tokenizer, doc: doc}
	err := p.parse(r)
	if err != nil {
		return p.doc, err
	}
	return p.doc, nil
}

// parse walks an html document and renders elements to a cell buffer document.
func (p *parser) parse(io.Reader) (err error) {
	for {
		switch p.tokenizer.Next() {
		case html.ErrorToken:
			err = p.tokenizer.Err()
		case html.StartTagToken:
			p.elStack = append(p.elStack, p.tokenizer.Token().Data) // push element
			fallthrough
		case html.SelfClosingTagToken:
			p.handleTags()
		case html.TextToken:
			p.handleText()
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
func (p *parser) handleText() {
	// Skip style tags
	if len(p.elStack) > 0 && p.elStack[len(p.elStack)-1] == "style" {
		return
	}
	p.doc.appendText(string(p.tokenizer.Text()))
}

// handleTags appends text representations of non-text elements (e.g. image alt
// tags) to the parser buffer.
func (p *parser) handleTags() {
	token := p.tokenizer.Token()
	switch token.Data {
	case "img":
		// Display alt text in place of images.
		for _, a := range token.Attr {
			switch a.Key {
			case "alt":
				text := fmt.Sprintf("\nImage alt text: %s\n", a.Val)
				p.doc.appendText(text)
			}
		}
	case "br":
		p.doc.appendText("\n")
	}
}
