package views

import (
	"strings"
	"testing"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/stretchr/testify/assert"
	"github.com/taylorskalyo/goreader/config"
	"github.com/taylorskalyo/goreader/epub"
	"github.com/taylorskalyo/goreader/state"
	"golang.org/x/sync/errgroup"
)

type testScreen struct {
	tcell.SimulationScreen
}

func (ts testScreen) String() string {
	var b strings.Builder
	cells, w, _ := ts.GetContents()
	for i, cell := range cells {
		if len(cell.Runes) > 0 {
			b.WriteString(string(cell.Bytes))
		}

		if i%w == w-1 {
			b.WriteString("\n")
		}
	}

	return b.String()
}

func newTestScreen(t *testing.T) testScreen {
	ts := testScreen{}
	ts.SimulationScreen = tcell.NewSimulationScreen("UTF-8")
	if err := ts.Init(); err != nil {
		t.Fatal(err)
	}

	return ts
}

func newTestApp(t *testing.T) *Application {
	t.Setenv("XDG_STATE_HOME", t.TempDir())
	state.ReloadEnv()

	return NewApplication()
}

func TestNavigation(t *testing.T) {
	eg := new(errgroup.Group)

	ts := newTestScreen(t)
	app := newTestApp(t)
	app.SetScreen(ts)

	rc, _ := epub.OpenReader("../epub/_test_files/alice.epub")
	defer rc.Close()

	eg.Go(app.Run)

	app.QueueUpdateDraw(func() {
		ts.SetSize(80, 20)
		app.OpenBook(rc.DefaultRendition())
	})

	// Given a keypress, verify a search pattern appears on the screen.
	for _, tc := range []struct {
		event  *tcell.EventKey
		search string
	}{
		// Forward / Backward
		{nil, "(?s)1 OF 4.*Cover"},
		{tcell.NewEventKey(tcell.KeyRune, 'f', tcell.ModNone), "2 OF 4"},
		{tcell.NewEventKey(tcell.KeyRune, 'f', tcell.ModNone), "3 OF 4"},
		{tcell.NewEventKey(tcell.KeyRune, 'f', tcell.ModNone), "(?s)4 OF 4.*Alt text: Cover"},
		{tcell.NewEventKey(tcell.KeyRune, 'f', tcell.ModNone), "(?s)1 OF 23.*Project Gutenberg's Alice's Adventures in Wonderland, by Lewis Carroll"},
		{tcell.NewEventKey(tcell.KeyRune, 'b', tcell.ModNone), "(?s)4 OF 4.*Alt text: Cover"},
		{tcell.NewEventKey(tcell.KeyRune, 'b', tcell.ModNone), "3 OF 4"},
		{tcell.NewEventKey(tcell.KeyRune, 'b', tcell.ModNone), "2 OF 4"},
		{tcell.NewEventKey(tcell.KeyRune, 'b', tcell.ModNone), "1 OF 4"},
		{tcell.NewEventKey(tcell.KeyRune, 'b', tcell.ModNone), "1 OF 4"},

		// Top / Bottom
		{tcell.NewEventKey(tcell.KeyRune, 'G', tcell.ModNone), "(?s)4 OF 4.*Alt text: Cover"},
		{tcell.NewEventKey(tcell.KeyRune, 'g', tcell.ModNone), "(?s)1 OF 4.*Cover"},

		// ChapterPrevious / ChapterNext
		{tcell.NewEventKey(tcell.KeyRune, 'H', tcell.ModNone), "(?s)1 OF 4.*Cover"},
		{tcell.NewEventKey(tcell.KeyRune, 'L', tcell.ModNone), "(?s)1 OF 23.*Project Gutenberg's Alice's Adventures in Wonderland, by Lewis Carroll"},
		{tcell.NewEventKey(tcell.KeyRune, 'H', tcell.ModNone), "(?s)1 OF 4.*Cover"},
		{tcell.NewEventKey(tcell.KeyRune, 'L', tcell.ModNone), "(?s)1 OF 23.*Project Gutenberg's Alice's Adventures in Wonderland, by Lewis Carroll"},
		{tcell.NewEventKey(tcell.KeyRune, 'L', tcell.ModNone), "(?s)1 OF 17.*CHAPTER I"},
		{tcell.NewEventKey(tcell.KeyRune, 'L', tcell.ModNone), "(?s)1 OF 22.*CHAPTER II"},

		{tcell.NewEventKey(tcell.KeyRune, 'G', tcell.ModNone), "-{80}"},

		// Up / Down
		{tcell.NewEventKey(tcell.KeyRune, 'j', tcell.ModNone), "-{80}"},
		{tcell.NewEventKey(tcell.KeyRune, 'k', tcell.ModNone), "away from her"},
		{tcell.NewEventKey(tcell.KeyRune, 'k', tcell.ModNone), "sorrowful tone"},
		{tcell.NewEventKey(tcell.KeyRune, 'k', tcell.ModNone), "hundred pounds!"},
		{tcell.NewEventKey(tcell.KeyRune, 'j', tcell.ModNone), "whole party swam"},
	} {
		if tc.event != nil {
			t.Logf("Simulating keypress: %s\n", config.KeyChordFromEvent(*tc.event))
			ts.InjectKey(tc.event.Key(), tc.event.Rune(), tc.event.Modifiers())

			// Wait for app to process the queued event and force it to re-draw the
			// screen.
			//
			// TODO: Is there a better way to do this?
			time.Sleep(50 * time.Millisecond)
			app.QueueUpdateDraw(func() {})
		}

		app.QueueUpdate(func() {
			t.Logf("Simulated screen state:\n%s", ts.String())
			assert.Regexp(t, tc.search, ts.String())
		})
	}

	ts.InjectKey(tcell.KeyRune, 'q', tcell.ModNone)
	assert.NoError(t, eg.Wait())
}
