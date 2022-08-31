package logging

// import (
// 	"context"
// 	"path/filepath"
// 	"syscall"
// 	"testing"

// 	"github.com/rs/zerolog"
// )

// type empty struct{}

// func BenchmarkCachedLogging_Log(b *testing.B) {
// 	b.ReportAllocs()

// 	dir := b.TempDir()

// 	file := filepath.Join(dir, "log.jsonl")

// 	zerologLogger := zerolog.Nop()

// 	ctx, cancel := context.WithCancel(context.Background())

// 	logger := NewCached[empty](ctx, zerologLogger, New[empty](context.Background(), file, zerologLogger, syscall.SIGHUP), CachedLoggingConfig{
// 		Workers:    2,
// 		BufferSize: 1000,
// 		FlushRate:  500,
// 	})

// 	b.ResetTimer()
// 	for i := 0; i < b.N; i++ {
// 		err := logger.Log(empty{})
// 		if err != nil {
// 			b.Fatal(err)
// 		}
// 	}

// 	cancel()
// }
