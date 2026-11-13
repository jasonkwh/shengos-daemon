package message

import (
	"github.com/vmihailenco/msgpack/v5"
)

func serialize(v any) ([]byte, error) {
	return msgpack.Marshal(v)
}

func deserialize(data []byte, v any) error {
	return msgpack.Unmarshal(data, v)
}
