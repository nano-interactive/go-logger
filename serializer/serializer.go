package serializer

import "bytes"

const defaultBufferSize = 8192 * 2

type (
	Interface[T any] interface {
		Serialize([]T) ([]byte, error)
	}
	PooledSerializer[T any] interface {
		Interface[T]
		getBuffer() *bytes.Buffer
	}

	PoolInterface[T any, TSerializer PooledSerializer[T]] interface {
		Acquire() TSerializer
		Release(TSerializer)
	}
)
