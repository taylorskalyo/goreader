package main

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/taylorskalyo/goreader/app"
	"github.com/taylorskalyo/goreader/epub"
	"github.com/taylorskalyo/goreader/nav"

	"github.com/adrg/xdg"
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

	title := filepath.Base(os.Args[1])

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

	chapter := getPage(title)

	a := app.NewApp(book, new(nav.Pager), chapter)

	chapter = a.Run()

	if a.Err() != nil {
		fmt.Fprintf(os.Stderr, "Exit with error: %s\n", a.Err().Error())
		os.Exit(1)
	}

	if err := savePage(chapter, title); err != nil {
		fmt.Fprintf(os.Stderr, "Error saving page: %s\n", err)
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

// savePage saves the name of the current epub (as a key) and the current chapter number (as an value)
// to a .json file in $XDG_STATE_HOME, either creating or updating it.
// It returns any error that occurs during the process.
func savePage(chapter int, title string) error {
	stateDir := xdg.StateHome
	appStateDir := filepath.Join(stateDir, "goreader")
	stateFile := filepath.Join(appStateDir, "last_read_pages.json")

	if err := os.MkdirAll(appStateDir, 0700); err != nil {
		return err
	}

	books := make(map[string]int)
	data, err := os.ReadFile(stateFile)

	if err != nil {
		if !os.IsNotExist(err) {
			return err
		}
	} else {
		if err = json.Unmarshal(data, &books); err != nil {
			return err
		}
	}

	books[title] = chapter

	data, err = json.MarshalIndent(books, "", " ")
	if err != nil {
		return err
	}

	if err := os.WriteFile(stateFile, data, 0644); err != nil {
		return err
	}

	return nil
}

// getPage opens the statefile named 'last_read_pages.json' in $XDG_STATE_HOME and looks for the
// number value for the title key. If not present, or if an error occurs, it simply returns 0.
func getPage(title string) int {
	stateDir := xdg.StateHome
	appStateDir := filepath.Join(stateDir, "goreader")
	stateFile := filepath.Join(appStateDir, "last_read_pages.json")

	books := make(map[string]int)

	data, err := os.ReadFile(stateFile)
	if err != nil {
		return 0
	}

	err = json.Unmarshal(data, &books)
	if err != nil {
		return 0
	}

	if chapter, exists := books[title]; exists {
		return chapter
	}

	return 0
}
