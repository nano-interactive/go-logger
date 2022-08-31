package serializer

import "github.com/stretchr/testify/mock"

type MockSerializer[T any] struct {
	mock.Mock
}

func (m *MockSerializer[T]) Serialize(data T) ([]byte, error) {
	args := m.Called(data)

	return args.Get(0).([]byte), args.Error(1)
}

func (m *MockSerializer[T]) SerializeMultipleWithDelimiter(data []T, delimiter rune) ([]byte, error) {
	args := m.Called(data, delimiter)

	return args.Get(0).([]byte), args.Error(1)
}
