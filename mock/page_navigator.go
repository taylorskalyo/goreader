package mock

import (
	"github.com/stretchr/testify/mock"
	"github.com/taylorskalyo/goreader/parse"
)

type MockPageNavigator struct {
	mock.Mock
}

func (p *MockPageNavigator) Draw() error {
	panic("not implemented") // TODO: Implement
}

func (p *MockPageNavigator) MaxScrollX() int {
	panic("not implemented") // TODO: Implement
}

func (p *MockPageNavigator) MaxScrollY() int {
	panic("not implemented") // TODO: Implement
}

func (p *MockPageNavigator) PageDown() bool {
	panic("not implemented") // TODO: Implement
}

func (p *MockPageNavigator) PageUp() bool {
	panic("not implemented") // TODO: Implement
}

func (p *MockPageNavigator) Pages() int {
	panic("not implemented") // TODO: Implement
}

func (p *MockPageNavigator) ScrollDown() {
	p.Called()
}

func (p *MockPageNavigator) ScrollLeft() {
	p.Called()
}

func (p *MockPageNavigator) ScrollRight() {
	p.Called()
}

func (p *MockPageNavigator) ScrollUp() {
	p.Called()
}

func (p *MockPageNavigator) SetDoc(_ parse.Cellbuf) {
	panic("not implemented") // TODO: Implement
}

func (p *MockPageNavigator) Size() (int, int) {
	panic("not implemented") // TODO: Implement
}

func (p *MockPageNavigator) ToBottom() {
	panic("not implemented") // TODO: Implement
}

func (p *MockPageNavigator) ToTop() {
	panic("not implemented") // TODO: Implement
}
