package logging

import (
	"context"
)

func logWorker[T any](ctx context.Context, log Log[T], ch <-chan T, flushRate, retryCount int) {
	cache := make([]T, 0, flushRate)
	retryQueue := make([]T, 0, flushRate)

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
			retryWrite()
			return
		case data, more := <-ch:
			if !more {
				retryWrite()
				return
			}

			if len(cache) >= flushRate {
				retryWrite()
			} else {
				cache = append(cache, data)
			}
		}
	}
}
