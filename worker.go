package logger

import (
	"context"
	"sync"
)

func logWorker[T any](ctx context.Context, wg *sync.WaitGroup, log Log[T], ch <-chan T, bufferSize, retryCount int) {
	cache := make([]T, 0, bufferSize)
	defer wg.Done()

	retryPool := &sync.Pool{
		New: func() interface{} {
			return make([]T, 0, bufferSize)
		},
	}

	retryWrite := func() {
		var err error

		if err = log.LogMultiple(cache); err == nil {
			cache = cache[:0]
			return
		}

		retryQueue := retryPool.Get().([]T)
		copy(retryQueue, cache)

		go func(retryQueue []T) {
			defer func() {
				retryQueue = retryQueue[:0]
				retryPool.Put(retryQueue)
			}()
			for i := 0; i < retryCount; i++ {
				if err = log.LogMultiple(retryQueue); err == nil {
					return
				}
			}
		}(retryQueue)

		cache = cache[:0]
	}

	for {
		select {
		case <-ctx.Done():
			goto flush
		case data, more := <-ch:
			var appended bool
			if len(cache) < bufferSize {
				cache = append(cache, data)
				appended = true
			}

			// Recalculate length again so we can flush the cache
			if len(cache) >= bufferSize {
				retryWrite()
			}

			if !appended {
				cache = append(cache, data)
			}

			if !more {
				goto flush
			}
		}
	}
flush:
	//Empty the buffered channel
	for data := range ch {
		cache = append(cache, data)
	}

	retryWrite()
}
