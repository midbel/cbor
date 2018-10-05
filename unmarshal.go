package cbor

import (
	"bufio"
	"bytes"
	"reflect"
)

func Unmarshal(bs []byte, v interface{}) error {
	r := bytes.NewReader(bs)
	return unmarshal(bufio.NewReader(r), reflect.ValueOf(v).Elem())
}

func unmarshal(r *bufio.Reader, v reflect.Value) error {
	b, err := r.ReadByte()
	if err != nil {
		return err
	}
	switch m, a, k := b&0xE0, b&0x1F, v.Kind(); m {
	case Uint:
	case Int:
	case Bin:
	case String:
	case Array:
		if k == reflect.Slice || k == reflect.Array {
			return nil
		}
		return fmt.Errorf("expected array like value: got %s", k)
	case Map:
		if k == reflect.Map {
			return nil
		}
		if k == reflect.Struct {
			return nil
		}
		return fmt.Errorf("expected map like value: got %s", k)
	case Tag:
	case Other:
	}
	return nil
}

func isInt(k reflect.Kind) bool {
	return k == reflect.Int || k == reflect.Int8 || k == reflect.Int16 || k == reflect.Int32 || k == reflect.Int64
}

func isUint(k reflect.Kind) bool {
	return k == reflect.Uint || k == reflect.Uint8 || k == reflect.Uint16 || k == reflect.Uint32 || k == reflect.Uint64
}

func isFloat(k reflect.Kind) bool {
	return k == reflect.Float32 || k == reflect.Float64
}
