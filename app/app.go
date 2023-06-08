package app

import (
	termbox "github.com/nsf/termbox-go"
	"github.com/taylorskalyo/goreader/epub"
	"github.com/taylorskalyo/goreader/nav"
	"github.com/taylorskalyo/goreader/parse"
)

type Application interface {
	Forward()
	Back()
	NextChapter()
	PrevChapter()

	PageNavigator() nav.PageNavigator
	Exit()
	Run()
	Err() error
}

// app is used to store the current state of the application.
type app struct {
	book    *epub.Rootfile
	pager   nav.PageNavigator
	chapter int

	err error

	exitSignal chan bool
}

// NewApp creates an App
func NewApp(b *epub.Rootfile, p nav.PageNavigator) Application {
	return &app{pager: p, book: b, exitSignal: make(chan bool, 1)}
}

// Run opens a book, renders its contents within the pager, and polls for
// terminal events until an error occurs or an exit event is detected.
func (a *app) Run() {
	if a.err = termbox.Init(); a.err != nil {
		return
	}
	defer termbox.Flush()
	defer termbox.Close()

	keymap, chmap := initNavigationKeys(a)

	if a.err = a.openChapter(); a.err != nil {
		return
	}

MainLoop:
	for {
		select {
		case <-a.exitSignal:
			break MainLoop
		default:
		}

		if a.err = a.pager.Draw(); a.err != nil {
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
	}
}

func (a *app) Err() error {
	return a.err
}

func (a *app) PageNavigator() nav.PageNavigator {
	return a.pager
}

func initNavigationKeys(a Application) (map[termbox.Key]func(), map[rune]func()) {
	keymap := map[termbox.Key]func(){
		// Pager
		termbox.KeyArrowDown:  a.PageNavigator().ScrollDown,
		termbox.KeyArrowUp:    a.PageNavigator().ScrollUp,
		termbox.KeyArrowRight: a.PageNavigator().ScrollRight,
		termbox.KeyArrowLeft:  a.PageNavigator().ScrollLeft,

		// Navigation
		termbox.KeyEsc: a.Exit,
	}

	chmap := map[rune]func(){
		// PageNavigator
		'j': a.PageNavigator().ScrollDown,
		'k': a.PageNavigator().ScrollUp,
		'h': a.PageNavigator().ScrollLeft,
		'l': a.PageNavigator().ScrollRight,
		'g': a.PageNavigator().ToTop,
		'G': a.PageNavigator().ToBottom,

		// Navigation
		'q': a.Exit,
		'f': a.Forward,
		'b': a.Back,
		'L': a.NextChapter,
		'H': a.PrevChapter,
	}

	return keymap, chmap
}

// Exit requests app termination.
func (a *app) Exit() {
	a.exitSignal <- true
}

// openChapter opens the current chapter and renders it within the pager.
func (a *app) openChapter() error {
	f, err := a.book.Spine.Itemrefs[a.chapter].Open()
	if err != nil {
		return err
	}
	doc, err := parse.ParseText(f, a.book.Manifest.Items)
	if err != nil {
		return err
	}
	a.pager.SetDoc(doc)

	return nil
}

// Forward pages down or opens the next chapter.
func (a *app) Forward() {
	if a.pager.PageDown() || a.chapter >= len(a.book.Spine.Itemrefs)-1 {
		return
	}

	// We reached the bottom.
	if a.NextChapter(); a.err == nil {
		a.pager.ToTop()
	}
}

// Back pages up or opens the previous chapter.
func (a *app) Back() {
	if a.pager.PageUp() || a.chapter <= 0 {
		return
	}

	// We reached the top.
	if a.PrevChapter(); a.err == nil {
		a.pager.ToBottom()
	}
}

// nextChapter opens the next chapter.
func (a *app) NextChapter() {
	if a.chapter >= len(a.book.Spine.Itemrefs)-1 {
		return
	}

	a.chapter++
	if a.err = a.openChapter(); a.err == nil {
		a.pager.ToTop()
	}
}

// prevChapter opens the previous chapter.
func (a *app) PrevChapter() {
	if a.chapter <= 0 {
		return
	}

	a.chapter--
	if a.err = a.openChapter(); a.err == nil {
		a.pager.ToTop()
	}
}
