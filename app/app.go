package app

import (
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/taylorskalyo/goreader/epub"
	"github.com/taylorskalyo/goreader/nav"
	"github.com/taylorskalyo/goreader/parse"
	"github.com/taylorskalyo/goreader/state"
)

type Application interface {
	Forward()
	Back()
	NextChapter()
	PrevChapter()
	GotoChapter(int) error

	PageNavigator() nav.PageNavigator
	Exit()
	Run()
	Err() error
}

// app is used to store the current state of the application.
type app struct {
	bookRC   *epub.ReadCloser
	pager    nav.PageNavigator
	progress state.Progress

	err error

	exitSignal chan bool
}

// NewApp creates an App
func NewApp(rc *epub.ReadCloser, p nav.PageNavigator) Application {
	return &app{pager: p, bookRC: rc, exitSignal: make(chan bool, 1)}
}

// Run opens a book, renders its contents within the pager, and polls for
// terminal events until an error occurs or an exit event is detected.
func (a *app) Run() {
	var screen tcell.Screen

	if screen, a.err = initScreen(); a.err != nil {
		return
	}
	defer screen.Fini()
	a.pager.SetScreen(screen)

	keymap, chmap := initNavigationKeys(a)

	if a.err = a.onStart(); a.err != nil {
		return
	}

	for {
		select {
		case <-a.exitSignal:
			a.err = a.onExit()
			return
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

func (a app) Err() error {
	return a.err
}

func (a app) PageNavigator() nav.PageNavigator {
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

// GotoChapter sets the current chapter and opens it.
func (a *app) GotoChapter(chapter int) error {
	a.progress.Chapter = chapter
	return a.openChapter()
}

// openChapter opens the current chapter and renders it within the pager.
func (a *app) openChapter() error {
	if a.progress.Chapter > len(a.book().Spine.Itemrefs) || a.progress.Chapter < 0 {
		// TODO: log warning
		a.progress.Chapter = 0
	}

	f, err := a.book().Spine.Itemrefs[a.progress.Chapter].Open()
	if err != nil {
		return err
	}

	doc, err := parse.ParseText(f, a.book().Manifest.Items)
	if err != nil {
		return err
	}
	a.pager.SetDoc(doc)

	return nil
}

// Forward pages down or opens the next chapter.
func (a *app) Forward() {
	if a.pager.PageDown() || a.progress.Chapter >= len(a.book().Spine.Itemrefs)-1 {
		return
	}

	// We reached the bottom.
	if a.NextChapter(); a.err == nil {
		a.pager.ToTop()
	}
}

// Back pages up or opens the previous chapter.
func (a *app) Back() {
	if a.pager.PageUp() || a.progress.Chapter <= 0 {
		return
	}

	// We reached the top.
	if a.PrevChapter(); a.err == nil {
		a.pager.ToBottom()
	}
}

// nextChapter opens the next chapter.
func (a *app) NextChapter() {
	if a.progress.Chapter >= len(a.book().Spine.Itemrefs)-1 {
		return
	}

	a.progress.Chapter++
	if a.err = a.openChapter(); a.err == nil {
		a.pager.ToTop()
	}
}

// prevChapter opens the previous chapter.
func (a *app) PrevChapter() {
	if a.progress.Chapter <= 0 {
		return
	}

	a.progress.Chapter--
	if a.err = a.openChapter(); a.err == nil {
		a.pager.ToTop()
	}
}

func (a app) book() *epub.Rootfile {
	return a.bookRC.DefaultRendition()
}

// bookID is the books unique identifier.
//
// > The EPUB creator MUST provide an identifier that is unique to one and only
// > one EPUB publication [1]
//
// [1]: https://www.w3.org/TR/epub/#dfn-dc-identifier
func (a app) bookID() string {
	if id := a.book().Identifier; id.Content != "" {
		return strings.Join([]string{id.Scheme, id.Content}, ":")
	}

	// TOOD: Log warning
	return ""
}

// onStart is run when the application starts.
func (a *app) onStart() error {
	a.progress = state.Load(a.bookID())
	if err := a.openChapter(); err != nil {
		return err
	}

	a.pager.SetPosition(a.progress.Position)

	return nil
}

// onExit is run when the application exits.
func (a app) onExit() error {
	a.progress.Position = a.pager.Position()
	a.progress.Title = a.book().Title

	return state.Store(a.bookID(), a.progress)
}
