package render

import (
	"io"
	"strings"

	"github.com/rivo/tview"
)

// wordWrapWriter wraps an io.Writer. As text is written, lines are wrapped at
// a given width before being passed to the underlying Writer.
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

func (w *wordWrapWriter) Write(p []byte) (n int, err error) {
	w.buffer.Write(p)
	lines := tview.WordWrap(w.buffer.String(), w.width)

	for i, line := range lines {
		if i == len(lines)-1 {
			// Keep the last line in the buffer
			w.buffer.Reset()
			w.buffer.WriteString(line)
			break
		}

		nLine, err := w.w.Write([]byte(line + "\n"))
		if err != nil {
			return n, err
		}
		n += nLine
	}

	return len(p), nil
}

func (w *wordWrapWriter) Flush() error {
	if w.buffer.Len() > 0 {
		_, err := w.w.Write([]byte(w.buffer.String()))
		w.buffer.Reset()

		return err
	}

	return nil
}

// implements table.WidthEnforcer
func tviewWidthEnforcer(text string, width int) string {
	lines := tview.WordWrap(text, width)

	return strings.Join(lines, "\n")
}
