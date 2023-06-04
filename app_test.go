package main

import (
	"testing"

	termbox "github.com/nsf/termbox-go"
	"github.com/stretchr/testify/assert"
)

func TestInitNavigationKeys(t *testing.T) {
	a := new(app)

	keymap, chmap := a.initNavigationKeys()

	assert.Contains(t, keymap, termbox.KeyArrowDown)
	assert.Contains(t, keymap, termbox.KeyArrowUp)
	assert.Contains(t, keymap, termbox.KeyArrowRight)
	assert.Contains(t, keymap, termbox.KeyArrowLeft)
	assert.Contains(t, keymap, termbox.KeyEsc)

	assert.Contains(t, chmap, 'j')
	assert.Contains(t, chmap, 'k')
	assert.Contains(t, chmap, 'h')
	assert.Contains(t, chmap, 'l')
	assert.Contains(t, chmap, 'g')
	assert.Contains(t, chmap, 'G')
	assert.Contains(t, chmap, 'q')
	assert.Contains(t, chmap, 'f')
	assert.Contains(t, chmap, 'b')
	assert.Contains(t, chmap, 'L')
	assert.Contains(t, chmap, 'H')
}
