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
	switch k := v.Kind(); k {
	case reflect.Map:
		return unmarshalMap(v, b, buf)
	case reflect.Slice:
		return unmarshalSlice(v, b, buf)
	case reflect.Struct:
		return unmarshalStruct(v, b, buf)
	case reflect.Interface:
		var f reflect.Value
		switch tag := b >> 5; {
		case tag == Int>>5:
			f = reflect.ValueOf(new(int)).Elem()
		case tag == Uint>>5:
			f = reflect.ValueOf(new(uint)).Elem()
		case tag == String>>5 || tag == Bin>>5:
			f = reflect.ValueOf(new(string)).Elem()
		case b == True || b == False:
			f = reflect.ValueOf(new(bool)).Elem()
		case b == Float32:
			f = reflect.ValueOf(new(float32)).Elem()
		case b == Float64:
			f = reflect.ValueOf(new(float64)).Elem()
		case b == Tag:
		default:
			return InvalidTagErr(b >> 5)
		}
		if err := decode(f, b, buf); err != nil {
			return err
		}
		v.Set(f)
	case reflect.Ptr:
		buf.UnreadByte()
		return unmarshal(v.Elem(), buf)
	default:
		return decode(v, b, buf)
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
	length := int(b & maskTag)
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
	length := int(b & maskTag)
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
		switch length := b & maskTag; {
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
			size = int(length & maskTag)
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
	case reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uint:
		if tag := b >> 5; tag != Uint {
			return InvalidTagErr(tag)
		}
		var val uint64
		switch length := b & maskTag; {
		case length == Len1:
			if b, err := buf.ReadByte(); err != nil {
				return err
			} else {
				val = uint64(b)
			}
		case length == Len2:
			val = uint64(binary.BigEndian.Uint16(buf.Next(2)))
		case length == Len4:
			val = uint64(binary.BigEndian.Uint32(buf.Next(4)))
		case length == Len8:
			val = uint64(binary.BigEndian.Uint64(buf.Next(8)))
		default:
			val = uint64(length)
		}
		v.SetUint(val)
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
	switch length := b & maskTag; {
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
