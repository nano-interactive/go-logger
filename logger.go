package logger

import (
	"io"

	"github.com/nano-interactive/go-logger/serializer"
)

var (
	_ io.Closer = &Logger[any, *serializer.Json[any]]{}
	_ Log[any]  = &Logger[any, *serializer.Json[any]]{}
)

type (
	Error interface {
		Print(string, ...any)
	}

	Log[T any] interface {
		Log(T) error
		LogMultiple([]T) error
	}

	Logger[T any, TSerializer serializer.Interface[T]] struct {
		error      Error
		serializer TSerializer
		handle     io.Writer
	}
)

func New[T any, TSerializer serializer.Interface[T]](w io.Writer, serializer TSerializer, modifiers ...Modifier[T]) *Logger[T, TSerializer] {
	cfg := Config[T]{
		logger: nopErrorLog,
	}

	for _, modifier := range modifiers {
		modifier(&cfg)
	}

	l := &Logger[T, TSerializer]{
		serializer: serializer,
		handle:     w,
	}

	return l
}

//go:inline
func (l *Logger[T, TSerializer]) Log(data T) error {
	many := [...]T{data}

	return l.LogMultiple(many[:])
}

const notEnoughBytesWritten = "{\"msg\":\"failed to write all data to the writer\",\"actualLen\":%d,\"expectedLen\":%d}"

func (l *Logger[T, TSerializer]) LogMultiple(data []T) error {
	rawData, err := l.serializer.Serialize(data)
	if err != nil {
		return err
	}

	n, err := l.handle.Write(rawData)
	if err != nil {
		return err
	}

	if n != len(rawData) {
		l.error.Print(notEnoughBytesWritten, n, len(rawData))
	}

	return nil
}

func (l *Logger[T, TSerializer]) Close() error {
	if closer, ok := l.handle.(io.Closer); ok {
		return closer.Close()
	}

	return nil
}
