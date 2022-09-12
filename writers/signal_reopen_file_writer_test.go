package writers

import (
	"io"
	"syscall"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/nano-interactive/go-logger/__mocks__/writer"
)

func TestNewSignalReopen(t *testing.T) {
	t.Parallel()
	assert := require.New(t)

	mockWriter := &writer.MockWriteCloser{}

	mockWriter.On("Close").Return(nil)

	buffer := NewSignalReopen(mockWriter, syscall.SIGHUP, func() io.WriteCloser {
		return &writer.MockWriteCloser{}
	})

	assert.NotNil(buffer)
	assert.NoError(buffer.Close())

	mockWriter.AssertExpectations(t)
}

func TestSignalReopenWriter_Write(t *testing.T) {
	t.Parallel()
	assert := require.New(t)

	mockWriter := &writer.MockWriteCloser{}

	data := []byte{0xA}

	mockWriter.On("Write", data).Return(1, nil)
	mockWriter.On("Close").Return(nil)

	buffer := NewSignalReopen(mockWriter, syscall.SIGHUP, func() io.WriteCloser {
		return &writer.MockWriteCloser{}
	})

	n, err := buffer.Write(data)

	assert.NoError(err)
	assert.Equal(1, n)

	assert.NotNil(buffer)
	assert.NoError(buffer.Close())

	mockWriter.AssertExpectations(t)
}
