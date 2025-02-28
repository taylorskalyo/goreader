package main

import (
	"errors"
	"log/slog"
	"os"

	"github.com/taylorskalyo/goreader/epub"
	"github.com/taylorskalyo/goreader/views"
)

var (
	errUsage = errors.New("exit with error")
)

func main() {
	if err := run(); err != nil {
		if err != errUsage {

			// Assume non-usage errors are fatal.
			slog.Error("Encountered fatal error.", "error", err)
		}

		os.Exit(1)
	}

	os.Exit(0)
}

func run() error {
	app := views.NewApplication()
	if err := app.Configure(); err != nil {
		return err
	}

	if len(os.Args) <= 1 {
		app.PrintUsage()
		return errUsage
	} else if os.Args[1] == "-h" {
		app.PrintHelp()
		return nil
	}

	rc, err := epub.OpenReader(os.Args[1])
	if err != nil {
		return err
	}
	defer rc.Close()

	go app.QueueUpdate(func() {
		app.OpenBook(rc.DefaultRendition())
	})

	return app.Run()
}
