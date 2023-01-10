package logger

import (
	"context"
	"sync"
)

var (
	retryPool     sync.Pool
	retryPoolOnce sync.Once
)

func getRetryPool[T any]() *sync.Pool {
	retryPoolOnce.Do(func() {
		retryPool = sync.Pool{
			New: func() any {
				return make([]T, 0)
			},
		}
	})

	return &retryPool
}

func logWorker[T any](ctx context.Context, wg *sync.WaitGroup, log Log[T], ch <-chan T, bufferSize, retryCount int) {
	cache := make([]T, bufferSize, bufferSize)
	idx := 0
	defer wg.Done()

	reset := func() {
		idx = 0
	}

	set := func(data T) {
		cache[idx] = data
		idx++
	}

	retryWrite := func() {
		var err error

		defer reset()

		if err = log.LogMultiple(cache); err == nil {
			return
		}

		retryQueue := getRetryPool[T]().Get().([]T)
		copy(retryQueue, cache) // size of the retryQueue is always equal to the size of the cache

		go func(retryQueue []T) {
			defer func() {
				retryQueue = retryQueue[:0]
				getRetryPool[T]().Put(retryQueue)
			}()

			for i := 0; i < retryCount; i++ {
				if err = log.LogMultiple(retryQueue); err == nil {
					return
				}
			}
		}(retryQueue)
	}

	for {
		select {
		case <-ctx.Done():
			goto flush
		case data, more := <-ch:
			var appended bool

			if idx < bufferSize-1 {
				set(data)
				appended = true
			}

			// Recalculate length again so we can flush the cache
			if idx == bufferSize-1 {
				retryWrite()
				reset()
			}

			if !appended {
				set(data)
			}

			if !more {
				goto flush
			}
		}
	}
flush:
	// Empty the buffered channel
	cache = cache[:idx]
	for data := range ch {
		cache = append(cache, data)
	}

	retryWrite()
}
