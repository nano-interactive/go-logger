package logging

import (
	"bytes"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/nano-interactive/go-logger/__mocks__/error_log"
	"github.com/nano-interactive/go-logger/__mocks__/serializer"
	"github.com/nano-interactive/go-logger/__mocks__/writer"
)

type logData struct {
	Name string `json:"name"`
}

func TestLogMultiple(t *testing.T) {
	t.Parallel()
	assert := require.New(t)

	buff := bytes.NewBuffer(make([]byte, 0, 100))

	ser := &serializer.MockSerializer[logData]{}

	data := []logData{
		{Name: "test 1"},
		{Name: "test 2"},
	}

	logger := &Logger[logData, *serializer.MockSerializer[logData]]{
		serializer: ser,
		handle:     buff,
		delimiter:  '\n',
	}

	ser.
		On("SerializeMultipleWithDelimiter", data, '\n').
		Return([]byte{0x1, '\n', 0x2, '\n'}, nil)

	err := logger.LogMultiple(data)

	assert.NoError(err)

	assert.EqualValues([]byte{0x1, '\n', 0x2, '\n'}, buff.Bytes())
	ser.AssertExpectations(t)
}

func TestLogMultipleErrorSerializer(t *testing.T) {
	t.Parallel()
	assert := require.New(t)

	ser := &serializer.MockSerializer[logData]{}

	data := []logData{
		{Name: "test 1"},
		{Name: "test 2"},
	}

	logger := &Logger[logData, *serializer.MockSerializer[logData]]{
		serializer: ser,
		delimiter:  '\n',
	}

	ser.
		On("SerializeMultipleWithDelimiter", data, '\n').
		Return([]byte{}, errors.New("failed to serialize"))

	err := logger.LogMultiple(data)

	assert.Error(err)
	assert.Equal("failed to serialize", err.Error())

	ser.AssertExpectations(t)
}

func TestLogMultipleErrorWithWriter(t *testing.T) {
	t.Parallel()
	assert := require.New(t)

	buff := &writer.MockWriteCloser{}

	ser := &serializer.MockSerializer[logData]{}

	data := []logData{
		{Name: "test 1"},
		{Name: "test 2"},
	}

	logger := &Logger[logData, *serializer.MockSerializer[logData]]{
		serializer: ser,
		handle:     buff,
		delimiter:  '\n',
	}

	ser.
		On("SerializeMultipleWithDelimiter", data, '\n').
		Return([]byte{0x1, '\n', 0x2, '\n'}, nil)

	buff.
		On("Write", []byte{0x1, '\n', 0x2, '\n'}).
		Return(0, errors.New("failed to write"))

	err := logger.LogMultiple(data)

	assert.Error(err)
	assert.Equal("failed to write", err.Error())

	ser.AssertExpectations(t)
	buff.AssertExpectations(t)
}

func TestLogMultipleNotEnoughBytesWritten(t *testing.T) {
	t.Parallel()
	assert := require.New(t)

	buff := &writer.MockWriteCloser{}

	ser := &serializer.MockSerializer[logData]{}

	data := []logData{
		{Name: "test 1"},
		{Name: "test 2"},
	}

	l := error_log.NewMockLogger()

	logger := &Logger[logData, *serializer.MockSerializer[logData]]{
		serializer: ser,
		error:      l,
		handle:     buff,
		delimiter:  '\n',
	}

	ser.
		On("SerializeMultipleWithDelimiter", data, '\n').
		Return([]byte{0x1, '\n', 0x2, '\n'}, nil)

	buff.
		On("Write", []byte{0x1, '\n', 0x2, '\n'}).
		Return(3, nil)

	err := logger.LogMultiple(data)

	assert.NoError(err)
	assert.EqualValues([]string{"{\"msg\":\"failed to write all data to the writer\",\"actualLen\":3,\"expectedLen\":4}"}, l.Buffer)

	ser.AssertExpectations(t)
	buff.AssertExpectations(t)
}

func TestLog(t *testing.T) {
	t.Parallel()
	assert := require.New(t)

	buff := bytes.NewBuffer(make([]byte, 0, 100))

	ser := &serializer.MockSerializer[logData]{}
	logger := &Logger[logData, *serializer.MockSerializer[logData]]{
		serializer: ser,
		handle:     buff,
		delimiter:  '\n',
	}

	ser.
		On("SerializeMultipleWithDelimiter", []logData{{Name: "test 1"}}, '\n').
		Return([]byte{0x1, '\n', 0x2}, nil)

	err := logger.Log(logData{Name: "test 1"})

	assert.NoError(err)

	assert.EqualValues([]byte{0x1, '\n', 0x2}, buff.Bytes())
	ser.AssertExpectations(t)
}
