package logging

type (
	Config[T any] struct {
		logger    Error
		delimiter rune
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
	flushRate:  1,
	retryCount: 1,
}


func WithErrorLogger[T any](err Error) Modifier[T] {
	return func(c *Config[T]) {
		c.logger = err
	}
}

func WithDelimiter[T any](delimiter rune) Modifier[T] {
	return func(c *Config[T]) {
		c.delimiter = delimiter
	}
}

func WithBufferSize(size int) ModifierCached {
	return func(c *CachedLoggingConfig) {
		c.bufferSize = size
	}
}

func WithWorkerPool(size int) ModifierCached {
	return func(c *CachedLoggingConfig) {
		c.bufferSize = size
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
