package logging

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"syscall"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"
)

type logData struct {
	Name string `json:"name"`
}

func TestCachedLogging_Log(t *testing.T) {
	t.Parallel()
	assert := require.New(t)
	dir := t.TempDir()
	logFile := filepath.Join(dir, "log.jsonl")
	file, _ := os.Create(logFile)
	_ = file.Close()

	zerologLogger := zerolog.New(zerolog.NewTestWriter(t))

	ctx, cancel := context.WithCancel(context.Background())

	defer cancel()
	logger := New[logData](ctx, logFile, zerologLogger, syscall.SIGHUP)

	err := logger.Log(logData{Name: "test"})

	assert.NoError(err)
	assert.FileExists(logFile)
	f, _ := os.Open(logFile)
	contents, _ := io.ReadAll(f)
	assert.Contains(string(contents), "{\"name\":\"test\"}\n")
	f.Close()
}
