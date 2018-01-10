package htmldoc

import (
	"bytes"
	"testing"
)

const expFormat = "Expected:\n'%v'\nBut got:\n'%v'\n"

func TestRender(t *testing.T) {
	testCases := []struct {
		name, src, exp string
	}{
		{
			"basic html",
			`<html><body>Hello, world!</body></html>`,
			"Hello, world!",
		},
		{
			"pretty printed html",
			`
			<html>
				<body>
					Hello, world!
				</body>
			</html>`,
			"Hello, world!",
		},
		{
			"horizontal rule",
			`<html><hr></html>`,
			"-----",
		},
		{
			"break",
			`<html>foo<br>bar</html>`,
			"foo\nbar",
		},
		{
			"paragraph",
			`<html><p>foo</p><p>bar</p></html>`,
			"foo\n\nbar",
		},
		{
			"nested paragraphs",
			`<html><p>foo<p>bar</p></p></html>`,
			"foo\n\nbar",
		},
		{
			"headings",
			`<html>
			<h1>One</h1>
			<h2>Two</h2>
			<h3>Three</h3>
			<h4>Four</h4>
			<h5>Five</h5>
			<h6>Six</h6>
			</html>`,
			"One\n\nTwo\n\nThree\n\nFour\n\nFive\n\nSix",
		},
		{
			"div",
			`<html>
			<div>foo</div>
			<div>
				bar
				<div>
				  baz
				</div>
			</div>
			</html>`,
			"foo\nbar\nbaz",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			buf := bytes.NewBuffer([]byte{})
			doc := Document{
				Reader: bytes.NewBufferString(tc.src),
				Writer: buf,
			}
			if err := doc.Render(); err != nil {
				t.Error(err.Error())
			}
			if buf.String() != tc.exp {
				t.Errorf(expFormat, tc.exp, buf.String())
			}
		})
	}
}

func TestRenderWithWidth(t *testing.T) {
	testCases := []struct {
		name, src, exp string
	}{
		{
			"horizontal rule",
			`<html><hr></html>`,
			"----------------------------------------",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			buf := bytes.NewBuffer([]byte{})
			doc := Document{
				Reader: bytes.NewBufferString(tc.src),
				Writer: buf,
				Width:  40,
			}
			if err := doc.Render(); err != nil {
				t.Error(err.Error())
			}
			if buf.String() != tc.exp {
				t.Errorf(expFormat, tc.exp, buf.String())
			}
		})
	}
}
