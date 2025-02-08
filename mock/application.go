package mock

import (
	"github.com/stretchr/testify/mock"
	"github.com/taylorskalyo/goreader/nav"
)

type MockApplication struct {
	mock.Mock
	pager nav.PageNavigator
}

func NewMockApplication(p nav.PageNavigator) *MockApplication {
	return &MockApplication{pager: p}
}

func (a *MockApplication) PageNavigator() nav.PageNavigator {
	a.Called()
	return a.pager
}

func (a *MockApplication) Exit() {
	a.Called()
}

func (a *MockApplication) Forward() {
	a.Called()
}

func (a *MockApplication) Back() {
	a.Called()
}

func (a *MockApplication) NextChapter() {
	a.Called()
}

func (a *MockApplication) PrevChapter() {
	a.Called()
}

func (a *MockApplication) Err() error {
	a.Called()
	return nil
}

func (a *MockApplication) Run() int {
	a.Called()
	return 0
}
