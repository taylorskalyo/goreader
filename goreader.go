package main

import (
	"archive/zip"
	"fmt"
	"os"

	"github.com/taylorskalyo/goreader/epub"
)

func main() {
	if len(os.Args) <= 1 {
		fmt.Fprintln(os.Stderr, "You must specify a file")
		os.Exit(1)
	}

	rc, err := epub.OpenReader(os.Args[1])
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
	defer rc.Close()
	book := rc.Rootfiles[0]

	a := app{book: book}
	if err := a.run(); err != nil {
		os.Exit(1)
	}
}
