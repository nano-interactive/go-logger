package serializer

const defaultBufferSize = 8192

type Interface[T any] interface {
	Serialize(v T) ([]byte, error)
	SerializeMultipleWithDelimiter(data []T, delimiter rune) ([]byte, error)
}
