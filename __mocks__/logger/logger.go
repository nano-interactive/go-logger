package logger

import "github.com/stretchr/testify/mock"

type MockLogger struct {
	mock.Mock
}

func (m *MockLogger) Log(data interface{}) error {
	args := m.Called(data)

	return args.Error(0)
}

func (m *MockLogger) LogMultiple(data []interface{}) error {
	args := m.Called(data)
	return args.Error(0)
}
