package nav

import (
	"github.com/gdamore/tcell/v2"
	"github.com/taylorskalyo/goreader/parse"
)

type PageNavigator interface {
	Draw()
	MaxScrollX() int
	MaxScrollY() int
	PageDown() bool
	PageUp() bool
	Pages() int
	ScrollDown()
	ScrollLeft()
	ScrollRight()
	ScrollUp()
	SetDoc(parse.Cellbuf)
	SetScreen(tcell.Screen)
	Size() (int, int)
	ToBottom()
	ToTop()
}

type Pager struct {
	scrollX int
	scrollY int
	doc     parse.Cellbuf
	screen  tcell.Screen
}

// setDoc sets the pager's cell buffer
func (p *Pager) SetDoc(doc parse.Cellbuf) {
	p.doc = doc
}

// setScreen sets the pager's screen
func (p *Pager) SetScreen(screen tcell.Screen) {
	p.screen = screen
}

// Draw displays a pager's cell buffer in the terminal.
func (p *Pager) Draw() {
	p.screen.Clear()

	width, height := p.screen.Size()
	var centerOffset int
	for y := 0; y < height; y++ {
		for x := 0; x < p.doc.Width; x++ {
			index := (y+p.scrollY)*p.doc.Width + x
			if index >= len(p.doc.Cells) || index <= 0 {
				continue
			}
			cell := p.doc.Cells[index]
			if width > p.doc.Width {
				centerOffset = (width - p.doc.Width) / 2
			}

			if len(cell.Runes) > 0 {
        // Calling SetContent with coordinates outside of the terminal viewport
        // results in a no-op.
				p.screen.SetContent(x+p.scrollX+centerOffset, y, cell.Runes[0], cell.Runes[1:], cell.Style)
			}
		}
	}

	p.screen.Show()
}

// scrollDown pans the pager's viewport down, without exceeding the underlying
// cell buffer document's boundaries.
func (p *Pager) ScrollDown() {
	if p.scrollY < p.MaxScrollY() {
		p.scrollY++
	}
}

// scrollUp pans the pager's viewport up, without exceeding the underlying cell
// buffer document's boundaries.
func (p *Pager) ScrollUp() {
	if p.scrollY > 0 {
		p.scrollY--
	}
}

// scrollLeft pans the pager's viewport left, without exceeding the underlying
// cell buffer document's boundaries.
func (p *Pager) ScrollLeft() {
	if p.scrollX < 0 {
		p.scrollX++
	}
}

// scrollRight pans the pager's viewport right, without exceeding the
// underlying cell buffer document's boundaries.
func (p *Pager) ScrollRight() {
	if p.scrollX > -p.MaxScrollX() {
		p.scrollX--
	}
}

// pageDown pans the pager's viewport down by a full page, without exceeding
// the underlying cell buffer document's boundaries.
func (p *Pager) PageDown() bool {
	_, viewHeight := p.screen.Size()
	if p.scrollY < p.MaxScrollY() {
		p.scrollY += viewHeight
		return true
	}

	return false
}

// pageUp pans the pager's viewport up by a full page, without exceeding the
// underlying cell buffer document's boundaries.
func (p *Pager) PageUp() bool {
	_, viewHeight := p.screen.Size()
	if p.scrollY > viewHeight {
		p.scrollY -= viewHeight
		return true
	} else if p.scrollY > 0 {
		p.scrollY = 0
		return true
	}

	return false
}

// toTop set's the pager's horizontal and vertical panning distance back to
// zero.
func (p *Pager) ToTop() {
	p.scrollX = 0
	p.scrollY = 0
}

// toBottom set's the pager's horizontal panning distance back to zero and
// vertical panning distance to the last viewport page.
func (p *Pager) ToBottom() {
	_, viewHeight := p.screen.Size()
	p.scrollX = 0
	p.scrollY = p.Pages() * viewHeight
}

// maxScrollX represents the pager's maximum horizontal scroll distance.
func (p *Pager) MaxScrollX() int {
	docWidth, _ := p.Size()
	viewWidth, _ := p.screen.Size()
	return docWidth - viewWidth
}

// maxScrollY represents the pager's maximum vertical scroll distance.
func (p *Pager) MaxScrollY() int {
	_, docHeight := p.Size()
	_, viewHeight := p.screen.Size()
	return docHeight - viewHeight
}

// size returns the width and height of the pager's underlying cell buffer
// document.
func (p *Pager) Size() (int, int) {
	height := len(p.doc.Cells) / p.doc.Width
	return p.doc.Width, height
}

// pages returns the number of times the pager's underlying cell buffer
// document can be split into viewport sized pages.
func (p *Pager) Pages() int {
	_, docHeight := p.Size()
	_, viewHeight := p.screen.Size()
	return docHeight / viewHeight
}
