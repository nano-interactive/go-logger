//go:build unix
// +build unix

package writers

import (
	"context"
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
	ctx, cancel := context.WithCancel(context.Background())

	defer cancel()

	errCh := make(chan error, 1)
	mockWriter.On("Close").Return(errors.New("error"))

	buffer := NewSignalReopen(mockWriter, ctx, os.Interrupt, func() io.WriteCloser {
		return replaceWriter
	}, errCh)

	p, _ := os.FindProcess(os.Getpid())

	assert.NoError(p.Signal(os.Interrupt))
	assert.Equal("error", (<-errCh).Error())
	assert.NotNil(buffer)
	assert.Equal(replaceWriter, (*buffer.handle.Load()).(*writer.MockWriteCloser))
	mockWriter.AssertExpectations(t)
}
