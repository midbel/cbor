package cbor

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
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
		err = unmarshalUint(r, a, v)
	case Int:
		err = unmarshalInt(r, a, v)
	case Bin:
	case String:
		err = unmarshalString(r, a, v)
	case Array:
		err = unmarshalArray(r, a, v)
	case Map:
		if k == reflect.Map {
			return nil
		}
		if k == reflect.Struct || k == reflect.Ptr {
			return unmarshalStruct(r, a, v)
		}
		return expectedType("map/struct", k)
	case Tag:
	case Other:
	}
	return err
}

// var (
// 	timeType = nil
// 	urlType = nil
// )

func unmarshalStruct(r io.Reader, a byte, v reflect.Value) error {
	size, err := sizeof(r, a)
	if err != nil {
		return err
	}
	for i := 0; i < size; i++ {

	}
	return nil
}

func unmarshalArray(r io.Reader, a byte, v reflect.Value) error {
	if k := v.Kind(); !(k == reflect.Array || k == reflect.Slice) {
		return expectedType("array/slice", k)
	}
	size, err := sizeof(r, a)
	if err != nil {
		return err
	}
	for i := 0; i < size; i++ {

	}
	return nil
}

func unmarshalString(r io.Reader, a byte, v reflect.Value) error {
	if k := v.Kind(); k != reflect.String {
		return expectedType("string", k)
	}
	size, err := sizeof(r, a)
	if err != nil {
		return err
	}
	bs := make([]byte, size)
	if _, err := io.ReadFull(r, bs); err != nil {
		return err
	}
	v.SetBytes(bs)
	return nil
}

func unmarshalInt(r io.Reader, a byte, v reflect.Value) error {
	var (
		i   int64
		err error
	)
	switch a {
	case Len1:
		var v int8
		err = binary.Read(r, binary.BigEndian, &v)
		i = int64(v)
	case Len2:
		var v int16
		err = binary.Read(r, binary.BigEndian, &v)
		i = int64(v)
	case Len4:
		var v int32
		err = binary.Read(r, binary.BigEndian, &v)
		i = int64(v)
	case Len8:
		var v int64
		err = binary.Read(r, binary.BigEndian, &v)
		i = int64(v)
	}
	if err != nil {
		return err
	}
	if k := v.Kind(); !isInt(k) {
		return expectedType("int", k)
	}
	v.SetInt(-1 - int64(i))
	return nil
}

func unmarshalUint(r io.Reader, a byte, v reflect.Value) error {
	var (
		i   uint64
		err error
	)
	switch a {
	default:
		i = uint64(a)
	case Len1:
		var v uint8
		err = binary.Read(r, binary.BigEndian, &v)
		i = uint64(v)
	case Len2:
		var v uint16
		err = binary.Read(r, binary.BigEndian, &v)
		i = uint64(v)
	case Len4:
		var v uint32
		err = binary.Read(r, binary.BigEndian, &v)
		i = uint64(v)
	case Len8:
		var v uint64
		err = binary.Read(r, binary.BigEndian, &v)
		i = uint64(v)
	}
	if err != nil {
		return err
	}
	switch k := v.Kind(); {
	case isUint(k):
		v.SetUint(i)
	case isInt(k):
		v.SetInt(int64(i))
	default:
		return expectedType("uint/int", k)
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

func expectedType(n string, k reflect.Kind) error {
	return fmt.Errorf("expected %s, got %s", n, k)
}
