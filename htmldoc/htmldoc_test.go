package htmldoc

import (
	"bytes"
	"testing"
)

const expFormat = "Expected: %v, but got: %v\n"

func TestRender(t *testing.T) {
	rawHTML := `<html><body>Hello, world!</body></html>`
	buf := bytes.NewBuffer([]byte{})
	doc := Document{
		Reader: bytes.NewBufferString(rawHTML),
		Writer: buf,
	}
	if err := doc.Render(); err != nil {
		t.Errorf("Unexepected error: %s", err.Error())
	}
	exp := "Hello, world!"
	if buf.String() != exp {
		t.Errorf(expFormat, exp, buf.String())
	}
}
