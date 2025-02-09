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
	Forward() error
	Back() error
	NextChapter() error
	PrevChapter() error
	GotoChapter(int) error

	PageNavigator() nav.PageNavigator
	Exit() error
	Run() error
}

// app is used to store the current state of the application.
type app struct {
	bookRC   *epub.ReadCloser
	pager    nav.PageNavigator
	progress state.Progress

	exitSignal bool
}

// NewApp creates an App
func NewApp(rc *epub.ReadCloser, p nav.PageNavigator) Application {
	return &app{pager: p, bookRC: rc, exitSignal: false}
}

// Run opens a book, renders its contents within the pager, and polls for
// terminal events until an error occurs or an exit event is detected.
func (a *app) Run() error {
	screen, err := initScreen()
	if err != nil {
		return err
	}
	defer screen.Fini()
	a.pager.SetScreen(screen)

	keymap, chmap := initNavigationKeys(a)

	if err = a.onStart(); err != nil {
		return err
	}

	for {
		if a.exitSignal {
			return a.onExit()
		}

		a.pager.Draw()

		ev := screen.PollEvent()
		switch ev := ev.(type) {
		case *tcell.EventKey:
			if action, ok := keymap[ev.Key()]; ok {
				err = action()
			} else if action, ok := chmap[ev.Rune()]; ok {
				err = action()
			}
			if err != nil {
				return err
			}
		}
	}
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

func initNavigationKeys(a Application) (map[tcell.Key]func() error, map[rune]func() error) {
	keymap := map[tcell.Key]func() error{
		// Pager
		tcell.KeyDown:  a.PageNavigator().ScrollDown,
		tcell.KeyUp:    a.PageNavigator().ScrollUp,
		tcell.KeyRight: a.PageNavigator().ScrollRight,
		tcell.KeyLeft:  a.PageNavigator().ScrollLeft,

		// Navigation
		tcell.KeyEsc: a.Exit,
	}
	chmap := map[rune]func() error{
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
func (a *app) Exit() error {
	a.exitSignal = true

	return nil
}

// openChapter opens the given chapter and renders it within the pager.
func (a *app) openChapter(chapter int) error {
	if a.progress.Chapter > len(a.book().Spine.Itemrefs) || a.progress.Chapter < 0 {
		// TODO: log warning
		return nil
	}

	a.progress.Chapter = chapter

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
func (a *app) Forward() error {
	if a.pager.PageDown() || a.progress.Chapter >= len(a.book().Spine.Itemrefs)-1 {
		return nil
	}

	// We reached the bottom.
	if err := a.NextChapter(); err != nil {
		return err
	}

	return a.pager.ToTop()
}

// Back pages up or opens the previous chapter.
func (a *app) Back() error {
	if a.pager.PageUp() || a.progress.Chapter <= 0 {
		return nil
	}

	// We reached the top.
	if err := a.PrevChapter(); err != nil {
		return err
	}

	return a.pager.ToBottom()
}

// GotoChapter sets the current chapter and opens it.
func (a *app) GotoChapter(chapter int) error {
	if chapter > len(a.book().Spine.Itemrefs) || a.progress.Chapter < 0 {
		return nil
	}

	if err := a.openChapter(chapter); err != nil {
		return err
	}

	return a.pager.ToTop()
}

// nextChapter opens the next chapter.
func (a *app) NextChapter() error {
	return a.GotoChapter(a.progress.Chapter + 1)
}

// prevChapter opens the previous chapter.
func (a *app) PrevChapter() error {
	return a.GotoChapter(a.progress.Chapter - 1)
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
	if err := a.openChapter(a.progress.Chapter); err != nil {
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
