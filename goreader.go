package main

import (
	"archive/zip"
	"fmt"
	"os"

	termbox "github.com/nsf/termbox-go"
	"github.com/taylorskalyo/goreader/epub"
)

type pager struct {
	scrollX int
	scrollY int
	cb      cellbuf
	chapter int
}

func main() {
	if len(os.Args) <= 1 {
		fmt.Fprintln(os.Stderr, "You must specify a file")
		os.Exit(1)
	}

	rc, err := epub.OpenReader(os.Args[1])
	if err != nil {
		var msg string
		switch err {
		case zip.ErrFormat, zip.ErrAlgorithm, zip.ErrChecksum:
			msg = fmt.Sprintf("cannot unzip contents: %s", err.Error())
		default:
			msg = err.Error()
		}
		fmt.Fprintf(os.Stderr, "Unable to open epub: %s\n", msg)
		os.Exit(1)
	}
	defer rc.Close()
	book := rc.Rootfiles[0]

	f, err := book.Spine.Itemrefs[0].Open()
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
	if err := termbox.Init(); err != nil {
		os.Exit(1)
	}
	cb, err := parseText(f)
	if err != nil {
		os.Exit(1)
	}
	p := pager{cb: cb, scrollX: 0, scrollY: 0}
	if err := p.run(book); err != nil {
		os.Exit(1)
	}
	termbox.Flush()
	termbox.Close()
}

func (p pager) copyToTerm() {
	_, height := termbox.Size()
	for y := 0; y < height; y++ {
		for x := 0; x < p.cb.width; x++ {
			index := (y+p.scrollY)*p.cb.width + x
			if index >= len(p.cb.cells) || index <= 0 {
				continue
			}
			cell := p.cb.cells[index]
			termbox.SetCell(x+p.scrollX, y, cell.Ch, cell.Fg, cell.Bg)
		}
	}
}

func (p *pager) draw() error {
	termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
	p.copyToTerm()
	return termbox.Flush()
}

func (p *pager) run(book *epub.Rootfile) error {
	for {
		if err := p.draw(); err != nil {
			return err
		}
		switch ev := termbox.PollEvent(); ev.Type {
		case termbox.EventKey:
			_, height := termbox.Size()
			switch ev.Key {
			case termbox.KeyEsc:
				return nil
			default:
				switch ev.Ch {
				case 'q':
					return nil
				case 'j':
					if p.scrollY < len(p.cb.cells)/p.cb.width-height {
						p.scrollY++
					}
				case 'k':
					if p.scrollY > 0 {
						p.scrollY--
					}
				case 'h':
					if p.scrollX > -p.cb.width {
						p.scrollX--
					}
				case 'l':
					if p.scrollX < 0 {
						p.scrollX++
					}
				case 'f':
					if p.scrollY < len(p.cb.cells)/p.cb.width-height {
						p.scrollY += height
					}
				case 'b':
					if p.scrollY > 0 {
						p.scrollY -= height
					}
				case 'L':
					if p.chapter < len(book.Spine.Itemrefs)-1 {
						p.scrollY = 0
						p.scrollX = 0
						p.chapter++
						f, err := book.Spine.Itemrefs[p.chapter].Open()
						if err != nil {
							return err
						}
						cb, err := parseText(f)
						if err != nil {
							return err
						}
						p.cb = cb
					}
				case 'H':
					if p.chapter > 0 {
						p.scrollY = 0
						p.scrollX = 0
						p.chapter--
						f, err := book.Spine.Itemrefs[p.chapter].Open()
						if err != nil {
							return err
						}
						cb, err := parseText(f)
						if err != nil {
							return err
						}
						p.cb = cb
					}
				}
			}
		}
	}
}
