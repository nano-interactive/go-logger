package logger

import (
	"io"

	"github.com/nano-interactive/go-logger/serializer"
)

var (
	_ io.Closer = &GenericLogger[any, *serializer.Json[any]]{}
	_ Log[any]  = &GenericLogger[any, *serializer.Json[any]]{}
)

type (
	Error interface {
		Print(string, ...any)
	}

	Log[T any] interface {
		Log(T) error
		LogMultiple([]T) error
	}

	GenericLogger[T any, TSerializer serializer.Interface[T]] struct {
		error      Error
		serializer TSerializer
		handle     io.Writer
	}
)

func New[T any, TSerializer serializer.Interface[T]](w io.Writer, serializer TSerializer, modifiers ...Modifier[T]) *GenericLogger[T, TSerializer] {
	cfg := Config[T]{
		logger: nopErrorLog,
	}

	for _, modifier := range modifiers {
		modifier(&cfg)
	}

	l := &GenericLogger[T, TSerializer]{
		serializer: serializer,
		handle:     w,
	}

	return l
}

//go:inline
func (l *GenericLogger[T, TSerializer]) Log(data T) error {
	many := [...]T{data}

	return l.LogMultiple(many[:])
}

func (l *GenericLogger[T, TSerializer]) LogMultiple(data []T) error {
	rawData, err := l.serializer.Serialize(data)
	if err != nil {
		l.error.Print(failedToSerializeTheData, err)
		return err
	}

	n, err := l.handle.Write(rawData)
	if err != nil {
		l.error.Print(failedToWriteToTheFile, "", err)
		return err
	}

	if n != len(rawData) {
		l.error.Print(notEnoughBytesWritten, n, len(rawData))
	}

	return nil
}

func (l *GenericLogger[T, TSerializer]) Close() error {
	if closer, ok := l.handle.(io.Closer); ok {
		return closer.Close()
	}

	return nil
}
