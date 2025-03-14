package render

import (
	"context"
	"fmt"
	"io"
	"path"
	"strings"

	_ "image/jpeg"
	_ "image/png"

	"github.com/jedib0t/go-pretty/v6/table"
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
	basepath  string
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
	item := r.content.Spine.Itemrefs[chapter]
	doc, err := item.Open()
	if err != nil {
		return err
	}

	r.parser = parser{
		tokenizer: html.NewTokenizer(doc),
		writer:    newWordWrapWriter(w, r.width),
		basepath:  path.Dir(item.HREF),
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
