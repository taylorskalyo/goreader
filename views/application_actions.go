package views

import (
	"context"
	"fmt"

	"github.com/taylorskalyo/goreader/config"
)

type actions map[config.Action]func()

func (app *Application) initActions() {
	app.actions = actions{
		config.ActionExit:            app.Stop,
		config.ActionUp:              app.Up,
		config.ActionDown:            app.Down,
		config.ActionLeft:            app.Left,
		config.ActionRight:           app.Right,
		config.ActionBackward:        app.Backward,
		config.ActionForward:         app.Forward,
		config.ActionTop:             app.Top,
		config.ActionBottom:          app.Bottom,
		config.ActionChapterNext:     app.ChapterNext,
		config.ActionChapterPrevious: app.ChapterPrevious,
	}

	// Sanity check to make sure we handle all of the configurable actions.
	for action, name := range config.ActionNames {
		if _, ok := app.actions[action]; !ok {
			panic(fmt.Sprintf("unhandled action \"%s\"", name))
		}
	}
}

func (app *Application) Up() {
	r, c := app.text.GetScrollOffset()
	app.text.ScrollTo(r-1, c)
	app.updateHeader()
}

func (app *Application) Down() {
	r, c := app.text.GetScrollOffset()
	app.text.ScrollTo(r+1, c)
	app.updateHeader()
}

func (app *Application) Left() {
	r, c := app.text.GetScrollOffset()
	app.text.ScrollTo(r, c-1)
}

func (app *Application) Right() {
	r, c := app.text.GetScrollOffset()
	app.text.ScrollTo(r, c+1)
}

func (app *Application) Backward() {
	_, _, _, height := app.text.GetRect()
	r, c := app.text.GetScrollOffset()
	if r > 0 {
		app.text.ScrollTo(r-height, c)
	} else {
		// At top of page, go to previous chapter
		prev := app.progress.Chapter - 1
		if prev >= 0 {
			app.gotoChapter(prev)
			app.text.ScrollTo(app.linecount-height, 0)
		}
	}

	app.updateHeader()
}

func (app *Application) Forward() {
	_, _, _, height := app.text.GetRect()
	r, c := app.text.GetScrollOffset()
	if r+height <= app.linecount {
		app.text.ScrollTo(r+height, c)
	} else {
		// At bottom of page, go to next chapter
		next := app.progress.Chapter + 1
		total := len(app.book.Spine.Itemrefs)
		if next < total {
			app.gotoChapter(next)
			app.text.ScrollToBeginning()
		}
	}

	app.updateHeader()
}

func (app *Application) Bottom() {
	_, _, _, height := app.text.GetRect()
	app.text.ScrollTo(app.linecount-height, 0)
	app.updateHeader()
}

func (app *Application) Top() {
	app.text.ScrollToBeginning()
	app.updateHeader()
}

func (app *Application) ChapterNext() {
	app.gotoChapter(app.progress.Chapter + 1)
	app.text.ScrollToBeginning()
	app.updateHeader()
}

func (app *Application) ChapterPrevious() {
	app.gotoChapter(app.progress.Chapter - 1)
	app.text.ScrollToBeginning()
	app.updateHeader()
}

func (app *Application) gotoChapter(n int) {
	total := len(app.book.Spine.Itemrefs)
	if n >= total || n < 0 {
		return
	}

	app.text.SetText("")
	app.progress.Chapter = n

	err := app.renderer.RenderChapter(context.TODO(), n, app.text)
	if err != nil {
		app.error("load chapter", err)
		return
	}

	app.linecount = app.text.GetWrappedLineCount()
}
