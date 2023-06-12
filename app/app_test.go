package app

import (
	"testing"
	"unicode"

	termbox "github.com/nsf/termbox-go"
	localMock "github.com/taylorskalyo/goreader/mock"
)

func TestInitNavigationKeys(t *testing.T) {
	p := &localMock.MockPageNavigator{}
	a := localMock.NewMockApplication(p)

	a.On("PageNavigator").Return(p)

	keymap, chmap := initNavigationKeys(a)

	var zeroValueTermboxKey termbox.Key

	verifyMethodCall := func(receiver string, methodName string, letterKey rune, tbKey termbox.Key) {
		if receiver == "pager" {
			p.On(methodName).Return()
		}
		if receiver == "app" {
			a.On(methodName).Return()
		}

		if unicode.IsLetter(letterKey) {
			chmap[letterKey]()
		}
		if tbKey != zeroValueTermboxKey {
			keymap[termbox.KeyArrowUp]()
		}

		p.AssertExpectations(t)
	}
	verifyMethodCall("pager", "ScrollUp", 'k', termbox.KeyArrowUp)
	verifyMethodCall("pager", "ScrollDown", 'j', termbox.KeyArrowDown)
	verifyMethodCall("pager", "ScrollLeft", 'h', termbox.KeyArrowLeft)
	verifyMethodCall("pager", "ScrollRight", 'l', termbox.KeyArrowRight)
	verifyMethodCall("pager", "ToTop", 'g', zeroValueTermboxKey)
	verifyMethodCall("pager", "ToBottom", 'G', zeroValueTermboxKey)
	verifyMethodCall("app", "Forward", 'f', zeroValueTermboxKey)
	verifyMethodCall("app", "Back", 'b', zeroValueTermboxKey)
	verifyMethodCall("app", "NextChapter", 'L', zeroValueTermboxKey)
	verifyMethodCall("app", "PrevChapter", 'H', zeroValueTermboxKey)
	verifyMethodCall("app", "Exit", 'q', termbox.KeyEsc)
}
