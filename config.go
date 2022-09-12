package logger

type (
	Config[T any] struct {
		logger Error
	}

	CachedLoggingConfig struct {
		bufferSize int
		workers    int
		flushRate  int
		retryCount int
	}

	Modifier[T any] func(*Config[T])

	ModifierCached func(*CachedLoggingConfig)
)

var defaultCachedConfig = CachedLoggingConfig{
	workers:    1,
	bufferSize: 1024,
	retryCount: 1,
}

func WithErrorLogger[T any](err Error) Modifier[T] {
	return func(c *Config[T]) {
		c.logger = err
	}
}

func WithBufferSize(size int) ModifierCached {
	return func(c *CachedLoggingConfig) {
		c.bufferSize = size
	}
}

func WithWorkerPool(size int) ModifierCached {
	return func(c *CachedLoggingConfig) {
		c.workers = size
	}
}

func WithFlushRate(rate int) ModifierCached {
	return func(c *CachedLoggingConfig) {
		c.flushRate = rate
	}
}

func WithRetryCount(count int) ModifierCached {
	return func(c *CachedLoggingConfig) {
		c.retryCount = count
	}
}
