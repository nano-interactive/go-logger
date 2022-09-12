package logger

import (
	"bytes"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	real_serializer "github.com/nano-interactive/go-logger/serializer"

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
		On("Serialize", data).
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
		On("Serialize", data).
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
		On("Serialize", data).
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
		On("Serialize", data).
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
		On("Serialize", []logData{{Name: "test 1"}}).
		Return([]byte{0x1, '\n', 0x2}, nil)

	err := logger.Log(logData{Name: "test 1"})

	assert.NoError(err)

	assert.EqualValues([]byte{0x1, '\n', 0x2}, buff.Bytes())
	ser.AssertExpectations(t)
}

type Data struct {
	Name string `json:"name"`
}

func TestLoggerIntegration(t *testing.T) {
	t.Parallel()
	assert := require.New(t)
	dir := t.TempDir()

	file, _ := os.OpenFile(filepath.Join(dir, "test.log"), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)

	t.Cleanup(func() {
		file.Close()
		os.Remove(filepath.Join(dir, "test.log"))
	})

	l := New[Data](file,real_serializer.NewJson[Data]())
	cachedLogger := NewCached[Data](l, WithBufferSize(100), WithFlushRate(5))

	assert.NoError(cachedLogger.Log(Data{Name: "test1"}))
	assert.NoError(cachedLogger.Log(Data{Name: "test2"}))
	assert.NoError(cachedLogger.Log(Data{Name: "test3"}))
	assert.NoError(cachedLogger.Log(Data{Name: "test4"}))
	assert.NoError(cachedLogger.Log(Data{Name: "test5"}))

	time.Sleep(1 * time.Second)
	_ = cachedLogger.Close()

	file, _ = os.OpenFile(filepath.Join(dir, "test.log"), os.O_RDONLY, 0o644)
	bytes, _ := io.ReadAll(file)

	lines := strings.Split(strings.TrimRight(string(bytes), "\n"), "\n")

	assert.Len(lines, 5)

	assert.EqualValues([]string{
		`{"name":"test1"}`,
		`{"name":"test2"}`,
		`{"name":"test3"}`,
		`{"name":"test4"}`,
		`{"name":"test5"}`,
	}, lines)
}
