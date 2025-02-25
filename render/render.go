package render

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/color"
	"io"
	"strings"

	_ "image/jpeg"
	_ "image/png"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/nfnt/resize"
	"github.com/rivo/tview"
	"github.com/taylorskalyo/goreader/config"
	"github.com/taylorskalyo/goreader/epub"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

var tableStyle = table.StyleDefault

// Renderer is responsible for rendering epub content.
type Renderer struct {
	content *epub.Package
	theme   config.Theme
	width   int
	parser  parser
}

// parser represents the current parsing state.
type parser struct {
	tagStack   []atom.Atom
	tableStack []table.Writer
	rowStack   [][]string
	cellStack  []strings.Builder

	tokenizer *html.Tokenizer
	newlines  int
	indents   int
	writer    *wordWrapWriter
}

// New returns a new epub Renderer.
func New(content *epub.Package) Renderer {
	return Renderer{
		content: content,
		width:   80,
		theme:   config.Default().Theme,
	}
}

// SetTheme sets style options for a Renderer.
func (r *Renderer) SetTheme(theme config.Theme) {
	r.theme = theme
}

// RenderChapter reads in an epub item, parses the content, and writes the
// rendered output to the given writer.
func (r *Renderer) RenderChapter(ctx context.Context, chapter int, w io.Writer) error {
	doc, err := r.content.Spine.Itemrefs[chapter].Open()
	if err != nil {
		return err
	}

	r.parser = parser{
		tokenizer: html.NewTokenizer(doc),
		writer:    newWordWrapWriter(w, r.width),
	}

	return r.render(ctx)
}

// tviewStyle constructs a tview style tag based on HTML tags in the tag stack.
func (r Renderer) tviewStyle(tags []atom.Atom) string {
	style := config.DefaultStyle()
	for _, tag := range tags {
		if s, ok := r.theme[tag.String()]; ok {
			style = style.Merge(s)
		}
	}

	return style.String()
}

// render walks an html document and renders elements to a writer.
func (r *Renderer) render(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
		}

		if err := r.handleToken(); err == io.EOF {
			r.parser.writer.Flush()
			return nil
		} else if err == io.EOF {
			return err
		}
	}
}

// handleToken is triggered when an HTML token is parsed.
func (r *Renderer) handleToken() error {
	tokenType := r.parser.tokenizer.Next()
	token := r.parser.tokenizer.Token()
	switch tokenType {
	case html.ErrorToken:
		return r.parser.tokenizer.Err()
	case html.StartTagToken:
		r.parser.tagStack = append(r.parser.tagStack, token.DataAtom) // push element
		return r.handleStartTag(token)
	case html.SelfClosingTagToken:
		return r.handleStartTag(token)
	case html.TextToken:
		return r.handleText(token)
	case html.EndTagToken:
		r.parser.tagStack = r.parser.tagStack[:len(r.parser.tagStack)-1] // pop element
		r.parser.indents = 0
		return r.handleEndTag(token)
	}

	return nil
}

// appendText appends text to the underlying writer.
func (r *Renderer) appendText(text string) error {
	if !hasText(text) {
		return nil
	}

	text = tview.Escape(text)

	pendingLines := strings.Repeat("\n", r.parser.newlines)
	pendingIndents := strings.Repeat(" ", r.parser.indents)
	text = fmt.Sprintf("%s%s%s", pendingLines, pendingIndents, text)

	r.parser.newlines = 0
	r.parser.indents = 0

	_, err := io.WriteString(r.parser.writeTarget(), text)

	return err
}

// handleText appends text elements to the parser buffer. It filters elements
// that should not be displayed as text (e.g. style blocks).
func (r *Renderer) handleText(token html.Token) error {
	// Skip style tags
	if len(r.parser.tagStack) > 0 && r.parser.tagStack[len(r.parser.tagStack)-1] == atom.Style {
		return nil
	}

	style := r.tviewStyle(r.parser.tagStack)
	if _, err := io.WriteString(r.parser.writer, style); err != nil {
		return err
	}

	text := processWhitespace(token.Data)
	return r.appendText(string(text))
}

// ensureNewlines ensures that there are at least this many pending newlines.
func (p *parser) ensureNewlines(n int) {
	if p.newlines >= n {
		return
	}

	p.newlines = n
}

// ensureIndents ensure that there are at least this many pending indents.
func (p *parser) ensureIndents(n int) {
	if p.indents >= n {
		return
	}

	p.indents = n
}

// determine if we are writing to a table or the main writer.
func (p parser) writeTarget() io.Writer {
	if len(p.cellStack) > 0 {
		return &p.cellStack[len(p.cellStack)-1]
	}

	return p.writer
}

// handleStartTag appends text representations of non-text elements (e.g. image
// alt tags) to the parser buffer.
func (r *Renderer) handleStartTag(token html.Token) (err error) {
	switch token.DataAtom {
	case atom.Img:
		err = r.handleImage(token)
	case atom.Br:
		r.parser.newlines++
	case atom.H1, atom.H2, atom.H3, atom.H4, atom.H5, atom.H6, atom.Title,
		atom.Div:
		r.parser.ensureNewlines(2)
	case atom.P:
		r.parser.ensureNewlines(2)
		r.parser.ensureIndents(2)
	case atom.Hr:
		r.parser.ensureNewlines(2)
		err = r.appendText(strings.Repeat(tableStyle.Box.MiddleHorizontal, r.width))
		r.parser.ensureNewlines(2)
	case atom.Table:
		t := table.NewWriter()
		r.parser.tableStack = append(r.parser.tableStack, t) // push table
	case atom.Th, atom.Td:
		r.parser.cellStack = append(r.parser.cellStack, strings.Builder{}) // push cell writer
	}

	return err
}

// handleEndTag appends text that has been buffered until an element has been
// fully parsed (e.g. tables).
func (r *Renderer) handleEndTag(token html.Token) (err error) {
	switch token.DataAtom {
	case atom.Tr:
		row := make([]string, len(r.parser.cellStack))
		for i, cell := range r.parser.cellStack {
			row[i] = strings.TrimSpace(cell.String())
		}
		r.parser.rowStack = append(r.parser.rowStack, row) // push row
		r.parser.cellStack = []strings.Builder{}           // pop all cell writers
	case atom.Table:
		t := r.parser.tableStack[len(r.parser.tableStack)-1]
		style := tableStyle
		style.Size.WidthMax = r.width
		style.Options.DrawBorder = true
		style.Options.SeparateRows = true
		style.Options.SeparateColumns = false

		colConfigs := r.columnConfigs(style)
		t.ImportGrid(r.parser.rowStack)
		t.SetColumnConfigs(colConfigs)
		t.SetStyle(style)
		r.parser.rowStack = [][]string{}                                       // pop all rows
		r.parser.tableStack = r.parser.tableStack[:len(r.parser.tableStack)-1] // pop table

		if len(colConfigs) > 0 {
			r.parser.ensureNewlines(2)
			if err := r.appendText(t.Render()); err != nil {
				return err
			}
			r.parser.ensureNewlines(2)
		}
	}

	return err
}

// columnConfigs configures columns to fit within the viewport. It sets each
// column's max width and handles text wrapping.
func (r *Renderer) columnConfigs(style table.Style) (configs []table.ColumnConfig) {
	columnCount := 0
	for _, row := range r.parser.rowStack {
		if len(row) > columnCount {
			columnCount = len(row)
		}
	}

	if columnCount < 1 {
		return configs
	}

	configs = make([]table.ColumnConfig, columnCount)

	// Determine width used up by table decorations.
	decorationWidth := 2 * columnCount // padding
	if style.Options.DrawBorder {
		decorationWidth += 2
	}
	if style.Options.SeparateColumns {
		decorationWidth += columnCount - 1
	}

	// Determine width availble for text.
	availableWidth := r.width - decorationWidth
	for i := range configs {
		configs[i] = table.ColumnConfig{
			Number:           i + 1,
			WidthMaxEnforcer: tviewWidthEnforcer,
		}
	}

	// Dynamically choose column width.
	strategies := []func(int, *[]table.ColumnConfig) bool{
		r.parser.tryFitColumn,
		r.parser.tryFairColumn,
	}
	for _, fn := range strategies {
		if ok := fn(availableWidth, &configs); ok {
			return configs
		}
	}

	return []table.ColumnConfig{}
}

// tryFitColumn fits each column width to its cell content. If the overall
// width is greater than the available width, it will abort.
func (p *parser) tryFitColumn(availableWidth int, configs *[]table.ColumnConfig) bool {
	maxWidths := make([]int, len(*configs))
	for _, row := range p.rowStack {
		for col, cell := range row {
			if len(cell) > maxWidths[col] {
				maxWidths[col] = len(cell)
			}
		}
	}

	totalMaxWidth := 0
	for _, width := range maxWidths {
		totalMaxWidth += width
	}

	if totalMaxWidth > availableWidth {
		return false
	}

	for i := range *configs {
		(*configs)[i].WidthMax = maxWidths[i]
	}

	return true
}

// tryFairColumn gives each column a proportion of the available width with a
// bias towards equal widths.
func (p *parser) tryFairColumn(availableWidth int, configs *[]table.ColumnConfig) bool {
	equalWidth := availableWidth / len(*configs)
	fairWidths := make([]int, len(*configs))
	for _, row := range p.rowStack {
		for col, cell := range row {
			fairWidth := (len(cell) + equalWidth) / 2
			if fairWidth > fairWidths[col] {
				fairWidths[col] = fairWidth
			}
		}
	}

	totalFairWidth := 0
	for _, width := range fairWidths {
		totalFairWidth += width
	}

	for i := range *configs {
		ratio := float64(fairWidths[i]) / float64(totalFairWidth)
		width := int(float64(availableWidth) * ratio)
		(*configs)[i].WidthMax = width
	}

	return true
}

// handleImage appends image elements to the parser buffer. It extracts alt
// text and renders images.
func (r *Renderer) handleImage(token html.Token) error {
	for _, a := range token.Attr {
		switch atom.Lookup([]byte(a.Key)) {
		case atom.Alt:
			text := fmt.Sprintf("Alt text: %s", a.Val)
			r.parser.ensureNewlines(1)
			if err := r.appendText(text); err != nil {
				return err
			}
			r.parser.ensureNewlines(1)
		case atom.Src:
			if err := r.handleImageSrc(a.Val); err != nil {
				return err
			}
		}
	}

	return nil
}

// handleImageSrc reads a referenced image and renders it to the parser buffer.
func (r *Renderer) handleImageSrc(href string) error {
	if r.parser.writeTarget() != r.parser.writer {
		// NOTE: rendering images inside tables is not supported at the moment as
		// this would add a lot of complexity.
		return nil
	}

	for _, item := range r.content.Items {
		if item.HREF == href {
			for _, line := range imageToText(item, r.width) {
				r.parser.ensureNewlines(1)
				if err := r.appendText(line); err != nil {
					return err
				}
				r.parser.ensureNewlines(1)
			}
			break
		}
	}

	return nil
}

// imageToText renders an image as lines of text.
func imageToText(item epub.Item, width int) []string {
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
	h := (bounds.Max.Y * width) / (bounds.Max.X * 2)
	img = resize.Resize(uint(width), uint(h), img, resize.Lanczos3)

	charGradient := []rune("MND8OZ$7I?+=~:,..")
	buf := new(bytes.Buffer)

	for y := 0; y < h; y++ {
		for x := 0; x < width; x++ {
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
