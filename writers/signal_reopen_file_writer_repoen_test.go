//go:build unix
// +build unix

package writers

import (
	"errors"
	"io"
	"os"
	"testing"

	"github.com/nano-interactive/go-logger/__mocks__/writer"
	"github.com/stretchr/testify/require"
)

func TestNewSignalReopen_ReplaceWriter_Error_On_Closer(t *testing.T) {
	t.Parallel()
	assert := require.New(t)

	mockWriter := &writer.MockWriteCloser{}
	replaceWriter := &writer.MockWriteCloser{}

	errCh := make(chan error, 1)
	mockWriter.On("Close").Return(errors.New("error"))

	buffer := NewSignalReopen(mockWriter, os.Interrupt, func() io.WriteCloser {
		return replaceWriter
	}, errCh)


	assert.NotNil(buffer)

	p, _ := os.FindProcess(os.Getpid())

	assert.NoError(p.Signal(os.Interrupt))
	assert.Equal("error", (<-errCh).Error())

	assert.Equal(replaceWriter, (*buffer.handle.Load()).(*writer.MockWriteCloser))

	mockWriter.AssertExpectations(t)
}

func TestNewSignalReopen_ReplaceWriter(t *testing.T) {
	t.Parallel()
	assert := require.New(t)

	mockWriter := &writer.MockWriteCloser{}
	replaceWriter := &writer.MockWriteCloser{}

	errCh := make(chan error, 1)

	mockWriter.On("Close").Return(nil)

	buffer := NewSignalReopen(mockWriter, os.Interrupt, func() io.WriteCloser {
		return replaceWriter
	}, errCh)

	assert.NotNil(buffer)

	p, _ := os.FindProcess(os.Getpid())

	assert.NoError(p.Signal(os.Interrupt))
	assert.Nil(<-errCh)
	assert.NotNil(buffer)
	assert.Equal(replaceWriter, (*buffer.handle.Load()).(*writer.MockWriteCloser))
	mockWriter.AssertExpectations(t)
}
