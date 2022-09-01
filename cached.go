package logging

import (
	"context"
	"io"
	"sync/atomic"
)

var (
	_ io.Closer = &CachedLogging[any]{}
	_ Log[any]  = &CachedLogging[any]{}
)

type CachedLogging[T any] struct {
	cancel context.CancelFunc
	logger Log[T]
	chs    []chan T
	chsLen uint64
	idx    uint64
}

func NewCached[T any](log Log[T], mods ...ModifierCached) *CachedLogging[T] {
	config := defaultCachedConfig
	ctx, cancel := context.WithCancel(context.Background())

	for _, mod := range mods {
		mod(&config)
	}

	if config.retryCount <= 0 {
		config.retryCount = 1
	}

	chs := make([]chan T, config.workers)
	cancelFns := make([]context.CancelFunc, 0, config.workers)

	for i := 0; i < config.workers; i++ {
		chs[i] = make(chan T, config.bufferSize)
		workerCtx, cancel := context.WithCancel(context.Background())
		cancelFns = append(cancelFns, cancel)
		go logWorker(workerCtx, log, chs[i], config.flushRate, config.retryCount)
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
		chsLen: uint64(len(chs)),
		cancel: cancel,
	}
}

func (l *CachedLogging[T]) Log(log T) error {
	idx := atomic.AddUint64(&l.idx, 1) % l.chsLen

	l.chs[idx] <- log
	return nil
}

func (l *CachedLogging[T]) LogMultiple(logs []T) error {
	length := uint64(len(logs))

	for i := uint64(0); i < length; i++ {
		// Evenly distribute the logs to the workers
		l.chs[i%l.chsLen] <- logs[i]
	}

	return nil
}

func (l *CachedLogging[T]) Close() error {
	c := l.cancel
	c()

	for _, ch := range l.chs {
		close(ch)
	}

	if logger, ok := l.logger.(io.Closer); ok {
		return logger.Close()
	}

	return nil
}
