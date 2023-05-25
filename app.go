package main

import (
	termbox "github.com/nsf/termbox-go"
	"github.com/taylorskalyo/goreader/epub"
)

// app is used to store the current state of the application.
type app struct {
	pager   pager
	book    *epub.Rootfile
	chapter int

	err error
}

// run opens a book, renders its contents within the pager, and polls for
// terminal events until an error occurs or an exit event is detected.
func (a *app) run() {
	if a.err = termbox.Init(); a.err != nil {
		return
	}
	defer termbox.Flush()
	defer termbox.Close()

	keymap := map[termbox.Key]func(){
		// Pager
		termbox.KeyArrowDown:  a.pager.scrollDown,
		termbox.KeyArrowUp:    a.pager.scrollUp,
		termbox.KeyArrowRight: a.pager.scrollRight,
		termbox.KeyArrowLeft:  a.pager.scrollLeft,

		// Navigation
		termbox.KeyEsc: a.exit,
	}

	chmap := map[rune]func(){
		// Pager
		'j': a.pager.scrollDown,
		'k': a.pager.scrollUp,
		'h': a.pager.scrollLeft,
		'l': a.pager.scrollRight,
		'g': a.pager.toTop,
		'G': a.pager.toBottom,

		// Navigation
		'q': a.exit,
		'f': a.forward,
		'b': a.back,
		'L': a.nextChapter,
		'H': a.prevChapter,
	}

	if a.err = a.openChapter(); a.err != nil {
		return
	}

	for {
		if a.err = a.pager.draw(); a.err != nil {
			return
		}

		ev := termbox.PollEvent()
		switch ev.Type {
		case termbox.EventKey:
			if action, ok := keymap[ev.Key]; ok {
				action()
			} else if action, ok := chmap[ev.Ch]; ok {
				action()
			}
		}

		if a.err != nil {
			break
		}
	}
}

// exit requests app termination.
func (a *app) exit() {
	a.err = exitRequest
}

// openChapter opens the current chapter and renders it within the pager.
func (a *app) openChapter() error {
	f, err := a.book.Spine.Itemrefs[a.chapter].Open()
	if err != nil {
		return err
	}
	doc, err := parseText(f, a.book.Manifest.Items)
	if err != nil {
		return err
	}
	a.pager.doc = doc

	return nil
}

// forward pages down or opens the next chapter.
func (a *app) forward() {
	if a.pager.pageDown() || a.chapter >= len(a.book.Spine.Itemrefs)-1 {
		return
	}

	// We reached the bottom.
	if a.nextChapter(); a.err == nil {
		a.pager.toTop()
	}
}

// back pages up or opens the previous chapter.
func (a *app) back() {
	if a.pager.pageUp() || a.chapter <= 0 {
		return
	}

	// We reached the top.
	if a.prevChapter(); a.err == nil {
		a.pager.toBottom()
	}
}

// nextChapter opens the next chapter.
func (a *app) nextChapter() {
	if a.chapter >= len(a.book.Spine.Itemrefs)-1 {
		return
	}

	a.chapter++
	if a.err = a.openChapter(); a.err == nil {
		a.pager.toTop()
	}
}

// prevChapter opens the previous chapter.
func (a *app) prevChapter() {
	if a.chapter <= 0 {
		return
	}

	a.chapter--
	if a.err = a.openChapter(); a.err == nil {
		a.pager.toTop()
	}
}
