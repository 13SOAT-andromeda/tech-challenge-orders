package mocks

import "github.com/stretchr/testify/mock"

type MockEmail struct {
	mock.Mock
}

func (m *MockEmail) Send(name string, email string, subject string, html string) error {
	args := m.Called(name, email, subject, html)

	return args.Error(0)
}
