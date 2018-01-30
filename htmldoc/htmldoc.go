package htmldoc

import (
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

// DisplayType describes the orientation and visibility of HTML elements.
type DisplayType uint8

// BoxType describes the orientation and visibility of boxes in the layout
// tree.
type BoxType uint8

// Value describes a CSS value.
type Value interface {
	Px() float32
}

// Keyword is a word with a special meaning.
type Keyword struct {
	string
}

// Px returns Keyword k represented as pixels.
func (k Keyword) Px() float32 {
	return 0.0
}

// Length describes a quantity and unit.
type Length struct {
	float32
	Unit Unit
}

// Px returns Length l represented as pixels.
func (l Length) Px() float32 {
	return l.float32
}

// Unit is a standard measure of a quantity
type Unit uint8

const (
	// NoneDisplay elements are not displayed.
	NoneDisplay DisplayType = iota
	// BlockDisplay elements are placed vertically within their container, from
	// top to bottom.
	BlockDisplay
	// InlineDisplay elements are placed horizontally within their container,
	// from left to right. If they reach the right edge of the container, they
	// will wrap around and continue on a new line below.
	InlineDisplay

	// AnonBlockBox boxes are containers used to separate block and inline boxes.
	//
	// https://www.w3.org/TR/CSS22/visuren.html#anonymous-block-level
	AnonBlockBox BoxType = iota
	// AnonInlineBox boxes are used to contain text that is not directly
	// contained inside a block container element.
	//
	// https://www.w3.org/TR/CSS22/visuren.html#anonymous
	AnonInlineBox
	// BlockBox boxes are placed vertically within their container, from top to
	// bottom.
	//
	// https://www.w3.org/TR/CSS22/visuren.html#block-boxes
	BlockBox
	// InlineBox boxes are placed horizontally within their container, from left
	// to right. If they reach the right edge of the container, they will wrap
	// around and continue on a new line below.
	//
	// https://www.w3.org/TR/CSS22/visuren.html#inline-boxes
	InlineBox

	// Px measures length in pixels.
	Px Unit = iota
)

// A Rect is a quadrilateral with four right angles.
type Rect struct {
	X, Y, Width, Height float32
}

// ExpandedBy returns a copy of Rect r, expanded by EdgeSizes e.
func (r Rect) ExpandedBy(e EdgeSizes) Rect {
	return Rect{
		X:      r.X - e.Left,
		Y:      r.Y - e.Top,
		Width:  r.Width + e.Left + e.Right,
		Height: r.Height + e.Top + e.Bottom,
	}
}

// Dimensions represents the dimensions of a layout box.
type Dimensions struct {
	Rect
	Padding, Border, Margin EdgeSizes
}

// PaddingBox represents the area of Dimensions d, plus padding.
func (d Dimensions) PaddingBox() Rect {
	return d.Rect.ExpandedBy(d.Padding)
}

// BorderBox represents the area of Dimensions d, plus padding, and borders.
func (d Dimensions) BorderBox() Rect {
	return d.PaddingBox().ExpandedBy(d.Border)
}

// MarginBox represents the area of Dimensions d, plus padding, borders, and
// margin.
func (d Dimensions) MarginBox() Rect {
	return d.BorderBox().ExpandedBy(d.Margin)
}

// EdgeSizes represents sizes of a particular property for each edge of a
// layout box.
type EdgeSizes struct {
	Left, Right, Top, Bottom float32
}

// A LayoutBox is a node in the layout tree.
type LayoutBox struct {
	Parent, FirstChild, LastChild, NextSibling, PrevSibling *LayoutBox

	BoxType    BoxType
	Dimensions Dimensions
}

// NewBox creates a new LayoutBox of type bt.
func NewBox(bt BoxType) *LayoutBox {
	return &LayoutBox{
		BoxType: bt,
	}
}

// appendChild adds a LayoutBox node c as a child of box. It will panic if c
// already has a parent or siblings.
func (box *LayoutBox) appendChild(c *LayoutBox) {
	if c.Parent != nil || c.PrevSibling != nil || c.NextSibling != nil {
		panic("htmldoc: AppendChild called for an attached child LayoutBox")
	}
	last := box.LastChild
	if last != nil {
		last.NextSibling = c
	} else {
		box.FirstChild = c
	}
	box.LastChild = c
	c.Parent = box
	c.PrevSibling = last
}

func displayType(n *html.Node) DisplayType {
	switch n.DataAtom {
	case atom.A, atom.Abbr, atom.B, atom.Bdo, atom.Big, atom.Br, atom.Button,
		atom.Cite, atom.Code, atom.Dfn, atom.Em, atom.I, atom.Img, atom.Input,
		atom.Kbd, atom.Label, atom.Map, atom.Object, atom.Q, atom.Samp,
		atom.Script, atom.Select, atom.Small, atom.Span, atom.Strong, atom.Sub,
		atom.Sup, atom.Textarea, atom.Td, atom.Time, atom.Tt, atom.Var:
		return InlineDisplay
	default:
		return BlockDisplay
	}
}

// LayoutTree transforms an HTML Node tree into a LayoutBox tree.
func LayoutTree(n *html.Node, d Dimensions) *LayoutBox {
	d.Height = 0.0
	root := buildLayoutTree(n)
	root.layout(d)
	return root
}

// buildLayoutTree builds a tree of LayoutBoxes. However, it does not perform
// any layout calculations.
func buildLayoutTree(n *html.Node) *LayoutBox {
	if n == nil {
		return nil
	}

	// Create the root LayoutBox node.
	var root *LayoutBox
	if n.Type == html.DocumentNode {
		root = NewBox(BlockBox)
	} else if n.DataAtom == 0 {
		root = NewBox(AnonInlineBox)
	} else {
		switch displayType(n) {
		case BlockDisplay:
			root = NewBox(BlockBox)
		case InlineDisplay:
			root = NewBox(InlineBox)
		default:
			panic("htmldoc: Root node has display: none")
		}
	}
	// Create the children.
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		child := buildLayoutTree(c)
		switch displayType(c) {
		case BlockDisplay:
			root.appendChild(child)
		case InlineDisplay:
			root.inlineContainer().appendChild(child)
		}
	}

	return root
}

// inlineContainer returns the container that should be used for new inline
// children.
func (box *LayoutBox) inlineContainer() *LayoutBox {
	switch box.BoxType {
	case BlockBox, AnonBlockBox:
		if box.LastChild == nil {
			box.appendChild(NewBox(AnonBlockBox))
		}
		return box.LastChild
	}

	return box
}

// layout positions a box.
func (box *LayoutBox) layout(d Dimensions) {
	switch box.BoxType {
	case BlockBox:
		box.layoutBlock(d)
	default:
		// TODO
	}
}

func (box *LayoutBox) layoutBlock(d Dimensions) {
	// Child width can depend on parent width, so calculate this box's width
	// before laying out children.
	box.setBlockWidth(d)

	// Determine where this box is located within its container.
	box.setBlockPosition(d)

	// Recursively lay out this box's children.
	box.layoutBlockChildren()
}

// setBlockWidth calculates the width of a block element.
//
// http://www.w3.org/TR/CSS2/visudet.html#blockwidth
func (box *LayoutBox) setBlockWidth(d Dimensions) {
	// Width defaults to auto
	var width Value
	auto := Keyword{"auto"}
	width = auto

	// Margin, border, and padding default to zero.
	zero := Length{}
	var marginLeft, marginRight Value
	marginLeft = zero
	marginRight = zero

	var borderLeft, borderRight Value
	borderLeft = zero
	borderRight = zero

	var paddingLeft, paddingRight Value
	paddingLeft = zero
	paddingRight = zero

	var total float32
	for _, l := range []float32{
		marginLeft.Px(),
		marginRight.Px(),
		borderLeft.Px(),
		borderRight.Px(),
		paddingLeft.Px(),
		paddingRight.Px(),
	} {
		total += l
	}

	// If width is not auto and the total is wider than the container width,
	// treat margins as 0.
	if v, ok := width.(Keyword); ok && v != auto && total > d.Width {
		if marginLeft != auto {
			marginLeft = zero
		}
		if marginRight != auto {
			marginRight = zero
		}
	}

	// Shrink dimensions to fit the container.
	underflow := d.Width - total

	if width == auto {
		if marginLeft == auto {
			marginLeft = zero
		}
		if marginRight == auto {
			marginRight = zero
		}
		if underflow >= 0.0 {
			// Width is the remaining space.
			width = Length{underflow, Px}
		} else {
			// Adjust the right margin instead so width is not negative.
			width = zero
			marginRight = Length{marginRight.Px() + underflow, Px}
		}
	} else {
		if marginLeft != auto && marginRight != auto {
			// If the values are overconstrained, calculate marginRight.
			marginRight = Length{marginRight.Px() + underflow, Px}
		} else if marginLeft != auto && marginRight == auto {
			// If only marginRight is auto, its value is the remaining space.
			marginRight = Length{underflow, Px}
		} else if marginLeft == auto && marginRight != auto {
			// If only marginLeft is auto, its value is the remaining space.
			marginLeft = Length{underflow, Px}
		} else if marginLeft == auto && marginRight == auto {
			// If marginLeft and marginRight are both auto, they both take the
			// remaining space.
			marginLeft = Length{underflow / 2.0, Px}
			marginRight = Length{underflow / 2.0, Px}
		}
	}

	box.Dimensions.Width = width.Px()

	box.Dimensions.Margin.Left = marginLeft.Px()
	box.Dimensions.Margin.Right = marginRight.Px()

	box.Dimensions.Border.Left = borderLeft.Px()
	box.Dimensions.Border.Right = borderRight.Px()

	box.Dimensions.Padding.Left = paddingLeft.Px()
	box.Dimensions.Padding.Right = paddingRight.Px()
}

// setBoxPosition positions an element within its container.
//
// http://www.w3.org/TR/CSS2/visudet.html#normal-block
func (box *LayoutBox) setBlockPosition(d Dimensions) {
	zero := Length{}

	// Margin, border, and padding default to zero.
	box.Dimensions.Margin.Top = zero.Px()
	box.Dimensions.Margin.Bottom = zero.Px()

	box.Dimensions.Border.Top = zero.Px()
	box.Dimensions.Border.Bottom = zero.Px()

	box.Dimensions.Padding.Top = zero.Px()
	box.Dimensions.Padding.Bottom = zero.Px()

	box.Dimensions.X = d.X
	box.Dimensions.X += box.Dimensions.Margin.Left
	box.Dimensions.X += box.Dimensions.Border.Left
	box.Dimensions.X += box.Dimensions.Padding.Left

	// Position the box below all the previous boxes in the container.
	box.Dimensions.Y = d.Height + d.Y
	box.Dimensions.Y += box.Dimensions.Margin.Top
	box.Dimensions.Y += box.Dimensions.Border.Top
	box.Dimensions.Y += box.Dimensions.Padding.Top
}

// layoutBlockChildren positions child boxes whithin their container.
func (box *LayoutBox) layoutBlockChildren() {
	for c := box.FirstChild; c != nil; c = c.NextSibling {
		c.layout(box.Dimensions)
		box.Dimensions.Height += c.Dimensions.MarginBox().Height
	}
}
