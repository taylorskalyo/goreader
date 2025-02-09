package mock

import (
	"github.com/gdamore/tcell/v2"
	"github.com/stretchr/testify/mock"
	"github.com/taylorskalyo/goreader/parse"
)

type MockPageNavigator struct {
	mock.Mock
}

func (p *MockPageNavigator) Draw() {
	panic("not implemented") // TODO: Implement
}

func (p *MockPageNavigator) MaxScrollX() int {
	panic("not implemented") // TODO: Implement
}

func (p *MockPageNavigator) MaxScrollY() int {
	panic("not implemented") // TODO: Implement
}

func (p *MockPageNavigator) PageDown() bool {
	p.Called()
	return false
}

func (p *MockPageNavigator) PageUp() bool {
	p.Called()
	return false
}

func (p *MockPageNavigator) Pages() int {
	panic("not implemented") // TODO: Implement
}

func (p *MockPageNavigator) ScrollDown() error {
	p.Called()

	return nil
}

func (p *MockPageNavigator) ScrollLeft() error {
	p.Called()

	return nil
}

func (p *MockPageNavigator) ScrollRight() error {
	p.Called()

	return nil
}

func (p *MockPageNavigator) ScrollUp() error {
	p.Called()

	return nil
}

func (p *MockPageNavigator) SetDoc(_ parse.Cellbuf) {
	panic("not implemented") // TODO: Implement
}

func (p *MockPageNavigator) SetScreen(_ tcell.Screen) {
	panic("not implemented") // TODO: Implement
}

func (p *MockPageNavigator) Size() (int, int) {
	panic("not implemented") // TODO: Implement
}

func (p *MockPageNavigator) ToBottom() error {
	p.Called()

	return nil
}

func (p *MockPageNavigator) ToTop() error {
	p.Called()

	return nil
}

func (p *MockPageNavigator) Position() float64 {
	p.Called()

	return 0
}

func (p *MockPageNavigator) SetPosition(float64) {
	p.Called()
}
