package app

import (
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/stretchr/testify/mock"
	localMock "github.com/taylorskalyo/goreader/mock"
)

func TestInitNavigationKeys(t *testing.T) {
	p := &localMock.MockPageNavigator{}
	a := localMock.NewMockApplication(p)

	a.On("PageNavigator").Return(p)

	keymap, chmap := initNavigationKeys(a)

	verifyMethodCall := func(receiver *mock.Mock, methodName string, input any) {
		receiver.On(methodName).Return()

		switch v := input.(type) {
		case rune:
			if fn, ok := chmap[v]; ok {
				fn()
			} else {
				t.Errorf("unhandled input character: %c", v)
			}
		case tcell.Key:
			if fn, ok := keymap[v]; ok {
				fn()
			} else {
				t.Errorf("unhandled input key: %#x", input)
			}
		default:
			t.Errorf("unhandled input: %+v", input)
		}

		p.AssertExpectations(t)
	}

	verifyMethodCall(&p.Mock, "ScrollUp", tcell.KeyUp)
	verifyMethodCall(&p.Mock, "ScrollUp", 'k')
	verifyMethodCall(&p.Mock, "ScrollDown", tcell.KeyDown)
	verifyMethodCall(&p.Mock, "ScrollDown", 'j')
	verifyMethodCall(&p.Mock, "ScrollLeft", tcell.KeyLeft)
	verifyMethodCall(&p.Mock, "ScrollLeft", 'h')
	verifyMethodCall(&p.Mock, "ScrollRight", tcell.KeyRight)
	verifyMethodCall(&p.Mock, "ScrollRight", 'l')
	verifyMethodCall(&p.Mock, "ToTop", 'g')
	verifyMethodCall(&p.Mock, "ToBottom", 'G')

	verifyMethodCall(&a.Mock, "Exit", tcell.KeyEsc)
	verifyMethodCall(&a.Mock, "Exit", 'q')
	verifyMethodCall(&a.Mock, "Forward", 'f')
	verifyMethodCall(&a.Mock, "Back", 'b')
	verifyMethodCall(&a.Mock, "NextChapter", 'L')
	verifyMethodCall(&a.Mock, "PrevChapter", 'H')
}
