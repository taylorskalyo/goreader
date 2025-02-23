package render

import (
	"io"
	"strings"

	"github.com/rivo/tview"
)

type wordWrapWriter struct {
	w      io.Writer
	width  int
	buffer strings.Builder
}

func newWordWrapWriter(w io.Writer, width int) *wordWrapWriter {
	return &wordWrapWriter{
		w:     w,
		width: width,
	}
}

func (www *wordWrapWriter) Write(p []byte) (n int, err error) {
	www.buffer.Write(p)
	lines := tview.WordWrap(www.buffer.String(), www.width)

	for i, line := range lines {
		if i == len(lines)-1 {
			// Keep the last line in the buffer
			www.buffer.Reset()
			www.buffer.WriteString(line)
			break
		}

		nLine, err := www.w.Write([]byte(line + "\n"))
		if err != nil {
			return n, err
		}
		n += nLine
	}

	return len(p), nil
}

func (www *wordWrapWriter) Flush() error {
	if www.buffer.Len() > 0 {
		_, err := www.w.Write([]byte(www.buffer.String()))
		www.buffer.Reset()

		return err
	}

	return nil
}
