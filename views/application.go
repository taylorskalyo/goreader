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

// Application represents the application view.
type Application struct {
	*tview.Application

	config  *config.Config
	actions actions

	progress state.Progress
	book     *epub.Rootfile

	linecount int
	renderer  render.Renderer

	text      *tview.TextView
	header    *tview.TextView
	footer    *tview.TextView
	container *tview.Flex
}

// NewApplication returns an empty application.
func NewApplication() *Application {
	app := &Application{
		Application: tview.NewApplication(),
	}
	app.initActions()

	config := config.Default()
	app.config = &config

	app.text = tview.NewTextView().
		SetDynamicColors(true).
		SetWrap(false).
		SetChangedFunc(func() {
			app.Draw()
		})
	app.SetInputCapture(app.inputHandler)
	app.SetBeforeDrawFunc(app.beforeDraw)

	app.header = tview.NewTextView().
		SetTextAlign(tview.AlignCenter).
		SetWrap(false).
		SetScrollable(false)

	app.footer = tview.NewTextView().
		SetTextAlign(tview.AlignCenter).
		SetWrap(false).
		SetScrollable(false)

	app.header.SetBorderPadding(0, 1, 0, 0)
	app.footer.SetBorderPadding(1, 0, 0, 0)

	app.container = tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(app.header, 2, 0, false).
		AddItem(app.text, 0, 1, true).
		AddItem(app.footer, 2, 0, false)

	root := tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(app.container, 80, 0, false).
		AddItem(nil, 0, 1, false)

	app.SetRoot(root, true).SetFocus(app.text).EnableMouse(true)

	return app
}

// OpenBook loads the book contents into the application and navigates to the
// last-open page. It loads the first page if the book has not previously been
// read.
func (app *Application) OpenBook(book *epub.Rootfile) {
	app.book = book
	app.renderer = render.New(&app.book.Package)
	app.renderer.SetTheme(app.config.Theme)
	app.footer.SetText(app.book.Title)
	app.loadProgress()
	app.gotoChapter(app.progress.Chapter)
	app.setPosition(app.progress.Position)
}

// printHelp prints the configured keybindings to stderr.
func (app Application) PrintHelp() {
	fmt.Fprintf(os.Stderr, "Configured keybindings:\n\n%s\n", app.config.Keybindings)
}

// printUsage prints application usage to stderr.
func (app Application) PrintUsage() {
	fmt.Fprintln(os.Stderr, "Usage: goreader [epub file]")
	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintln(os.Stderr, "-h             print keybindings")
}

// Stop wraps tview.Application.Stop(). It saves reading progress then causes
// Run() to return.
func (app *Application) Stop() {
	app.progress.Position = app.getPosition()
	app.progress.Title = app.book.Title

	if err := state.StoreProgress(app.bookID(), app.progress); err != nil {
		app.error("save progress", err)
	}

	app.Application.Stop()
}

// configure loads the application configuration from a file. If the file does
// not exist or an error occurs, the default configuration will be used.
func (app *Application) Configure() error {
	var err error
	if app.config, err = config.Load(); err != nil {
		app.error("load config", err)
		app.warn("Using default config.")
	}

	// Rather than running with no way to stop, exit with an error when Exit
	// action is not configured.
	if !app.hasAction(config.ActionExit) {
		app.warn(fmt.Sprintf("No keybinding for Exit action. Edit %s to add keybinding for Exit action.", config.ConfigFile))
		return errors.New("Exit action not configured")
	}

	return nil
}

// hasAction checks to see if the application has an action configured.
func (app Application) hasAction(target config.Action) bool {
	for _, action := range app.config.Keybindings {
		if action == target {
			return true
		}
	}

	return false
}

// loadProgress loads the reading progress for the currently opened book.
func (app *Application) loadProgress() {
	var err error
	app.progress, err = state.LoadProgress(app.bookID())

	if err != nil && !os.IsNotExist(err) {
		app.error("load progress", err)
	}
}

// setPosition seeks to a given position within the open chapter.
func (app *Application) setPosition(pos float64) {
	app.text.ScrollTo(int(pos*float64(app.linecount)), 0)
}

// getPosition returns the position within the open chapter.
func (app *Application) getPosition() float64 {
	r, _ := app.text.GetScrollOffset()

	return float64(r) / float64(app.linecount)
}

// beforeDraw is executed before every Draw() call of the application.
func (app *Application) beforeDraw(s tcell.Screen) bool {
	if app.book != nil {
		app.updateHeader()
	}

	return false
}

// updateHeader populates the application's header window.
func (app *Application) updateHeader() {
	r, _ := app.text.GetScrollOffset()
	_, _, _, height := app.text.GetRect()

	// tview.TextView.Draw() keeps the vertical offset within these bounds.
	// However, updateHeader gets called before tview.TextView.Draw().
	if r < 0 {
		r = 0
	} else if r > app.linecount-height {
		r = app.linecount - height
	}

	pages := (height + app.linecount - 1) / height
	cur := (height+r-1)/height + 1

	// Try to find chapter title.
	ref := app.book.Spine.Itemrefs[app.progress.Chapter]
	if title := app.book.ItemName(ref.HREF); title != "" {
		app.header.SetText(fmt.Sprintf("%s • %d OF %d", title, cur, pages))
	} else {
		app.header.SetText(fmt.Sprintf("%d OF %d", cur, pages))
	}
}

// inputHandler intercepts input events. If the application has an action
// configured for an event, it will be triggered here.
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

// warn suspends the application and then writes warning messages to stderr.
func (app Application) warn(msg string, args ...any) {
	app.suspend(func() {
		slog.Warn(msg, args...)
	})
}

// error suspends the application and then writes error messages to stderr.
// Callers must supply an error message.
func (app Application) error(msg string, err error, args ...any) {
	args = append(args, "error", err)
	app.suspend(func() {
		slog.Error(fmt.Sprintf("Failed to %s.", msg), args...)
	})
}

// suspend calls tview.Application.Suspend() with the given function. Unlike
// tview.Application.Suspend(), the function will still be invoked if the
// application is already suspended.
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
	id := app.book.Identifier
	if id.Content != "" {
		return fmt.Sprintf("%s:%s", id.Scheme, id.Content)
	}

	app.warn("Book is missing identifier (e.g. ISBN); loading and saving reading progress may not work as expected.", "scheme", id.Scheme, "content", id.Content)

	return fmt.Sprintf("title:%s", app.book.Title)
}
