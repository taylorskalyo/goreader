package views

import (
	"errors"
	"fmt"
	"log/slog"
	"os"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/taylorskalyo/goreader/config"
	"github.com/taylorskalyo/goreader/epub"
	"github.com/taylorskalyo/goreader/render"
	"github.com/taylorskalyo/goreader/state"
)

var (
	ErrNoExitAction = errors.New("Exit action not configured")
	ErrNoArgs       = errors.New("missing arguments")
)

type Application struct {
	*tview.Application

	config  *config.Config
	actions actions

	progress state.Progress
	book     epub.Package

	linecount int
	renderer  render.Renderer

	text      *tview.TextView
	header    *tview.TextView
	footer    *tview.TextView
	container *tview.Flex
}

func NewApplication() *Application {
	app := &Application{
		Application: tview.NewApplication(),
	}
	app.initActions()

	app.text = tview.NewTextView().
		SetDynamicColors(true).
		SetWrap(false).
		SetChangedFunc(func() {
			app.Draw()
		})
	app.text.SetBorderPadding(1, 1, 0, 0)
	app.SetInputCapture(app.inputHandler)

	app.header = tview.NewTextView().
		SetTextAlign(tview.AlignCenter).
		SetWrap(false).
		SetScrollable(false)

	app.footer = tview.NewTextView().
		SetTextAlign(tview.AlignCenter).
		SetWrap(false).
		SetScrollable(false)

	app.container = tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(app.header, 1, 0, false).
		AddItem(app.text, 0, 1, true).
		AddItem(app.footer, 1, 0, false)

	root := tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(app.container, 80, 2, false).
		AddItem(nil, 0, 1, false)

	app.SetRoot(root, true).SetFocus(app.text).EnableMouse(true)

	return app
}

func (app *Application) Run() error {
	if len(os.Args) <= 1 {
		return ErrNoArgs
	}

	if err := app.configure(); err != nil {
		return err
	}

	rc, err := epub.OpenReader(os.Args[1])
	if err != nil {
		return err
	}
	defer rc.Close()

	app.book = rc.DefaultRendition().Package
	app.renderer = render.New(&app.book)
	app.renderer.SetTheme(app.config.Theme)
	app.footer.SetText(app.book.Title)

	go app.QueueUpdate(func() {
		app.loadProgress()
		app.gotoChapter(app.progress.Chapter)
		app.setPosition(app.progress.Position)
		app.updateHeader()
	})

	return app.Application.Run()
}

func (app *Application) Stop() {
	app.progress.Position = app.getPosition()
	app.progress.Title = app.book.Title

	if err := state.StoreProgress(app.bookID(), app.progress); err != nil {
		app.error("save progress", err)
	}

	app.Application.Stop()
}

func (app *Application) configure() error {
	var err error
	if app.config, err = config.Load(); err != nil {
		app.error("load config", err)
		app.warn("Using default config.")
	}

	// Refuse to run if there's no exit action configured.
	hasExit := false
	for _, action := range app.config.Keybindings {
		if action == config.ActionExit {
			hasExit = true
			break
		}
	}

	if !hasExit {
		app.warn(fmt.Sprintf("Make sure you configure a keybinding for the Exit action in %s", config.ConfigFile))
		return ErrNoExitAction
	}

	return nil
}

func (app *Application) loadProgress() {
	var err error
	app.progress, err = state.LoadProgress(app.bookID())

	if err != nil && !os.IsNotExist(err) {
		app.error("load progress", err)
	}
}

func (app *Application) setPosition(pos float64) {
	app.text.ScrollTo(int(pos*float64(app.linecount)), 0)
}

func (app *Application) getPosition() float64 {
	r, _ := app.text.GetScrollOffset()

	return float64(r) / float64(app.linecount)
}

func (app *Application) updateHeader() {
	r, _ := app.text.GetScrollOffset()
	if r < 1 {
		r = 1
	} else if r > app.linecount {
		r = app.linecount
	}

	_, _, _, height := app.text.GetRect()
	pages := (height + app.linecount - 1) / height
	cur := int(float64(r)/float64(height)+0.5) + 1 // closest page relative to offset

	app.header.SetText(fmt.Sprintf("CHAPTER %d - %d OF %d", app.progress.Chapter+1, cur, pages))
}

func (app *Application) inputHandler(event *tcell.EventKey) *tcell.EventKey {
	if event == nil {
		return nil
	}

	chord := config.KeyChordFromEvent(*event)
	if action, ok := app.config.Keybindings[chord]; ok {
		if fn, ok := app.actions[action]; ok {
			fn()
		}
	}

	// Ignore unhandled bindings.
	return nil
}

func (app Application) warn(msg string, args ...any) {
	app.suspend(func() {
		slog.Warn(msg, args...)
	})
}

func (app Application) error(msg string, err error, args ...any) {
	args = append(args, "error", err)
	app.suspend(func() {
		slog.Error(fmt.Sprintf("Failed to %s.", msg), args...)
	})
}

func (app Application) suspend(fn func()) {
	if !app.Suspend(fn) {
		fn()
	}
}

// > The EPUB creator MUST provide an identifier that is unique to one and only
// > one EPUB publication [1]
//
// [1]: https://www.w3.org/TR/epub/#dfn-dc-identifier
func (app Application) bookID() string {
	if id := app.book.Identifier; id.Content != "" {
		return fmt.Sprintf("%s:%s", id.Scheme, id.Content)
	} else {
		app.warn("Book is missing identifier (e.g. ISBN); loading and saving reading progress may not work as expected.", "scheme", id.Scheme, "content", id.Content)
	}

	return fmt.Sprintf("title:%s", app.book.Title)
}
