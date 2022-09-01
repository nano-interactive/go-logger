package logging

import (
	"context"
	"io"
	"sync/atomic"
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

func logWorker[T any](ctx context.Context, log Log[T], ch <-chan T, flushRate, retryCount int) {
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
			}
		}

		cache = cache[:0]
		retryQueue = retryQueue[:0]
	}

	for {
		select {
		case <-ctx.Done():
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

var DefaultCachedConfig = CachedLoggingConfig{
	Workers:    1,
	BufferSize: 1024,
	FlushRate:  1,
	RetryCount: 1,
}

func NewCached[T any](ctx context.Context, log Log[T], cfg ...CachedLoggingConfig) *CachedLogging[T] {
	config := DefaultCachedConfig

	if len(cfg) > 0 {
		fromArgs := cfg[0]

		if fromArgs.Workers > 1 {
			config.Workers = fromArgs.Workers
		}

		if fromArgs.BufferSize > 0 {
			config.BufferSize = fromArgs.BufferSize
		}

		if fromArgs.FlushRate > 0 {
			config.FlushRate = fromArgs.FlushRate
		}
		if fromArgs.RetryCount > 1 {
			config.RetryCount = fromArgs.RetryCount
		}
	}

	chs := make([]chan T, config.Workers)
	cancelFns := make([]context.CancelFunc, 0, config.Workers)

	for i := 0; i < config.Workers; i++ {
		chs[i] = make(chan T, config.BufferSize)
		workerCtx, cancel := context.WithCancel(context.Background())
		cancelFns = append(cancelFns, cancel)
		go logWorker(workerCtx, log, chs[i], config.FlushRate, config.RetryCount)
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
