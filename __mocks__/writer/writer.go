package writer

import "github.com/stretchr/testify/mock"


type MockWriter struct {
	mock.Mock
}

func (w *MockWriter) Write(data []byte) (int, error) {
	args := w.Called(data)

	return args.Int(0), args.Error(1)
}
