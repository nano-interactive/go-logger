package logging

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/nano-interactive/go-logger/__mocks__/logger"
)

func TestNewWithCancel(t *testing.T) {
	t.Parallel()
	assert := require.New(t)

	mockLogger := &logger.MockLogger{}

	cached := NewCached[any](
		mockLogger,
		WithBufferSize(100),
		WithWorkerPool(5),
		WithFlushRate(100),
		WithRetryCount(0),
	)

	assert.NotNil(cached)
	assert.NoError(cached.Close())
}
