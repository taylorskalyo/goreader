package render

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"path"
	"sort"

	"github.com/nfnt/resize"
	"github.com/taylorskalyo/goreader/epub"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

// handleImage appends image elements to the parser buffer. It extracts alt
// text and renders images.
func (r *Renderer) handleImage(token html.Token) error {
	// Sort attributes so that image source is handled before image alt text.
	sort.Slice(token.Attr, func(i, j int) bool {
		return token.Attr[i].Key > token.Attr[j].Key
	})

	for _, a := range token.Attr {
		switch atom.Lookup([]byte(a.Key)) {
		case atom.Src:
			if err := r.handleImageSrc(a.Val); err != nil {
				return err
			}
		case atom.Alt:
			text := fmt.Sprintf("Alt text: %s", a.Val)
			r.parser.ensureNewlines(1)
			if err := r.appendText(text); err != nil {
				return err
			}
			r.parser.ensureNewlines(1)
		}
	}

	return nil
}

// handleImageSrc reads a referenced image and renders it to the parser buffer.
func (r *Renderer) handleImageSrc(href string) error {
	if r.parser.writeTarget() != r.parser.writer {
		// NOTE: rendering images inside tables is not supported at the moment as
		// this would add a lot of complexity.
		return nil
	}

	if !path.IsAbs(href) {
		href = path.Join(r.parser.basepath, href)
	}

	for _, item := range r.content.Items {
		if item.HREF == href {
			for _, line := range imageToText(item, r.width) {
				r.parser.ensureNewlines(1)
				if err := r.appendText(line); err != nil {
					return err
				}
				r.parser.ensureNewlines(1)
			}
			break
		}
	}

	return nil
}

// imageToText renders an image as lines of text.
func imageToText(item epub.Item, width int) []string {
	lines := []string{}
	r, err := item.Open()
	if err != nil {
		return lines
	}

	img, _, err := image.Decode(r)
	if err != nil {
		return lines
	}
	bounds := img.Bounds()

	// Assume a character height to width ratio of 2:1.
	h := (bounds.Max.Y * width) / (bounds.Max.X * 2)
	img = resize.Resize(uint(width), uint(h), img, resize.Lanczos3)

	charGradient := []rune("MND8OZ$7I?+=~:,..")
	buf := new(bytes.Buffer)

	for y := 0; y < h; y++ {
		for x := 0; x < width; x++ {
			c := color.GrayModel.Convert(img.At(x, y))
			y := c.(color.Gray).Y
			pos := (len(charGradient) - 1) * int(y) / 255
			buf.WriteRune(charGradient[pos])
		}
		lines = append(lines, buf.String())
		buf = new(bytes.Buffer)
	}

	return lines
}
