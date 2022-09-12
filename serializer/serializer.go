package serializer

const defaultBufferSize = 8192

type Interface[T any] interface {
	Serialize([]T) ([]byte, error)
}
