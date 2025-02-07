package parse

import (
	"bufio"
	"bytes"
	"fmt"
	"image"
	"image/color"
	"io"
	"strings"
	"unicode"

	_ "image/jpeg"
	_ "image/png"

	"github.com/gdamore/tcell/v2"
	"github.com/nfnt/resize"
	"github.com/taylorskalyo/goreader/epub"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

type parser struct {
	tagStack  []atom.Atom
	tokenizer *html.Tokenizer
	doc       Cellbuf
	items     []epub.Item
}

type Cellbuf struct {
	Cells   []tcell.SimCell
	Width   int
	lmargin int
	col     int
	row     int
	space   bool
	style   tcell.Style
}

// setCell changes a cell's attributes in the cell buffer document at the given
// position.
func (c *Cellbuf) setCell(x, y int, runes []rune, style tcell.Style) {
	// Grow in steps of 1024 when out of space.
	for y*c.Width+x >= len(c.Cells) {
		c.Cells = append(c.Cells, make([]tcell.SimCell, 1024)...)
	}
	c.Cells[y*c.Width+x] = tcell.SimCell{Runes: runes, Style: style}
}

// style sets the foreground/background attributes for future cells in the cell
// buffer document based on HTML tags in the tag stack.
func (c *Cellbuf) setStyle(tags []atom.Atom) {
	style := tcell.StyleDefault
	for _, tag := range tags {
		switch tag {
		case atom.B, atom.Strong, atom.Em:
			style = style.Bold(true)
		case atom.I:
			style = style.Foreground(tcell.ColorOlive)
		case atom.Title:
			style = style.Foreground(tcell.ColorMaroon)
		case atom.H1:
			style = style.Foreground(tcell.ColorPurple)
		case atom.H2:
			style = style.Foreground(tcell.ColorNavy)
		case atom.H3, atom.H4, atom.H5, atom.H6:
			style = style.Foreground(tcell.ColorTeal)
		}
	}
	c.style = style
}

// appendText appends text to the cell buffer document.
func (c *Cellbuf) appendText(str string) {
	if len(str) <= 0 {
		return
	}
	if c.col < c.lmargin {
		c.col = c.lmargin
	}
	runes := []rune(str)
	if unicode.IsSpace(runes[0]) {
		c.space = true
	}
	scanner := bufio.NewScanner(strings.NewReader(strings.TrimSpace(str)))
	scanner.Split(bufio.ScanWords)
	for scanner.Scan() {
		if c.col != c.lmargin && c.space {
			c.col++
		}
		word := []rune(scanner.Text())
		if len(word) > c.Width-c.col {
			c.row++
			c.col = c.lmargin
		}
		for _, r := range word {
			c.setCell(c.col, c.row, []rune{r}, c.style)
			c.col++
		}
		c.space = true
	}
	if !unicode.IsSpace(runes[len(runes)-1]) {
		c.space = false
	}
}

// parseText takes in html content via an io.Reader and returns a buffer
// containing only plain text.
func ParseText(r io.Reader, items []epub.Item) (Cellbuf, error) {
	tokenizer := html.NewTokenizer(r)
	doc := Cellbuf{Width: 80}
	p := parser{tokenizer: tokenizer, doc: doc, items: items}
	err := p.parse(r)
	if err != nil {
		return p.doc, err
	}
	return p.doc, nil
}

// parse walks an html document and renders elements to a cell buffer document.
func (p *parser) parse(io.Reader) (err error) {
	for {
		tokenType := p.tokenizer.Next()
		token := p.tokenizer.Token()
		switch tokenType {
		case html.ErrorToken:
			err = p.tokenizer.Err()
		case html.StartTagToken:
			p.tagStack = append(p.tagStack, token.DataAtom) // push element
			fallthrough
		case html.SelfClosingTagToken:
			p.handleStartTag(token)
		case html.TextToken:
			p.handleText(token)
		case html.EndTagToken:
			p.tagStack = p.tagStack[:len(p.tagStack)-1] // pop element
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
func (p *parser) handleText(token html.Token) {
	// Skip style tags
	if len(p.tagStack) > 0 && p.tagStack[len(p.tagStack)-1] == atom.Style {
		return
	}
	p.doc.setStyle(p.tagStack)
	p.doc.appendText(string(token.Data))
}

// handleStartTag appends text representations of non-text elements (e.g. image alt
// tags) to the parser buffer.
func (p *parser) handleStartTag(token html.Token) {
	switch token.DataAtom {
	case atom.Img:
		p.handleImage(token)
	case atom.Br:
		p.doc.row++
		p.doc.col = p.doc.lmargin
	case atom.H1, atom.H2, atom.H3, atom.H4, atom.H5, atom.H6, atom.Title,
		atom.Div, atom.Tr:
		p.doc.row += 2
		p.doc.col = p.doc.lmargin
	case atom.P:
		p.doc.row += 2
		p.doc.col = p.doc.lmargin
		p.doc.col += 2
	case atom.Hr:
		p.doc.row++
		p.doc.col = 0
		p.doc.appendText(strings.Repeat("-", p.doc.Width))
	}
}

// handleImage appends image elements to the parser buffer. It extracts alt
// text and converts images to ascii art.
func (p *parser) handleImage(token html.Token) {
	for _, a := range token.Attr {
		switch atom.Lookup([]byte(a.Key)) {
		case atom.Alt:
			text := fmt.Sprintf("Alt text: %s", a.Val)
			p.doc.appendText(text)
			p.doc.row++
			p.doc.col = p.doc.lmargin
		case atom.Src:
			for _, item := range p.items {
				if item.HREF == a.Val {
					for _, line := range imageToText(item) {
						p.doc.appendText(line)
						p.doc.col = p.doc.lmargin
						p.doc.row++
					}
					break
				}
			}
		}
	}
}

func imageToText(item epub.Item) []string {
	lines := []string{}
	r, err := item.Open()
	if err != nil {
		return lines
	}

	img, _, err := image.Decode(r)
	if err != nil {
		return lines
	}
	bounds := img.Bounds()

	// Assume a character height to width ratio of 2:1.
	w := 80
	h := (bounds.Max.Y * w) / (bounds.Max.X * 2)
	img = resize.Resize(uint(w), uint(h), img, resize.Lanczos3)

	charGradient := []rune("MND8OZ$7I?+=~:,..")
	buf := new(bytes.Buffer)

	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			c := color.GrayModel.Convert(img.At(x, y))
			y := c.(color.Gray).Y
			pos := (len(charGradient) - 1) * int(y) / 255
			buf.WriteRune(charGradient[pos])
		}
		lines = append(lines, buf.String())
		buf = new(bytes.Buffer)
	}

	return lines
}
