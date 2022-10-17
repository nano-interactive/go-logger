package logger

import (
	"context"
	"sync"
)

func logWorker[T any](ctx context.Context, wg *sync.WaitGroup, log Log[T], ch <-chan T, bufferSize, retryCount int) {
	cache := make([]T, 0, bufferSize)
	retryQueue := make([]T, 0, bufferSize)
	defer wg.Done()

	retryWrite := func() {
		var err error

		hasCopied := false
		dataToWrite := cache

		for i := 0; i < retryCount; i++ {
			if err = log.LogMultiple(dataToWrite); err == nil {
				cache = cache[:0]
				break
			}

			if !hasCopied {
				copy(retryQueue, cache)
				dataToWrite = retryQueue
				hasCopied = true
			}

			if i == 0 {
				cache = cache[:0]
			}
		}

		retryQueue = retryQueue[:0]
	}

	for {
		select {
		case <-ctx.Done():
			goto flush
		case data, more := <-ch:
			if len(cache) >= bufferSize {
				retryWrite()
			}

			cache = append(cache, data)

			if !more {
				goto flush
			}
		}
	}
flush:
	for data := range ch {
		cache = append(cache, data)
	}

	retryWrite()
}
