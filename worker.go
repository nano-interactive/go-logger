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
		count := len(cache)

		dataToWrite := cache
		hasCopied := false
		for i := 0; i < retryCount || err == nil; i++ {
			if count == 0 {
				break
			}

			if err = log.LogMultiple(dataToWrite); err == nil {
				break
			}

			if !hasCopied {
				copy(retryQueue, cache)
				dataToWrite = retryQueue
				hasCopied = true
			}
		}

		cache = cache[:0]
		retryQueue = retryQueue[:0]
	}

	for {
		select {
		case <-ctx.Done():
			goto flush
		case data, more := <-ch:
			if !more {
				goto flush
			}

			if len(cache) >= bufferSize {
				retryWrite()
			} else {
				cache = append(cache, data)
			}
		}
	}
flush:
	retryWrite()
}
