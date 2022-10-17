package logger

import (
	"bytes"
	"context"
	"errors"
	"io"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	realSerializer "github.com/nano-interactive/go-logger/serializer"

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

	logger := &GenericLogger[logData, *serializer.MockSerializer[logData]]{
		serializer: ser,
		handle:     buff,
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

	logger := &GenericLogger[logData, *serializer.MockSerializer[logData]]{
		serializer: ser,
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

	logger := &GenericLogger[logData, *serializer.MockSerializer[logData]]{
		serializer: ser,
		handle:     buff,
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

	logger := &GenericLogger[logData, *serializer.MockSerializer[logData]]{
		serializer: ser,
		error:      l,
		handle:     buff,
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
	logger := &GenericLogger[logData, *serializer.MockSerializer[logData]]{
		serializer: ser,
		handle:     buff,
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
		_ = file.Close()
		_ = os.Remove(filepath.Join(dir, "test.log"))
	})

	l := New[Data](file, realSerializer.NewJson[Data]())
	cachedLogger := NewCached[Data](context.Background(), l, WithBufferSize(100), WithFlushRate(5))

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

func BenchmarkGenericLogger_OneByOne(b *testing.B) {
	type data struct {
		name    string
		surname string
	}

	dir := b.TempDir()
	s := realSerializer.NewJson[data]()
	file, err := os.OpenFile(filepath.Join(dir, "test.json"), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)

	if err != nil {
		b.Fatalf("Failed to open file: %v", err)
	}

	logger := New[data, *realSerializer.Json[data]](file, s)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := logger.Log(data{
			name:    "test",
			surname: "test",
		})

		if err != nil {
			b.Errorf("Failed to log the data: %v", err)
		}
	}
	b.StopTimer()

	err = logger.Close()
	if err != nil {
		b.Errorf("Failed to close the logger: %v", err)
	}

	numOfLines := b.N

	content, _ := os.ReadFile(filepath.Join(dir, "test.json"))
	lines := strings.Split(string(content), "\n")
	if lines[len(lines)-1] == "" {
		numOfLines++
	}

	if len(lines) != numOfLines {
		b.Errorf("Expected %d lines, got %d", numOfLines, len(lines))
	}
}

func BenchmarkGenericLogger_Cached(b *testing.B) {
	type data struct {
		name    string
		surname string
	}

	dir := b.TempDir()
	s := realSerializer.NewJson[data]()
	file, err := os.OpenFile(filepath.Join(dir, "test.json"), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)

	if err != nil {
		b.Fatalf("Failed to open file: %v", err)
	}

	logger := New[data, *realSerializer.Json[data]](file, s)
	cachedLogger := NewCached[data](context.Background(), logger, WithBufferSize(50), WithFlushRate(50), WithWorkerPool(2))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := cachedLogger.Log(data{
			name:    "test",
			surname: "test",
		})

		if err != nil {
			b.Errorf("Failed to log the data: %v", err)
		}
	}
	b.StopTimer()

	if err := cachedLogger.Close(); err != nil {
		b.Errorf("Failed to close the logger: %v", err)
	}

	if err != nil {
		b.Fatalf("Failed to close logger: %v", err)
	}

	numOfLines := b.N

	content, _ := os.ReadFile(filepath.Join(dir, "test.json"))
	lines := strings.Split(string(content), "\n")

	if lines[len(lines)-1] == "" {
		numOfLines++
	}

	if len(lines) != numOfLines {
		b.Errorf("Expected %d lines, got %d", numOfLines, len(lines))
	}
}

func BenchmarkGenericLogger_Batches(b *testing.B) {
	type data struct {
		name    string
		surname string
	}

	dir := b.TempDir()
	s := realSerializer.NewJson[data]()

	file, err := os.OpenFile(filepath.Join(dir, "test.json"), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)

	if err != nil {
		b.Fatalf("Failed to open file: %v", err)
	}

	logger := New[data, *realSerializer.Json[data]](file, s)

	batches := make([][]data, 0, b.N)

	numOfLines := 0
	for i := 0; i < b.N; i++ {
		items := make([]data, 0, 500)

		for j := 0; j < 500; j++ {
			items = append(items, data{
				name:    "test",
				surname: "test",
			})
		}

		batches = append(batches, items)
		numOfLines += len(items)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := logger.LogMultiple(batches[i])

		if err != nil {
			b.Errorf("Failed to log the data: %v", err)
		}
	}
	b.StopTimer()

	if err := logger.Close(); err != nil {
		b.Errorf("Failed to close the logger: %v", err)
	}

	content, _ := os.ReadFile(filepath.Join(dir, "test.json"))
	lines := strings.Split(string(content), "\n")

	if lines[len(lines)-1] == "" {
		numOfLines++
	}

	if len(lines) != numOfLines {
		b.Errorf("Expected %d lines, got %d", numOfLines, len(lines))
	}
}

func BenchmarkGenericLogger_Cached_Batches(b *testing.B) {
	type data struct {
		name    string
		surname string
	}

	dir := b.TempDir()
	s := realSerializer.NewJson[data]()

	file, err := os.OpenFile(filepath.Join(dir, "test.json"), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)

	if err != nil {
		b.Fatalf("Failed to open file: %v", err)
	}

	logger := New[data, *realSerializer.Json[data]](file, s)

	cachedLogger := NewCached[data](context.Background(), logger, WithBufferSize(1000), WithFlushRate(1000), WithWorkerPool(5))

	batches := make([][]data, 0, b.N)

	numOfLines := 0
	for i := 0; i < b.N; i++ {
		value := rand.Intn(100)
		items := make([]data, 0, value)

		for j := 0; j < value; j++ {
			items = append(items, data{
				name:    "test",
				surname: "test",
			})
		}

		batches = append(batches, items)
		numOfLines += value
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		err := cachedLogger.LogMultiple(batches[i])

		if err != nil {
			b.Errorf("Failed to log the data: %v", err)
		}
	}
	b.StopTimer()

	if err := cachedLogger.Close(); err != nil {
		b.Fatal(err)
	}

	content, _ := os.ReadFile(filepath.Join(dir, "test.json"))
	lines := strings.Split(string(content), "\n")

	if lines[len(lines)-1] == "" {
		numOfLines++
	}

	if len(lines) != numOfLines {
		b.Errorf("Expected %d lines, got %d", numOfLines, len(lines))
	}
}
