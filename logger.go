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

	GenericLoggerPooled[T any, TSerializer serializer.PooledSerializer[T]] struct {
		error  Error
		handle io.Writer
		pool   serializer.PoolInterface[T, TSerializer]
	}
)

var (
	_ Log[any] = &GenericLoggerPooled[any, *serializer.PoolJsonSerializer[any]]{}
	_ Log[any] = &GenericLogger[any, *serializer.Json[any]]{}
)

func NewWithPooledSerializer[T any, TSerializer serializer.PooledSerializer[T]](w io.Writer, serializer serializer.PoolInterface[T, TSerializer], modifiers ...Modifier[T]) *GenericLoggerPooled[T, TSerializer] {
	cfg := Config[T]{
		logger: nopErrorLog,
	}

	for _, modifier := range modifiers {
		modifier(&cfg)
	}

	l := &GenericLoggerPooled[T, TSerializer]{
		pool:   serializer,
		handle: w,
	}

	return l
}

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

func serialize[T any, TSerialize serializer.Interface[T]](handle io.Writer, errorLog Error, serializer TSerialize, data []T) error {
	rawData, err := serializer.Serialize(data)
	if err != nil {
		if errorLog != nil {
			errorLog.Print(failedToSerializeTheData, err)
		}
		return err
	}

	n, err := handle.Write(rawData)
	if err != nil {
		if errorLog != nil {
			errorLog.Print(failedToWriteToTheFile, "", err)
		}
		return err
	}

	if n != len(rawData) {
		if errorLog != nil {
			errorLog.Print(notEnoughBytesWritten, n, len(rawData))
		}
	}

	return nil
}

func (l *GenericLogger[T, TSerializer]) Log(data T) error {
	many := [...]T{data}

	return l.LogMultiple(many[:])
}

func (l *GenericLogger[T, TSerializer]) LogMultiple(data []T) error {
	return serialize(l.handle, l.error, l.serializer, data)
}

func (l *GenericLogger[T, TSerializer]) Close() error {
	if closer, ok := l.handle.(io.Closer); ok {
		return closer.Close()
	}

	return nil
}

func (l *GenericLoggerPooled[T, TSerializer]) Log(data T) error {
	many := [...]T{data}

	return l.LogMultiple(many[:])
}

func (l *GenericLoggerPooled[T, TSerializer]) LogMultiple(data []T) error {
	s := l.pool.Acquire()
	defer l.pool.Release(s)
	return serialize(l.handle, l.error, s, data)
}

func (l *GenericLoggerPooled[T, TSerializer]) Close() error {
	if closer, ok := l.handle.(io.Closer); ok {
		return closer.Close()
	}

	return nil
}
