package logging

import (
	"context"
)

type (
	Config[T any] struct {
		logger     Error
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
