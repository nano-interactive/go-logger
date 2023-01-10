package serializer

import (
	"bytes"
	"encoding/json"
	"sync"
)

type (
	Json[T any] struct{}

	PoolJson[T any] struct {
		pool sync.Pool
	}

	PoolJsonSerializer[T any] struct {
		buf *bytes.Buffer
	}
)

var (
	_ Interface[any]        = &Json[any]{}
	_ PooledSerializer[any] = &PoolJsonSerializer[any]{}
)

func NewJson[T any]() *Json[T] {
	return &Json[T]{}
}

func (j *Json[T]) Serialize(data []T) ([]byte, error) {
	buf := bytes.NewBuffer(nil)
	buf.Grow(defaultBufferSize)
	enc := json.NewEncoder(buf)

	for _, v := range data {
		err := enc.Encode(v)
		if err != nil {
			return nil, err
		}
	}

	return buf.Bytes(), nil
}

func NewPoolJson[T any](buff *bytes.Buffer) *PoolJsonSerializer[T] {
	return &PoolJsonSerializer[T]{
		buf: buff,
	}
}

func (j *PoolJsonSerializer[T]) Serialize(data []T) ([]byte, error) {
	enc := json.NewEncoder(j.buf)
	enc.SetEscapeHTML(false)

	for _, v := range data {
		err := enc.Encode(v)
		if err != nil {
			return nil, err
		}
	}

	return j.buf.Bytes(), nil
}

func (j *PoolJsonSerializer[T]) getBuffer() *bytes.Buffer {
	return j.buf
}
