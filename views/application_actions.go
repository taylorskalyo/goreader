package views

import (
	"context"
	"fmt"

	"github.com/taylorskalyo/goreader/config"
)

type actions map[config.Action]func()

// initActions initializes configurable actions.
func (app *Application) initActions() {
	app.actions = actions{
		config.ActionExit:            app.Stop,
		config.ActionUp:              app.Up,
		config.ActionDown:            app.Down,
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

// Up scrolls the application viewport up by one line.
func (app *Application) Up() {
	r, c := app.text.GetScrollOffset()
	if r > 0 {
		app.text.ScrollTo(r-1, c)
	}
}

// Down scrolls the application viewport down by one line.
func (app *Application) Down() {
	r, c := app.text.GetScrollOffset()
	_, _, _, height := app.text.GetRect()
	if r < app.linecount-height {
		app.text.ScrollTo(r+1, c)
	}
}

// Backward navigates backward by a page within the viewport. If at the top of
// a chapter, the viewport will navigate to the bottom of the previous chapter.
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
}

// Forward navigates forward by a page within the viewport. If at the bottom of
// a chapter, the viewport will navigate to the top of the next chapter.
func (app *Application) Forward() {
	_, _, _, height := app.text.GetRect()
	r, c := app.text.GetScrollOffset()

	if r < app.linecount-height {
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
}

// Bottom navigates to the bottom of the current chapter.
func (app *Application) Bottom() {
	_, _, _, height := app.text.GetRect()
	app.text.ScrollTo(app.linecount-height, 0)
}

// Top navigates to the top of the current chapter.
func (app *Application) Top() {
	app.text.ScrollToBeginning()
}

// ChapterNext navigates to the next chapter.
func (app *Application) ChapterNext() {
	app.gotoChapter(app.progress.Chapter + 1)
	app.text.ScrollToBeginning()
}

// ChapterPrevious navigates to the previous chapter.
func (app *Application) ChapterPrevious() {
	app.gotoChapter(app.progress.Chapter - 1)
	app.text.ScrollToBeginning()
}

// gotoChapter navigates to a specific chapter.
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

	app.linecount = app.text.GetOriginalLineCount()
}
