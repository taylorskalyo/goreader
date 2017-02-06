# epub

Minimal epub library written in Go

[![GoDoc](https://img.shields.io/badge/godoc-reference-blue.svg)](https://godoc.org/github.com/taylorskalyo/goreader/epub)

## Installation

``` shell
go get github.com/taylorskalyo/goreader/epub
```

## Basic usage

``` golang
import "github.com/taylorskalyo/goreader/epub"

rc, err := epub.OpenReader(os.Args[1])
if err != nil {
	panic(err)
}
defer rc.Close()

// The rootfile (content.opf) lists all of the contents of an epub file.
// There may be multiple rootfiles, although typically there is only one.
book := rc.Rootfiles[0]

// Print book title.
fmt.Println(book.Title)

// List the IDs of files in the book's spine.
for _, item := range book.Spine.Itemrefs {
	fmt.Println(item.ID)
}
```
