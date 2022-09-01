package writers

import (
	"context"
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
	ctx, cancel := context.WithCancel(context.Background())

	defer cancel()

	buffer := NewSignalReopen(mockWriter, ctx, syscall.SIGHUP, func() io.WriteCloser {
		return &writer.MockWriteCloser{}
	})

	assert.NotNil(buffer)
}
