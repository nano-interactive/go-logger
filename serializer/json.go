package serializer

import (
	"bytes"
	"encoding/json"
	"sync"
)

type (
	Json[T any] struct {
		jsonEncoders sync.Pool
	}

	jsonEncoder struct {
		enc *json.Encoder
		buf *bytes.Buffer
	}
)

var _ Interface[any] = &Json[any]{}

func NewJson[T any]() *Json[T] {
	return &Json[T]{
		jsonEncoders: sync.Pool{
			New: func() any {
				buf := bytes.NewBuffer(make([]byte, 0, defaultBufferSize))

				enc := json.NewEncoder(buf)

				return &jsonEncoder{
					buf: buf,
					enc: enc,
				}
			},
		},
	}
}

func (j *Json[T]) Serialize(data []T) ([]byte, error) {
	enc := j.jsonEncoders.Get().(*jsonEncoder)
	defer j.jsonEncoders.Put(enc)
	defer enc.buf.Reset()

	for _, v := range data {
		err := enc.enc.Encode(v)
		if err != nil {
			return nil, err
		}
	}

	return enc.buf.Bytes(), nil
}
