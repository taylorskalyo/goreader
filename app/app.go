package app

import (
	"github.com/gdamore/tcell/v2"
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
	Run() int
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
func NewApp(b *epub.Rootfile, p nav.PageNavigator, chapter int) Application {
	return &app{pager: p, book: b, chapter: chapter, exitSignal: make(chan bool, 1)}
}

// Run opens a book, renders its contents within the pager, and polls for
// terminal events until an error occurs or an exit event is detected.
func (a *app) Run() int {
	var screen tcell.Screen

	if screen, a.err = initScreen(); a.err != nil {
		return 0
	}
	defer screen.Fini()
	a.pager.SetScreen(screen)

	keymap, chmap := initNavigationKeys(a)

	if a.err = a.openChapter(); a.err != nil {
		return 0
	}

	for {
		select {
		case <-a.exitSignal:
			return a.chapter
		default:
		}

		a.pager.Draw()

		ev := screen.PollEvent()
		switch ev := ev.(type) {
		case *tcell.EventKey:
			if action, ok := keymap[ev.Key()]; ok {
				action()
			} else if action, ok := chmap[ev.Rune()]; ok {
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

func initScreen() (tcell.Screen, error) {
	screen, err := tcell.NewScreen()
	if err == nil {
		err = screen.Init()
	}

	return screen, err
}

func initNavigationKeys(a Application) (map[tcell.Key]func(), map[rune]func()) {
	keymap := map[tcell.Key]func(){
		// Pager
		tcell.KeyDown:  a.PageNavigator().ScrollDown,
		tcell.KeyUp:    a.PageNavigator().ScrollUp,
		tcell.KeyRight: a.PageNavigator().ScrollRight,
		tcell.KeyLeft:  a.PageNavigator().ScrollLeft,

		// Navigation
		tcell.KeyEsc: a.Exit,
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
