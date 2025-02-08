package main

import (
	"archive/zip"
	"fmt"
	"os"

	"github.com/taylorskalyo/goreader/app"
	"github.com/taylorskalyo/goreader/epub"
	"github.com/taylorskalyo/goreader/nav"
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

	a := app.NewApp(rc, new(nav.Pager))
	a.Run()

	if a.Err() != nil {
		fmt.Fprintf(os.Stderr, "Exit with error: %s\n", a.Err().Error())
		os.Exit(1)
	}

	os.Exit(0)
}

func printUsage() {
	fmt.Fprintln(os.Stderr, "goreader [epub_file]")
	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintln(os.Stderr, "-h		print keybindings")
}

func printHelp() {
	fmt.Fprintln(os.Stderr, "Key                  Action")
	fmt.Fprintln(os.Stderr)
	fmt.Fprintln(os.Stderr, "q / Esc              Quit")
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
