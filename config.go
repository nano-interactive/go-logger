package logging

import (
	"context"

	"github.com/nano-interactive/go-logger/serializer"
)

type (
	Config[T any] struct {
		logger     Error
		serializer serializer.Interface[T]
		ctx        context.Context
		delimiter  rune
	}

	Modifier[T any] func(*Config[T])
)

func WithLogger[T any](err Error) Modifier[T] {
	return func(c *Config[T]) {
		c.logger = err
	}
}

func WithSerializer[T any](serializer serializer.Interface[T]) Modifier[T] {
	return func(c *Config[T]) {
		c.serializer = serializer
	}
}

func WithContext[T any](ctx context.Context) Modifier[T] {
	return func(c *Config[T]) {
		c.ctx = ctx
	}
}

func WithDelimiter[T any](delimiter rune) Modifier[T] {
	return func(c *Config[T]) {
		c.delimiter = delimiter
	}
}
