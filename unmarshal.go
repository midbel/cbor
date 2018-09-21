package cbor

import (
  "bytes"
  "reflect"
)

func Unmarshal(bs []byte, v interface{}) error {
	return unmarshal(bytes.NewReader(bs), reflect.ValueOf(v).Elem())
}

func unmarshal(b *bytes.Reader, v reflect.Value) error {
	return nil
}
