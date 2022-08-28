package logging

import (
	"context"
	"io"
	"sync/atomic"

	"github.com/rs/zerolog"
)

var _ io.Closer = &CachedLogging[any]{}

type CachedLoggingConfig struct {
	Workers    int `json:"workers" mapstructure:"workers"`
	BufferSize int `json:"buffer_size" mapstructure:"buffer_size"`
	FlushRate  int `json:"flush_rate" mapstructure:"flush_rate"`
	RetryCount int `json:"retry_count" mapstructure:"retry_count"`
}

type CachedLogging[T any] struct {
	logger Log[T]
	chs    []chan T
	idx    uint64
}

func logWorker[T any](ctx context.Context, logger zerolog.Logger, log Log[T], ch <-chan T, flushRate, retryCount int) {
	cache := make([]T, 0, flushRate)
	retryQueue := make([]T, 0, flushRate)

	retryWrite := func() {
		var err error
		count := len(cache)
		cache = cache[:0]

		dataToWrite := cache
		hasCopied := false
		for i := 0; i < retryCount || err == nil; i++ {
			if count > 0 {
				err = log.LogMultiple(dataToWrite)
				if err == nil {
					break
				}

				if !hasCopied {
					copy(retryQueue, cache)
					cache = cache[:0]
					dataToWrite = retryQueue
					hasCopied = true
				}

				logger.Error().
					Err(err).
					Int("retry_count", i).
					Int("max_retry_count", retryCount).
					Int("flushRate", flushRate).
					Bool("is_dropping_payload", i == retryCount-1).
					Msg("Failed to log data")
			}
		}

		cache = cache[:0]
		retryQueue = retryQueue[:0]
	}

	for {
		select {
		case <-ctx.Done():
			logger.Info().Int("logs_count", len(cache)).Msg("Flushing rest of the cached logs to file")

			retryWrite()

			return
		case data := <-ch:
			if len(cache) >= flushRate {
				retryWrite()
			} else {
				cache = append(cache, data)
			}
		}
	}
}

func NewCached[T any](ctx context.Context, logger zerolog.Logger, log Log[T], config CachedLoggingConfig) *CachedLogging[T] {
	chs := make([]chan T, config.Workers)
	cancelFns := make([]context.CancelFunc, 0, config.Workers)

	for i := 0; i < config.Workers; i++ {
		chs[i] = make(chan T, config.BufferSize)
		workerCtx, cancel := context.WithCancel(context.Background())
		cancelFns = append(cancelFns, cancel)
		go logWorker(workerCtx, logger, log, chs[i], config.FlushRate, config.RetryCount)
	}

	go func(ctx context.Context) {
		<-ctx.Done()

		for _, cancel := range cancelFns {
			cancel()
		}
	}(ctx)

	return &CachedLogging[T]{
		logger: log,
		chs:    chs,
	}
}

func (l *CachedLogging[T]) Log(log T) error {
	idx := atomic.AddUint64(&l.idx, 1) % uint64(len(l.chs))

	l.chs[idx] <- log
	return nil
}

func (l *CachedLogging[T]) LogMultiple(logs []T) error {
	for i, item := range logs {
		// Evenly distribute the logs to the workers
		l.chs[i%len(l.chs)] <- item
	}

	return nil
}

func (l *CachedLogging[T]) Close() error {
	for _, ch := range l.chs {
		close(ch)
	}

	if logger, ok := l.logger.(io.Closer); ok {
		return logger.Close()
	}

	return nil
}
