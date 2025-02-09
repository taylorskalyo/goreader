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

func (a *MockApplication) Exit() error {
	a.Called()

	return nil
}

func (a *MockApplication) Forward() error {
	a.Called()

	return nil
}

func (a *MockApplication) Back() error {
	a.Called()

	return nil
}

func (a *MockApplication) GotoChapter(int) error {
	a.Called()

	return nil
}

func (a *MockApplication) NextChapter() error {
	a.Called()

	return nil
}

func (a *MockApplication) PrevChapter() error {
	a.Called()

	return nil
}

func (a *MockApplication) Run() error {
	a.Called()

	return nil
}
