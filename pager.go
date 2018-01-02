package main

import termbox "github.com/nsf/termbox-go"

type pager struct {
	scrollX int
	scrollY int
	cb      cellbuf
}

// draw displays a pager's cell buffer in the terminal.
func (p pager) draw() error {
	termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)

	_, height := termbox.Size()
	for y := 0; y < height; y++ {
		for x := 0; x < p.cb.width; x++ {
			index := (y+p.scrollY)*p.cb.width + x
			if index >= len(p.cb.cells) || index <= 0 {
				continue
			}
			cell := p.cb.cells[index]

			// Calling SetCell with coordinates outside of the terminal window
			// results in a no-op.
			termbox.SetCell(x+p.scrollX, y, cell.Ch, cell.Fg, cell.Bg)
		}
	}

	return termbox.Flush()
}

// scrollDown pans the pager's viewport down, without exceeding the underlying
// cell buffer document's boundaries.
func (p *pager) scrollDown() bool {
	if p.scrollY < p.maxScrollY() {
		p.scrollY++
		return true
	}

	return false
}

// scrollUp pans the pager's viewport up, without exceeding the underlying cell
// buffer document's boundaries.
func (p *pager) scrollUp() bool {
	if p.scrollY > 0 {
		p.scrollY--
		return true
	}

	return false
}

// scrollLeft pans the pager's viewport left, without exceeding the underlying
// cell buffer document's boundaries.
func (p *pager) scrollLeft() bool {
	if p.scrollX > -p.maxScrollX() {
		p.scrollX--
		return true
	}

	return false
}

// scrollRight pans the pager's viewport right, without exceeding the
// underlying cell buffer document's boundaries.
func (p *pager) scrollRight() bool {
	if p.scrollX < 0 {
		p.scrollX++
		return true
	}

	return false
}

// pageDown pans the pager's viewport down by a full page, without exceeding
// the underlying cell buffer document's boundaries.
func (p *pager) pageDown() bool {
	_, viewHeight := termbox.Size()
	if p.scrollY < p.maxScrollY() {
		p.scrollY += viewHeight
		return true
	}

	return false
}

// pageUp pans the pager's viewport up by a full page, without exceeding the
// underlying cell buffer document's boundaries.
func (p *pager) pageUp() bool {
	_, viewHeight := termbox.Size()
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
func (p *pager) toTop() {
	p.scrollX = 0
	p.scrollY = 0
}

// toBottom set's the pager's horizontal panning distance back to zero and
// vertical panning distance to its maximum value.
func (p *pager) toBottom() {
	_, viewHeight := termbox.Size()
	p.scrollX = 0
	p.scrollY = p.pages() * viewHeight
}

// maxScrollX represents the pager's maximum horizontal scroll distance.
func (p pager) maxScrollX() int {
	docWidth, _ := p.size()
	viewWidth, _ := termbox.Size()
	return docWidth - viewWidth
}

// maxScrollX represents the pager's maximum vertical scroll distance.
func (p pager) maxScrollY() int {
	_, docHeight := p.size()
	_, viewHeight := termbox.Size()
	return docHeight - viewHeight
}

// size returns the width and height of the pager's underlying cell buffer
// document.
func (p pager) size() (int, int) {
	height := len(p.cb.cells) / p.cb.width
	return p.cb.width, height
}

// pages returns the number of times the pager's underlying cell buffer
// document can be split into viewport sized pages.
func (p pager) pages() int {
	_, docHeight := p.size()
	_, viewHeight := termbox.Size()
	return docHeight / viewHeight
}