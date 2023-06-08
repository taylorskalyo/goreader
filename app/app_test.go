package app

import (
	"testing"

	termbox "github.com/nsf/termbox-go"
	localMock "github.com/taylorskalyo/goreader/mock"
)

func TestInitNavigationKeys(t *testing.T) {
	p := &localMock.MockPageNavigator{}
	a := localMock.NewMockApplication(p)

	a.On("PageNavigator").Return(p)
	a.On("Exit").Return()
	a.On("Forward").Return()
	a.On("Back").Return()
	a.On("NextChapter").Return()
	a.On("PrevChapter").Return()

	p.On("ScrollDown").Return()
	p.On("ScrollUp").Return()
	p.On("ScrollLeft").Return()
	p.On("ScrollRight").Return()
	p.On("ToTop").Return()
	p.On("ToBottom").Return()

	keymap, chmap := initNavigationKeys(a)

	keymap[termbox.KeyArrowDown]()
	keymap[termbox.KeyArrowUp]()
	keymap[termbox.KeyArrowRight]()
	keymap[termbox.KeyArrowLeft]()
	keymap[termbox.KeyEsc]()

	chmap['j']()
	chmap['k']()
	chmap['h']()
	chmap['l']()
	chmap['g']()
	chmap['G']()

	chmap['q']()
	chmap['f']()
	chmap['b']()
	chmap['L']()
	chmap['H']()

	p.AssertExpectations(t)
}
