package logging

import (
	"context"
	"io"
	"sync/atomic"

	"github.com/nano-interactive/go-logger/serializer"
)

var (
	_ io.Closer = &Logger[any]{}
	_ Log[any]  = &Logger[any]{}
)

type (
	Error interface {
		Print(string, ...any)
	}

	Log[T any] interface {
		Log(T) error
		LogMultiple([]T) error
		swap(io.Writer) *io.Writer
	}

	Logger[T any] struct {
		error      Error
		ctx        context.Context
		serializer serializer.Interface[T]
		handle     *atomic.Pointer[io.Writer]
		delimiter  rune
	}
)

// #nosec G304
// func open(file string, logger zerolog.Logger) *os.File {
// 	f, err := os.OpenFile(file, os.O_CREATE|os.O_APPEND|os.O_WRONLY, defaultFilePermissions)
// 	if err != nil {
// 		logger.Fatal().
// 			Err(err).
// 			Str("file", file).
// 			Str("flags", "O_CREATE|O_APPEND|O_WRONLY").
// 			Msg("failed to open log file")
// 	}

// 	return f
// }

func New[T any](w io.Writer, modifiers ...Modifier[T]) *Logger[T] {
	cfg := Config[T]{
		ctx:        context.Background(),
		serializer: serializer.NewJson[T](),
		logger:     nopErrorLog,
		delimiter:  '\n',
	}

	for _, modifier := range modifiers {
		modifier(&cfg)
	}

	l := &Logger[T]{
		ctx:        cfg.ctx,
		serializer: cfg.serializer,
		handle:     new(atomic.Pointer[io.Writer]),
	}

	l.handle.Store(&w)

	return l

	// go func() {
	// 	closeFile := func(h *os.File) {
	// 		if err := h.Close(); err != nil {
	// 			logger.Fatal().
	// 				Err(err).
	// 				Str("file", file).
	// 				Str("flags", "O_CREATE|O_APPEND|O_WRONLY").
	// 				Msg("Failed to close old log file handle")
	// 		}
	// 	}

	// 	for {
	// 		select {
	// 		case <-ctx.Done():
	// 			return
	// 		case <-ch:
	// 			logger.Info().Msg("reopening log file")
	// 			old := l.handle.Swap(open(file, logger))
	// 			closeFile(old)
	// 		}
	// 	}
	// }()
}

//go:inline
func (l *Logger[T]) Log(data T) error {
	many := make([]T, 0, 1)
	many = append(many, data)

	return l.LogMultiple(many[:])
}

const notEnoughBytesWritten = "{\"msg\":\"failed to write all data to the writer\",\"actualLen\":%d,\"expectedLen\":%d}"

func (l *Logger[T]) LogMultiple(data []T) error {
	rawData, err := l.serializer.SerializeMultipleWithDelimiter(data, l.delimiter)
	if err != nil {
		return err
	}

	file := l.handle.Load()

	n, err := (*file).Write(rawData)
	if err != nil {
		return err
	}

	if n != len(rawData) {
		l.error.Print(notEnoughBytesWritten, n, len(rawData))
	}

	return nil
}

func (l *Logger[T]) swap(w io.Writer) *io.Writer {
	return l.handle.Swap(&w)
}

func (l *Logger[T]) Close() error {
	file := l.handle.Load()

	if closer, ok := (*file).(io.Closer); ok {
		return closer.Close()
	}

	return nil
}
