package cbor

import (
	// "bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"reflect"
	"strings"
)

type reader interface {
	io.ByteReader
	io.Reader
}

func Unmarshal(bs []byte, v interface{}) error {
	r := bytes.NewReader(bs)
	return unmarshal(r, reflect.ValueOf(v).Elem())
}

func unmarshal(r reader, v reflect.Value) error {
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
			return unmarshalMap(r, a, v)
		}
		if k == reflect.Struct || k == reflect.Ptr {
			return unmarshalStruct(r, a, v)
		}
		return expectedType("map/struct", k)
	case Other:
		return unmarshalSimple(r, a, v)
	case Tag:
		return unmarshalTagged(r, a, v)
	}
	return err
}

func unmarshalTagged(r reader, a byte, v reflect.Value) error {
	switch a {
	default:
	case Len1:
		b, err := r.ReadByte()
		if err != nil {
			return err
		}
		a = b
	case Len2:
	case Len4:
	case Len8:
	}
	switch a {
	default:
		return fmt.Errorf("unsupported tagged item %02x", a)
	case TagURI, TagRFC3339, TagUnix:
		return unmarshal(r, v)
	}
	return nil
}

func unmarshalSimple(r reader, a byte, v reflect.Value) error {
	switch k := v.Kind(); a {
	default:
		if isInt(k) {
			v.SetInt(int64(a))
		} else if isUint(k) {
			v.SetUint(uint64(a))
		} else {
			return expectedType("int/uint", k)
		}
	case Simple:
		b, err := r.ReadByte()
		if err != nil {
			return err
		}
		if isInt(k) {
			v.SetInt(int64(b))
		} else if isUint(k) {
			v.SetUint(uint64(b))
		} else {
			return expectedType("int/uint", k)
		}
	case False:
		if k != reflect.Bool {
			return expectedType("bool", k)
		}
		v.SetBool(false)
	case True:
		if k != reflect.Bool {
			return expectedType("bool", k)
		}
		v.SetBool(true)
	case Nil, Undefined:
	case Float16:
		return fmt.Errorf("float16 not yet supported")
	case Float32:
		if k == reflect.Float32 || k == reflect.Float64 {
			var f float32
			if err := binary.Read(r, binary.BigEndian, &f); err != nil {
				return err
			}
			v.SetFloat(float64(f))
		} else {
			return expectedType("float32", k)
		}
	case Float64:
		if k == reflect.Float64 {
			var f float64
			if err := binary.Read(r, binary.BigEndian, &f); err != nil {
				return err
			}
			v.SetFloat(f)
		} else {
			return expectedType("float64", k)
		}
	}
	return nil
}

func unmarshalMap(r reader, a byte, v reflect.Value) error {
	size, err := sizeof(r, a)
	if err != nil {
		return err
	}
	if v.IsNil() {
		v.Set(reflect.MakeMapWithSize(v.Type(), size))
	}
	seen := make(map[string]struct{})
	for i := 0; i < size; i++ {
		k := reflect.New(v.Type().Key()).Elem()
		if err := unmarshal(r, k); err != nil {
			return err
		}
		if _, ok := seen[k.String()]; ok {
			return fmt.Errorf("duplicate field %s", k)
		}
		seen[k.String()] = struct{}{}

		f := reflect.New(v.Type().Elem()).Elem()
		if err := unmarshal(r, f); err != nil {
			return err
		}
		v.SetMapIndex(k, f)
	}
	return nil
}

func unmarshalStruct(r reader, a byte, v reflect.Value) error {
	size, err := sizeof(r, a)
	if err != nil {
		return err
	}
	vs := make(map[string]reflect.Value)
	for i, t := 0, v.Type(); i < v.NumField(); i++ {
		f := v.Field(i)
		if !f.CanSet() {
			continue
		}
		j := t.Field(i)
		switch n := j.Tag.Get("cbor"); strings.ToLower(n) {
		case "-":
			continue
		case "":
			vs[j.Name] = f
		default:
			vs[n] = f
		}
	}
	seen := make(map[string]struct{})
	for i := 0; i < size; i++ {
		var k string
		f := reflect.New(reflect.TypeOf(k)).Elem()
		if err := unmarshal(r, f); err != nil {
			return err
		}
		k = f.String()
		if _, ok := seen[k]; ok {
			return fmt.Errorf("duplicate field %s", k)
		}
		seen[k] = struct{}{}

		f, ok := vs[k]
		if !ok {
			return fmt.Errorf("field not found %s", k)
		}
		if err := unmarshal(r, f); err != nil {
			return err
		}
	}
	return nil
}

func unmarshalArray(r reader, a byte, v reflect.Value) error {
	if k := v.Kind(); !(k == reflect.Array || k == reflect.Slice) {
		return expectedType("array/slice", k)
	}
	size, err := sizeof(r, a)
	if err != nil {
		return err
	}
	if k := v.Kind(); k == reflect.Array && size >= v.Len() {
		return fmt.Errorf("array length too short (got: %d, want: %d)", v.Len(), size)
	}
	if v.IsNil() {
		v.Set(reflect.MakeSlice(v.Type(), size, size))
	}
	for i := 0; i < size; i++ {
		var f reflect.Value
		if i < v.Len() {
			f = v.Index(i)
		} else {
			f = reflect.New(v.Type().Elem()).Elem()
		}
		if err := unmarshal(r, f); err != nil {
			return err
		}
		if i >= v.Len() {
			v.Set(reflect.Append(v, f))
		}
	}
	return nil
}

func unmarshalString(r reader, a byte, v reflect.Value) error {
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
	v.SetString(string(bs))
	// v.SetBytes(bs)
	return nil
}

func unmarshalInt(r reader, a byte, v reflect.Value) error {
	var (
		i   int64
		err error
	)
	switch a {
	default:
		i = int64(a)
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

func unmarshalUint(r reader, a byte, v reflect.Value) error {
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
