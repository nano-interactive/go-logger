package logger

import (
	"os"

	"github.com/nano-interactive/go-logger/serializer"
)

type (
	FileLogger[T any, TSerializer serializer.Interface[T]] struct {
		error      Error
		serializer TSerializer
		path       string
		flags      int
		mode       os.FileMode
	}

	FileLoggerPooled[T any, TSerializer serializer.PooledSerializer[T]] struct {
		pool  serializer.PoolInterface[T, TSerializer]
		error Error
		path  string
		flags int
		mode  os.FileMode
	}
)

var (
	_ Log[any] = &FileLogger[any, *serializer.Json[any]]{}
	_ Log[any] = &FileLoggerPooled[any, *serializer.PoolJsonSerializer[any]]{}
)

func NewFileLoggerWithPoolSerializer[T any, TSerializer serializer.PooledSerializer[T]](path string, flags int, mode os.FileMode, serializer serializer.PoolInterface[T, TSerializer], error ...Error) *FileLoggerPooled[T, TSerializer] {
	var errLog Error = nil

	if len(error) > 0 {
		errLog = error[0]
	}

	return &FileLoggerPooled[T, TSerializer]{
		path:  path,
		flags: flags,
		mode:  mode,
		error: errLog,
		pool:  serializer,
	}
}

func NewFileLogger[T any, TSerializer serializer.Interface[T]](path string, flags int, mode os.FileMode, serializer TSerializer, error ...Error) *FileLogger[T, TSerializer] {
	var errLog Error = nil

	if len(error) > 0 {
		errLog = error[0]
	}

	return &FileLogger[T, TSerializer]{
		serializer: serializer,
		path:       path,
		flags:      flags,
		mode:       mode,
		error:      errLog,
	}
}

//go:inline
func serializeToFile[T any, TSerializer serializer.Interface[T]](errorLog Error, path string, flags int, mode os.FileMode, serializer TSerializer, data []T) error {
	rawData, err := serializer.Serialize(data)
	if err != nil {
		if errorLog != nil {
			errorLog.Print(failedToSerializeTheData, err)
		}
		return err
	}

	file, err := os.OpenFile(path, flags, mode)

	if err != nil {
		if errorLog != nil {
			errorLog.Print(failedToOpenFile, path, err)
		}
		return err
	}

	defer func(file *os.File) {
		err := file.Close()
		if err != nil && errorLog != nil {
			errorLog.Print(failedToCloseTheFile, path, err)
		}
	}(file)

	n, err := file.Write(rawData)
	if err != nil {
		if errorLog != nil {
			errorLog.Print(failedToWriteToTheFile, path, err)
		}
		return err
	}

	if n != len(rawData) && errorLog != nil {
		errorLog.Print(notEnoughBytesWritten, n, len(rawData))
	}

	return nil
}

func (l *FileLogger[T, TSerializer]) LogMultiple(data []T) error {
	return serializeToFile(l.error, l.path, l.flags, l.mode, l.serializer, data)
}

func (l *FileLogger[T, TSerializer]) Log(data T) error {
	many := [...]T{data}

	return l.LogMultiple(many[:])
}

func (l *FileLoggerPooled[T, TSerializer]) Log(data T) error {
	many := [...]T{data}
	return l.LogMultiple(many[:])
}

func (l *FileLoggerPooled[T, TSerializer]) LogMultiple(data []T) error {
	s := l.pool.Acquire()
	defer l.pool.Release(s)
	return serializeToFile(l.error, l.path, l.flags, l.mode, s, data)
}
