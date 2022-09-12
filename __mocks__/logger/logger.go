package logger

import (
	"io"

	"github.com/stretchr/testify/mock"
)

var _ io.Closer = &MockLogger{}

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


func (m *MockLogger) Close() error {
	return nil
}
