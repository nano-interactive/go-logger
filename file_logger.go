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
)

func NewFileLogger[T any, TSerializer serializer.Interface[T]](path string, flags int, mode os.FileMode, serializer TSerializer, error ...Error) *FileLogger[T, TSerializer] {
	var errLog Error

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
func (l *FileLogger[T, TSerializer]) Log(data T) error {
	many := [...]T{data}

	return l.LogMultiple(many[:])
}

func (l *FileLogger[T, TSerializer]) LogMultiple(data []T) error {
	rawData, err := l.serializer.Serialize(data)
	if err != nil {
		if l.error != nil {
			l.error.Print(failedToSerializeTheData, err)
		}
		return err
	}

	file, err := os.OpenFile(l.path, l.flags, l.mode)

	if err != nil {
		if l.error != nil {
			l.error.Print(failedToOpenFile, l.path, err)
		}
		return err
	}

	defer func(file *os.File) {
		err := file.Close()
		if err != nil && l.error != nil {
			l.error.Print(failedToCloseTheFile, l.path, err)
		}
	}(file)

	n, err := file.Write(rawData)
	if err != nil {
		if l.error != nil {
			l.error.Print(failedToWriteToTheFile, l.path, err)
		}
		return err
	}

	if n != len(rawData) && l.error != nil {
		l.error.Print(notEnoughBytesWritten, n, len(rawData))
	}

	return nil
}
