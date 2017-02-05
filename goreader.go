package main

import (
	"fmt"
	"os"

	"github.com/gdamore/tcell"
	"github.com/gdamore/tcell/views"
	"github.com/taylorskalyo/goreader/epub"
)

type boxL struct {
	views.BoxLayout
}

var (
	box  = &boxL{}
	app  = &views.Application{}
	book *epub.ReadCloser
)

func (m *boxL) HandleEvent(ev tcell.Event) bool {
	switch ev := ev.(type) {
	case *tcell.EventKey:
		if ev.Key() == tcell.KeyEscape {
			app.Quit()
			return true
		}
	}
	return m.BoxLayout.HandleEvent(ev)
}

func main() {
	if len(os.Args) <= 1 {
		fmt.Fprintln(os.Stderr, "You must specify a file")
		os.Exit(1)
	}

	book, err := epub.OpenReader(os.Args[1])
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
	defer book.Close()

	title := views.NewTextBar()
	title.SetStyle(tcell.StyleDefault.
		Foreground(tcell.ColorBlack).
		Background(tcell.ColorYellow))
	title.SetCenter(book.Rootfiles[0].Title, tcell.StyleDefault)
	page := views.NewText()

	page.SetText("book contents")
	page.SetStyle(tcell.StyleDefault.
		Foreground(tcell.ColorBlack).
		Background(tcell.ColorWhite))

	page.SetAlignment(views.VAlignTop | views.HAlignLeft)

	box.SetOrientation(views.Vertical)
	box.AddWidget(title, 0)
	box.AddWidget(page, 1)

	app.SetRootWidget(box)
	if err := app.Run(); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}
