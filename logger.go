package logging

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"os"
	"os/signal"
	"sync"
	"sync/atomic"

	"github.com/rs/zerolog"
)

const (
	defaultBufferSize      = 8192
	defaultFilePermissions = 0o644
)

var _ io.Closer = &Logger[any]{}

type (
	Log[T any] interface {
		Log(T) error
		LogMultiple([]T) error
	}

	Logger[T any] struct {
		file   string
		logger zerolog.Logger
		pool   *sync.Pool
		handle *atomic.Pointer[os.File]
	}
)

// #nosec G304
func open(file string, logger zerolog.Logger) *os.File {
	f, err := os.OpenFile(file, os.O_CREATE|os.O_APPEND|os.O_WRONLY, defaultFilePermissions)
	if err != nil {
		logger.Fatal().
			Err(err).
			Str("file", file).
			Str("flags", "O_CREATE|O_APPEND|O_WRONLY").
			Msg("failed to open log file")
	}

	return f
}

func New[T any](ctx context.Context, file string, logger zerolog.Logger, reopenSignal os.Signal) *Logger[T] {
	ch := make(chan os.Signal, 1000)

	signal.Notify(ch, reopenSignal)

	l := &Logger[T]{
		file:   file,
		logger: logger,
		pool: &sync.Pool{
			New: func() any {
				return bytes.NewBuffer(make([]byte, 0, defaultBufferSize))
			},
		},
		handle: new(atomic.Pointer[os.File]),
	}

	l.handle.Store(open(file, logger))

	go func() {
		closeFile := func(h *os.File) {
			if err := h.Close(); err != nil {
				logger.Fatal().
					Err(err).
					Str("file", file).
					Str("flags", "O_CREATE|O_APPEND|O_WRONLY").
					Msg("Failed to close old log file handle")
			}
		}

		for {
			select {
			case <-ctx.Done():
				return
			case <-ch:
				logger.Info().Msg("reopening log file")
				old := l.handle.Swap(open(file, logger))
				closeFile(old)
			}
		}
	}()

	return l
}

//go:inline
func (l *Logger[T]) Log(data T) error {
	many := make([]T, 0, 1)
	many = append(many, data)

	return l.LogMultiple(many[:])
}

func (l *Logger[T]) LogMultiple(data []T) error {
	buffer := l.pool.Get().(*bytes.Buffer)
	defer buffer.Reset()
	defer l.pool.Put(buffer)

	for _, d := range data {
		data, err := json.Marshal(d)
		if err != nil {
			l.logger.Error().
				Interface("data", d).
				Msg("Failed to serialize data")
			continue
		}

		buffer.Write(data)
		buffer.WriteByte('\n')
	}

	rawData := buffer.Bytes()

	file := l.handle.Load()

	n, err := file.Write(rawData)
	if err != nil {
		l.logger.Error().
			Err(err).
			Str("file", l.file).
			Str("flags", "O_CREATE|O_APPEND|O_WRONLY").
			Int("expectedLen", len(rawData)).
			Msg("Failed to Write data to file")
		return err
	}

	if n != len(rawData) {
		l.logger.Warn().
			Str("file", l.file).
			Int("expected", len(rawData)).
			Int("actual", n).
			Msg("Failed to write all data")
	}

	return nil
}

func (l *Logger[T]) Close() error {
	file := l.handle.Load()

	return file.Close()
}
