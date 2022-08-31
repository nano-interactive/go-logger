package logging

import (
	"bytes"
	"errors"
	"io"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/nano-interactive/go-logger/__mocks__/serializer"
	"github.com/nano-interactive/go-logger/__mocks__/writer"
)

type logData struct {
	Name string `json:"name"`
}

func TestLogMultiple(t *testing.T) {
	t.Parallel()
	assert := require.New(t)

	w := new(atomic.Pointer[io.Writer])
	buff := bytes.NewBuffer(make([]byte, 0, 100))
	var writer io.Writer = buff
	w.Store(&writer)

	serializer := &serializer.MockSerializer[logData]{}

	data := []logData{
		{Name: "test 1"},
		{Name: "test 2"},
	}

	logger := &Logger[logData]{
		serializer: serializer,
		handle:     w,
		delimiter:  '\n',
	}

	serializer.
		On("SerializeMultipleWithDelimiter", data, '\n').
		Return([]byte{0x1, '\n', 0x2, '\n'}, nil)

	err := logger.LogMultiple(data)

	assert.NoError(err)

	assert.EqualValues([]byte{0x1, '\n', 0x2, '\n'}, buff.Bytes())
	serializer.AssertExpectations(t)
}

func TestLogMultipleErrorSerializer(t *testing.T) {
	t.Parallel()
	assert := require.New(t)

	serializer := &serializer.MockSerializer[logData]{}

	data := []logData{
		{Name: "test 1"},
		{Name: "test 2"},
	}

	logger := &Logger[logData]{
		serializer: serializer,
		delimiter:  '\n',
	}

	serializer.
		On("SerializeMultipleWithDelimiter", data, '\n').
		Return([]byte{}, errors.New("failed to serialize"))

	err := logger.LogMultiple(data)

	assert.Error(err)
	assert.Equal("failed to serialize", err.Error())

	serializer.AssertExpectations(t)
}

func TestLogMultipleErrorWithWriter(t *testing.T) {
	t.Parallel()
	assert := require.New(t)

	w := new(atomic.Pointer[io.Writer])
	buff := &writer.MockWriter{}
	var writer io.Writer = buff
	w.Store(&writer)

	serializer := &serializer.MockSerializer[logData]{}

	data := []logData{
		{Name: "test 1"},
		{Name: "test 2"},
	}

	logger := &Logger[logData]{
		serializer: serializer,
		handle:     w,
		delimiter:  '\n',
	}

	serializer.
		On("SerializeMultipleWithDelimiter", data, '\n').
		Return([]byte{0x1, '\n', 0x2, '\n'}, nil)

	buff.
		On("Write", []byte{0x1, '\n', 0x2, '\n'}).
		Return(0, errors.New("failed to write"))

	err := logger.LogMultiple(data)

	assert.Error(err)
	assert.Equal("failed to write", err.Error())

	serializer.AssertExpectations(t)
	buff.AssertExpectations(t)
}

func TestLogMultipleNotEnoughBytesWritten(t *testing.T) {
	t.Parallel()
	assert := require.New(t)

	w := new(atomic.Pointer[io.Writer])
	buff := &writer.MockWriter{}
	var writer io.Writer = buff
	w.Store(&writer)

	serializer := &serializer.MockSerializer[logData]{}

	data := []logData{
		{Name: "test 1"},
		{Name: "test 2"},
	}

	l := newMockLogger()

	logger := &Logger[logData]{
		serializer: serializer,
		error:      l,
		handle:     w,
		delimiter:  '\n',
	}

	serializer.
		On("SerializeMultipleWithDelimiter", data, '\n').
		Return([]byte{0x1, '\n', 0x2, '\n'}, nil)

	buff.
		On("Write", []byte{0x1, '\n', 0x2, '\n'}).
		Return(3, nil)

	err := logger.LogMultiple(data)

	assert.NoError(err)
	assert.EqualValues([]string{"{\"msg\":\"failed to write all data to the writer\",\"actualLen\":3,\"expectedLen\":4}"}, l.Buffer)

	serializer.AssertExpectations(t)
	buff.AssertExpectations(t)
}

func TestLog(t *testing.T) {
	t.Parallel()
	assert := require.New(t)

	w := new(atomic.Pointer[io.Writer])
	buff := bytes.NewBuffer(make([]byte, 0, 100))
	var writer io.Writer = buff
	w.Store(&writer)

	serializer := &serializer.MockSerializer[logData]{}
	logger := &Logger[logData]{
		serializer: serializer,
		handle:     w,
		delimiter:  '\n',
	}

	serializer.
		On("SerializeMultipleWithDelimiter", []logData{{Name: "test 1"}}, '\n').
		Return([]byte{0x1, '\n', 0x2}, nil)

	err := logger.Log(logData{Name: "test 1"})

	assert.NoError(err)

	assert.EqualValues([]byte{0x1, '\n', 0x2}, buff.Bytes())
	serializer.AssertExpectations(t)
}
