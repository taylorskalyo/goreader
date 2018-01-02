package main

import (
	"archive/zip"
	"fmt"
	"os"

	termbox "github.com/nsf/termbox-go"
	"github.com/taylorskalyo/goreader/epub"
)

type app struct {
	pager   pager
	book    *epub.Rootfile
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

	a := app{book: book}
	if err := a.run(); err != nil {
		os.Exit(1)
	}
}

func (a *app) run() error {
	defer termbox.Flush()
	defer termbox.Close()

	if err := termbox.Init(); err != nil {
		return err
	}
	if err := a.openChapter(); err != nil {
		return err
	}

	for {
		if err := a.pager.draw(); err != nil {
			return err
		}
		switch ev := termbox.PollEvent(); ev.Type {
		case termbox.EventKey:
			switch ev.Key {
			case termbox.KeyEsc:
				return nil
			default:
				switch ev.Ch {
				case 'q':
					return nil
				case 'j':
					a.pager.scrollDown()
				case 'k':
					a.pager.scrollUp()
				case 'h':
					a.pager.scrollLeft()
				case 'l':
					a.pager.scrollRight()
				case 'f':
					if !a.pager.pageDown() {
						if err := a.nextChapter(); err != nil {
							return err
						}
					}
				case 'b':
					if !a.pager.pageUp() {
						if err := a.prevChapter(); err != nil {
							return err
						}
					}
				case 'L':
					if err := a.nextChapter(); err != nil {
						return err
					}
				case 'H':
					if err := a.prevChapter(); err != nil {
						return err
					}
				}
			}
		}
	}
}

func (a *app) openChapter() error {
	f, err := a.book.Spine.Itemrefs[a.chapter].Open()
	if err != nil {
		return err
	}
	cb, err := parseText(f)
	if err != nil {
		return err
	}
	a.pager.cb = cb

	return nil
}

func (a *app) nextChapter() error {
	if a.chapter < len(a.book.Spine.Itemrefs)-1 {
		a.chapter++
		if err := a.openChapter(); err != nil {
			return err
		}
		a.pager.toTop()
	}

	return nil
}

func (a *app) prevChapter() error {
	if a.chapter > 0 {
		a.chapter--
		if err := a.openChapter(); err != nil {
			return err
		}
		a.pager.toBottom()
	}

	return nil
}
