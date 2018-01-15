package htmldoc

import (
	"bufio"
	"io"
	"regexp"
	"strings"
	"unicode"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

const (
	inline Display = iota
	block

	atomText atom.Atom = 0

	maxUint = ^uint(0)
	maxInt  = int(maxUint >> 1)
)

type Display int

type Node struct {
	Parent, FirstChild, LastChild, PrevSibling, NextSibling *Node

	// X and Y describe the location of the current node in a rendered document.
	X, Y int

	// HeadOffset and TailOffset describe the position of inline text inside a
	// a node's bounding box.
	HeadOffset, TailOffset int

	// RenderWidth and RenderHeight describe the width and height of the node's
	// bounding box within a rendered document.
	RenderWidth, RenderHeight int

	contentWidth, minWidth int

	colWidths []int

	html.Node
}

func (n Node) Display() Display {
	switch n.DataAtom {
	case atomText, atom.A, atom.Abbr, atom.B, atom.Bdo, atom.Big, atom.Br,
		atom.Button, atom.Cite, atom.Code, atom.Dfn, atom.Em, atom.I, atom.Img,
		atom.Input, atom.Kbd, atom.Label, atom.Map, atom.Object, atom.Q, atom.Samp,
		atom.Script, atom.Select, atom.Small, atom.Span, atom.Strong, atom.Sub,
		atom.Sup, atom.Textarea, atom.Time, atom.Tt, atom.Var:
		return inline
	}

	return block
}

func (n Node) setMinWidth() {
	switch n.DataAtom {
	case atomText:
		// For text elements, use the length of the longest word.
		scanner := bufio.NewScanner(strings.NewReader(n.Data))
		scanner.Split(bufio.ScanWords)
		for scanner.Scan() {
			if width := len(scanner.Text()); width > n.minWidth {
				n.minWidth = width
			}
		}
	case atom.Img:
		// For images, use the length of the alt text or a fixed minWidth.
		max := 5
		for _, a := range n.Attr {
			if atom.Lookup([]byte(a.Key)) == atom.Alt {
				if width := len(a.Val); width > max {
					max = width
				}
			}
		}
		n.minWidth = max
	case atom.Hr:
		// For hr elements, use a fixed minWidth.
		n.minWidth = 5
	case atom.Tr:
		// For table rows, use the sum of each cell minWidth.
		for c := n.FirstChild; c != nil; c = n.NextSibling {
			n.minWidth += c.minWidth
		}
	default:
		// For everything else, use the max minWidth of all child elements.
		for c := n.FirstChild; c != nil; c = n.NextSibling {
			if c.minWidth > n.minWidth {
				n.minWidth = c.minWidth
			}
		}
	}
}

func (n Node) calcColWidths(tWidth int) (widths []int) {
	if n.DataAtom != atom.Table {
		return widths
	}

	// Initialize widths with each column's minWidth. Sum column content widths
	// in cWidths.
	var cWidths []int
	var col, row int
	for r := n.FirstChild; r != nil; r = r.NextSibling {
		col = 0
		for c := r.FirstChild; c != nil; c = c.NextSibling {
			if len(widths) <= col {
				widths = append(widths, 0)
			}
			if len(cWidths) <= col {
				cWidths = append(cWidths, 0)
			}
			if c.minWidth > widths[col] {
				widths[col] = c.minWidth
			}
			cWidths[col] += c.contentWidth
			col++
		}
		row++
	}

	// Adjust tWidth to account for borders.
	tWidth -= col + 1

	// Calculate average content width for each column. Do not exceed tWidth.
	for col = 0; col < len(cWidths); col++ {
		cWidths[col] /= row
		if cWidths[col] > tWidth {
			cWidths[col] = tWidth
		}
	}

	// Shrink columns until they fit tWidth.
	sum := 0
	for _, w := range widths {
		sum += w
	}
	for sum > tWidth {
		// Shrink from right to left.

		for col = len(widths) - 1; widths[col] <= 0 && col > 0; col++ {
		}
		widths[col]--
		sum--
	}

	// Grow columns until they fit tWidth.
	sum = 0
	for _, w := range widths {
		sum += w
	}
	for sum < tWidth {
		// Pick the column with the highest cWidth to width ratio to grow.
		for i := range widths {
			if cWidths[i]/(widths[i]+1) > cWidths[col]/(widths[col]+1) {
				col = i
			}
		}
		widths[col]++
		sum++
	}

	return widths
}

func (n Node) setContentWidth() {
	switch n.DataAtom {
	case atomText:
		// For text elements, use the length of the element's data.
		n.contentWidth = len(n.Data)
	case atom.Img, atom.Hr, atom.Table:
		// For image and hr elements, take up all available space.
		n.contentWidth = maxInt
	case atom.Tr:
		// For table rows, use the sum of each cell contentWidth.
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			n.contentWidth += c.contentWidth
		}
	default:
		// For everything else, use the max contentWidth of all child elements.
		// Group inline elements together when possible.
		inlineWidth := 0
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			switch c.Display() {
			case inline:
				inlineWidth += c.contentWidth
				if inlineWidth > n.contentWidth {
					n.contentWidth = inlineWidth
				}
			default:
				inlineWidth = 0
				if c.contentWidth > n.contentWidth {
					n.contentWidth = c.contentWidth
				}
			}
		}
	}
}

// trimSpace removes extra formatting spaces in pretty printed HTML.
func trimSpace(prev, cur *Node) *Node {
	if cur.Display() == block && prev != nil {
		// If we encounter a block element, trim spaces from the right of the last
		// text element.
		prev.Data = strings.TrimRightFunc(prev.Data, unicode.IsSpace)
		prev = nil
	} else if cur.Type == html.TextNode && len(cur.Data) > 0 {
		// If we encounter a text element, ensure at most one single space exists
		// between this text and the previous text.
		var appendSpace bool
		if prev != nil {
			tnData := []rune(prev.Data)
			appendSpace = unicode.IsSpace(tnData[len(tnData)-1])
			prev.Data = strings.TrimRightFunc(prev.Data, unicode.IsSpace)
			if appendSpace {
				prev.Data += " "
			}
		} else {
			appendSpace = true
		}
		prependSpace := !appendSpace && unicode.IsSpace([]rune(cur.Data)[0])
		cur.Data = strings.TrimLeftFunc(cur.Data, unicode.IsSpace)
		if len(cur.Data) > 0 {
			if prependSpace {
				cur.Data = " " + cur.Data
			}
			prev = cur
		}
	}

	// Remove newlines; replace multiple spaces with a single space.
	re := regexp.MustCompile(`\s+`)
	cur.Data = re.ReplaceAllString(cur.Data, " ")

	return prev
}

func Parse(r io.Reader) (*Node, error) {
	hn, err := html.Parse(r)
	if err != nil {
		return nil, err
	}

	var inlineText *Node

	// Recursively convert *html.Node into *Node.
	c := &Node{
		Node:   *hn,
		Parent: nil,
	}
	for {
		if c.FirstChild == nil && c.Node.FirstChild != nil {
			// Create and visit child node.
			c.FirstChild = &Node{Node: *c.Node.FirstChild, Parent: c}
			c = c.FirstChild
			inlineText = trimSpace(inlineText, c)
		} else if c.NextSibling == nil && c.Node.NextSibling != nil {
			// Create and visit sibling node.
			c.NextSibling = &Node{Node: *c.Node.NextSibling, Parent: c}
			c.NextSibling.PrevSibling = c
			c = c.NextSibling
			inlineText = trimSpace(inlineText, c)
		} else if c.Parent != nil {
			// Return to parent.
			c.Parent.LastChild = c
			c = c.Parent
			c.setMinWidth()
			c.setContentWidth()
		} else {
			// All nodes have been visited.
			break
		}
	}

	return c, nil
}

func (n *Node) Render(w int) {
	// Traverse tree
	// - Set width/height and x/y fields
	c := n
	c.RenderHeight = -1
	for {
		if c == n && c.RenderHeight != -1 {
			break
		}
	}
}

func (n Node) String() string {
	// Not necessary, but could be useful to other users and for testing

	// Create cell array
	// Traverse tree, filling cell array
	// Iterate over cell array, writing each cell's rune to string buffer
	// Store current attribute/ANSI sequence in a var
	// When attribute/ANSI sequence changes, write the new value to the string buffer
	return ""
}
