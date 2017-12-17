package main

import (
	"archive/zip"
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
		var msg string
		switch err {
		case zip.ErrFormat, zip.ErrAlgorithm, zip.ErrChecksum:
			msg = fmt.Sprintf("cannot unzip contents: %s", err.Error())
		default:
			msg = err.Error()
		}
		fmt.Fprintf(os.Stderr, "Unable to open epub: %s\n", msg)
		os.Exit(1)
	}
	defer book.Close()

	title := views.NewTextBar()
	title.SetStyle(tcell.StyleDefault.
		Foreground(tcell.ColorWhite).
		Background(tcell.ColorBlack))
	title.SetCenter(book.Rootfiles[0].Title, tcell.StyleDefault)
	page := views.NewTextArea()

	f, err := book.Rootfiles[0].Spine.Itemrefs[0].Open()
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
	text, err := parseText(f)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
	page.SetContent(text.String())
	page.SetStyle(tcell.StyleDefault.
		Foreground(tcell.ColorBlack).
		Background(tcell.ColorWhite))

	box.SetOrientation(views.Vertical)
	box.AddWidget(title, 0)
	box.AddWidget(page, 1)

	app.SetRootWidget(box)
	if err := app.Run(); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}
