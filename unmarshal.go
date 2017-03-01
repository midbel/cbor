package cbor

import (
	"bytes"
	"encoding/binary"
	"math"
	"reflect"
)

func Unmarshal(data []byte, d interface{}) ([]byte, error) {
	return runUnmarshal(data, d)
}

func runUnmarshal(data []byte, d interface{}) ([]byte, error) {
	buf := bytes.NewBuffer(data)
	if err := unmarshal(reflect.ValueOf(d).Elem(), buf); err != nil {
		return buf.Bytes(), err
	}
	return buf.Bytes(), nil
}

func unmarshal(v reflect.Value, buf *bytes.Buffer) error {
	b, err := buf.ReadByte()
	if err != nil {
		return err
	}
	if b == Nil || b == Undefined {
		return nil
	}
	switch major, info := b>>5, b&mask; major {
	case Uint>>5:
		return decodeUint(v, info, buf)
	case Int>>5:
	case Bin>>5, String>>5:
	case Slice>>5:
	case Map>>5:
	case Tag>>5:
	case Other>>5:
	}
	return nil
}

func unmarshalStruct(v reflect.Value, b byte, buf *bytes.Buffer) error {
	if tag := b >> 5; tag != Map>>5 {
		return InvalidTagErr(tag)
	}
	for i := 0; i < v.NumField(); i++ {
		f := v.Field(i)
		if !f.CanSet() {
			continue
		}
		if err := unmarshal(f, buf); err != nil {
			return err
		}
	}
	return nil
}

func unmarshalSlice(v reflect.Value, b byte, buf *bytes.Buffer) error {
	if tag := b >> 5; tag != Slice>>5 {
		return InvalidTagErr(tag)
	}
	length := int(b & mask)
	for i := 0; i < length; i++ {
		f := reflect.New(v.Type().Elem()).Elem()
		if err := unmarshal(f, buf); err != nil {
			return err
		}
		v.Set(reflect.Append(v, f))
	}
	return nil
}

func unmarshalMap(v reflect.Value, b byte, buf *bytes.Buffer) error {
	if tag := b >> 5; tag != Map>>5 {
		return InvalidTagErr(tag)
	}
	length := int(b & mask)
	for i := 0; i < length; i++ {
		key := reflect.New(v.Type().Key()).Elem()
		if err := unmarshal(key, buf); err != nil {
			return err
		}
		value := reflect.New(v.Type().Elem()).Elem()
		if err := unmarshal(value, buf); err != nil {
			return err
		}
		v.SetMapIndex(key, value)
	}
	return nil
}

func decode(v reflect.Value, b byte, buf *bytes.Buffer) error {
	switch k := v.Kind(); k {
	case reflect.String:
		if b := b >> 5; !(b == Bin>>5 || b == String>>5) {
			return InvalidTagErr(b)
		}
		var size int
		switch length := b & mask; {
		case length == Len1:
			if b, err := buf.ReadByte(); err != nil {
				return err
			} else {
				size = int(b)
			}
		case length == Len2:
			size = int(binary.BigEndian.Uint16(buf.Next(2)))
		case length == Len4:
			size = int(binary.BigEndian.Uint32(buf.Next(4)))
		case length == Len8:
			size = int(binary.BigEndian.Uint64(buf.Next(8)))
		default:
			size = int(length & mask)
		}
		v.SetString(string(buf.Next(size)))
	case reflect.Bool:
		switch b {
		case True:
			v.SetBool(true)
		case False:
			v.SetBool(false)
		default:
			return InvalidTagErr(b)
		}
	case reflect.Float32:
		if b != Float32 {
			return InvalidTagErr(b)
		}
		var f uint32
		if err := binary.Read(buf, binary.BigEndian, &f); err != nil {
			return err
		}
		v.SetFloat(float64(math.Float32frombits(f)))
	case reflect.Float64:
		if b != Float64 {
			return InvalidTagErr(b)
		}
		var f uint64
		if err := binary.Read(buf, binary.BigEndian, &f); err != nil {
			return err
		}
		v.SetFloat(math.Float64frombits(f))
	case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Int:
		i, err := decodeInt(b, buf)
		if err != nil {
			return err
		}
		v.SetInt(i)
	default:
		return UnsupportedTypeErr(k)
	}
	return nil
}

func decodeInt(b byte, buf *bytes.Buffer) (int64, error) {
	tag := b >> 5
	if tag == Uint {
		var result uint64
		val := reflect.ValueOf(&result).Elem()
		if err := decode(val, b, buf); err != nil {
			return 0, err
		}
		return int64(val.Uint()), nil
	}
	var val int64
	switch length := b & mask; {
	case length == Len1:
		if b, err := buf.ReadByte(); err != nil {
			return 0, err
		} else {
			val = int64(b)
		}
	case length == Len2:
		val = int64(binary.BigEndian.Uint16(buf.Next(2)))
	case length == Len4:
		val = int64(binary.BigEndian.Uint32(buf.Next(4)))
	case length == Len8:
		val = int64(binary.BigEndian.Uint64(buf.Next(8)))
	default:
		val = int64(length)
	}
	return -1 - val, nil
}

func decodeUint(v reflect.Value, info byte, buf *bytes.Buffer) error {
	var value uint64
	switch info {
	case Len1:
		b, err := buf.ReadByte()
		if err != nil {
			return err
		}
		value = uint64(b)
	case Len2:
		value = uint64(binary.BigEndian.Uint16(buf.Next(2)))
	case Len4:
		value = uint64(binary.BigEndian.Uint16(buf.Next(4)))
	case Len8:
		value = uint64(binary.BigEndian.Uint16(buf.Next(8)))
	default:
		value = uint64(info)
	}
	var f reflect.Value
	switch k := v.Kind(); k {
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		f = reflect.ValueOf(value)
	case reflect.Interface:
		f = reflect.New(reflect.TypeOf(value)).Elem()
		f.SetUint(value)
	default:
		return UnsupportedTypeErr(k)
	}
	v.Set(f)

	return nil
}
