package logger

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/nano-interactive/go-logger/__mocks__/error_log"
	"github.com/nano-interactive/go-logger/__mocks__/serializer"
	realSerializer "github.com/nano-interactive/go-logger/serializer"
)

func TestFileLogger_Log_NoError(t *testing.T) {
	t.Parallel()
	assert := require.New(t)

	dir := t.TempDir()

	path := filepath.Join(dir, "test.json")
	logger := error_log.NewMockLogger()
	mockSerializer := &serializer.MockSerializer[any]{}

	mockSerializer.On("Serialize", []any{"test"}).Return([]byte("test"), nil)

	fileLogger := FileLogger[any, *serializer.MockSerializer[any]]{
		error:      logger,
		serializer: mockSerializer,
		path:       path,
		flags:      os.O_CREATE | os.O_WRONLY | os.O_APPEND,
		mode:       0644,
	}

	err := fileLogger.Log("test")

	assert.NoError(err)
	assert.Empty(logger.Buffer)

	content, err := os.ReadFile(path)

	assert.NoError(err)
	assert.EqualValues([]byte("test"), content)
}

func TestFileLogger_Log_SerializerError(t *testing.T) {
	t.Parallel()
	assert := require.New(t)

	dir := t.TempDir()

	logger := error_log.NewMockLogger()
	mockSerializer := &serializer.MockSerializer[any]{}

	mockSerializer.On("Serialize", []any{"test"}).Return(nil, errors.New("failed to serialize"))

	fileLogger := FileLogger[any, *serializer.MockSerializer[any]]{
		error:      logger,
		serializer: mockSerializer,
		path:       filepath.Join(dir, "test.json"),
		flags:      os.O_CREATE | os.O_WRONLY | os.O_APPEND,
		mode:       0644,
	}

	err := fileLogger.Log("test")

	assert.Error(err)
	assert.EqualError(err, "failed to serialize")
	assert.Len(logger.Buffer, 1)
	assert.Equal(fmt.Sprintf(failedToSerializeTheData, err), logger.Buffer[0])
}

func BenchmarkFileLogger_OneByOne(b *testing.B) {
	type data struct {
		name    string
		surname string
	}

	dir := b.TempDir()
	s := realSerializer.NewJson[data]()
	logger := NewFileLogger[data](filepath.Join(dir, "test.json"), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644, s)

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

func BenchmarkFileLogger_Cached(b *testing.B) {
	type data struct {
		name    string
		surname string
	}

	dir := b.TempDir()
	s := realSerializer.NewJson[data]()
	logger := NewFileLogger[data](filepath.Join(dir, "test.json"), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644, s)

	cachedLogger := NewCached[data](context.Background(), logger, WithBufferSize(50), WithWorkerPool(2))

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

	err := cachedLogger.Close()
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

func BenchmarkFileLogger_Batches(b *testing.B) {
	type data struct {
		name    string
		surname string
	}

	dir := b.TempDir()
	s := realSerializer.NewJson[data]()
	logger := NewFileLogger[data](filepath.Join(dir, "test.json"), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644, s)

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

	content, _ := os.ReadFile(filepath.Join(dir, "test.json"))
	lines := strings.Split(string(content), "\n")

	if lines[len(lines)-1] == "" {
		numOfLines++
	}

	if len(lines) != numOfLines {
		b.Errorf("Expected %d lines, got %d", numOfLines, len(lines))
	}
}

func BenchmarkFileLogger_Cached_Batches(b *testing.B) {
	type data struct {
		name    string
		surname string
	}

	dir := b.TempDir()
	s := realSerializer.NewJson[data]()
	logger := NewFileLogger[data](filepath.Join(dir, "test.json"), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644, s)

	cachedLogger := NewCached[data](context.Background(), logger, WithBufferSize(1000),  WithWorkerPool(5))

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

	err := cachedLogger.Close()
	if err != nil {
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
