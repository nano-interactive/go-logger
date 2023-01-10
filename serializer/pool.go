package serializer

import (
	"bytes"
	"sync"
)

type (
	PooledSerializerCreator[T any, TSerializer PooledSerializer[T]] func(*bytes.Buffer) *TSerializer

	Pool[T any, TSerializer PooledSerializer[T]] struct {
		pool sync.Pool
	}
)

var (
	_ PoolInterface[any, *PoolJsonSerializer[any]] = &Pool[any, *PoolJsonSerializer[any]]{}
)

func NewPool[T any, TSerializer PooledSerializer[T]](create PooledSerializerCreator[T, TSerializer]) *Pool[T, TSerializer] {
	return &Pool[T, TSerializer]{
		pool: sync.Pool{
			New: func() any {
				buf := bytes.NewBuffer(nil)
				buf.Grow(defaultBufferSize)
				return create(buf)
			},
		},
	}
}

func (j *Pool[T, TSerializer]) Acquire() TSerializer {
	return j.pool.Get().(TSerializer)
}

func (j *Pool[T, TSerializer]) Release(s TSerializer) {
	s.getBuffer().Reset()
	j.pool.Put(s)
}
