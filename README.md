# goreader

Terminal epub reader

[![Build Status](https://travis-ci.org/taylorskalyo/goreader.svg?branch=master)](https://travis-ci.org/taylorskalyo/goreader)
[![Go Report Card](https://goreportcard.com/badge/github.com/taylorskalyo/goreader)](https://goreportcard.com/report/github.com/taylorskalyo/goreader)

Goreader is a minimal ereader application that runs in the terminal. Images are displayed as ASCII art. Commands are based on less.

## Installation

``` shell
go get github.com/taylorskalyo/goreader
```

## Usage

``` shell
goreader [epub_file]
```

### Keybindings

| Key | Action            |
| --- | ----------------- |
| `q` | Quit              |
| `j` | Scroll up         |
| `k` | Scroll down       |
| `h` | Scroll left       |
| `l` | Scroll right      |
| `H` | Previous chapter  |
| `L` | Next chapter      |
| `g` | Top of chapter    |
| `G` | Bottom of chapter |
