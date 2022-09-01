package writer

import "github.com/stretchr/testify/mock"

type MockWriteCloser struct {
	mock.Mock
}

func (w *MockWriteCloser) Write(data []byte) (int, error) {
	args := w.Called(data)

	return args.Int(0), args.Error(1)
}

func (w *MockWriteCloser) Close() error {
	args := w.Called()

	return args.Error(0)
}
