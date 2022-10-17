package serializer

import "github.com/stretchr/testify/mock"

type MockSerializer[T any] struct {
	mock.Mock
}

func (m *MockSerializer[T]) Serialize(data []T) ([]byte, error) {
	args := m.Called(data)

	result := args.Get(0)

	if result != nil {
		return result.([]byte), args.Error(1)
	}

	return nil, args.Error(1)
}
