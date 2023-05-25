package main

import (
	"archive/zip"
	"errors"
	"fmt"
	"os"

	"github.com/taylorskalyo/goreader/epub"
)

var (
	// exitRequest is used as a return value from the main event loop to indicate that
	// the app should exit.
	exitRequest = errors.New("exit requested")
)

func main() {
	if len(os.Args) <= 1 {
		printUsage()
		os.Exit(1)
	}

	if os.Args[1] == "-h" || os.Args[1] == "--help" {
		printHelp()
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
	a.run()

	if a.err == nil || a.err == exitRequest {
		os.Exit(0)
	} else {
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Fprintln(os.Stderr, "goreader [epub_file]")
	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintln(os.Stderr, "-h		print keybindings")
}

func printHelp() {
	fmt.Fprintln(os.Stderr, "Key                  Action")
	fmt.Fprintln(os.Stderr)
	fmt.Fprintln(os.Stderr, "q                    Quit")
	fmt.Fprintln(os.Stderr, "k / Up arrow         Scroll up")
	fmt.Fprintln(os.Stderr, "j / Down arrow       Scroll down")
	fmt.Fprintln(os.Stderr, "h / Left arrow       Scroll left")
	fmt.Fprintln(os.Stderr, "l / Right arrow      Scroll right")
	fmt.Fprintln(os.Stderr, "b                    Previous page")
	fmt.Fprintln(os.Stderr, "f                    Next page")
	fmt.Fprintln(os.Stderr, "H                    Previous chapter")
	fmt.Fprintln(os.Stderr, "L                    Next chapter")
	fmt.Fprintln(os.Stderr, "g                    Top of chapter")
	fmt.Fprintln(os.Stderr, "G                    Bottom of chapter")
}
