package cbor

import (
	"bytes"
	"encoding/binary"
	"math"
	"net/url"
	"reflect"
	"regexp"
	"time"
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
	/*if b == Nil&mask || b == Undefined&mask {
		return nil
	}*/
	switch major, info := b>>5, b&mask; major {
	case Uint >> 5:
		return decodeUint(v, info, buf)
	case Int >> 5:
		return decodeInt(v, info, buf)
	case Bin >> 5, String >> 5:
		return decodeString(v, info, buf)
	case Slice >> 5:
		return unmarshalSlice(v, info, buf)
	case Map >> 5:
		return unmarshalMap(v, info, buf)
	case Tag >> 5:
		return decodeTag(v, info, buf)
	case Other >> 5:
		return decodeOther(v, info, buf)
	}
	return nil
}

func unmarshalMap(v reflect.Value, info byte, buf *bytes.Buffer) error {
	var length int
	switch info {
	case Len1:
		b, _ := buf.ReadByte()
		length = int(b)
	case Len2:
		length = int(binary.BigEndian.Uint16(buf.Next(2)))
	case Len4:
		length = int(binary.BigEndian.Uint32(buf.Next(4)))
	case Len8:
		length = int(binary.BigEndian.Uint64(buf.Next(8)))
	default:
		length = int(info)
	}
	switch k := v.Kind(); k {
	case reflect.Struct:
		for i := 0; i < v.NumField(); i++ {
			f := v.Field(i)
			if !f.CanSet() {
				continue
			}
			if err := unmarshal(f, buf); err != nil {
				return err
			}
		}
	case reflect.Map:
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
	default:
		return UnsupportedTypeErr(k)
	}
	return nil
}

func unmarshalSlice(v reflect.Value, info byte, buf *bytes.Buffer) error {
	if v.Kind() != reflect.Slice {
		return UnsupportedTypeErr(v.Kind())
	}
	var length int
	switch info {
	case Len1:
		b, _ := buf.ReadByte()
		length = int(b)
	case Len2:
		length = int(binary.BigEndian.Uint16(buf.Next(2)))
	case Len4:
		length = int(binary.BigEndian.Uint32(buf.Next(4)))
	case Len8:
		length = int(binary.BigEndian.Uint64(buf.Next(8)))
	default:
		length = int(info)
	}
	for i := 0; i < length; i++ {
		f := reflect.New(v.Type().Elem()).Elem()
		if err := unmarshal(f, buf); err != nil {
			return err
		}
		v.Set(reflect.Append(v, f))
	}
	return nil
}

func decodeInt(v reflect.Value, info byte, buf *bytes.Buffer) error {
	var value int64
	switch info {
	case Len1:
		b, err := buf.ReadByte()
		if err != nil {
			return err
		}
		value = int64(b)
	case Len2:
		value = int64(binary.BigEndian.Uint16(buf.Next(2)))
	case Len4:
		value = int64(binary.BigEndian.Uint32(buf.Next(4)))
	case Len8:
		value = int64(binary.BigEndian.Uint64(buf.Next(8)))
	default:
		value = int64(info)
	}

	var f reflect.Value
	switch k, value := v.Kind(), -1-value; k {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		f = reflect.ValueOf(value)
	case reflect.Interface:
		f = reflect.New(reflect.TypeOf(value)).Elem()
		f.SetInt(value)
	default:
		return UnsupportedTypeErr(k)
	}
	v.Set(f)

	return nil
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
		value = uint64(binary.BigEndian.Uint32(buf.Next(4)))
	case Len8:
		value = uint64(binary.BigEndian.Uint64(buf.Next(8)))
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

func decodeTag(v reflect.Value, info byte, buf *bytes.Buffer) error {
	if info >= Item {
		b, err := buf.ReadByte()
		if err != nil {
			return nil
		}
		info = b
	}
	switch info {
	case IsoTime:
		var str string
		f := reflect.ValueOf(&str).Elem()
		if err := unmarshal(f, buf); err != nil {
			return err
		}
		t, err := time.Parse(time.RFC3339, str)
		if err != nil {
			return err
		}
		v.Set(reflect.ValueOf(t))
	case URI:
		var str string
		f := reflect.ValueOf(&str).Elem()
		if err := unmarshal(f, buf); err != nil {
			return err
		}
		u, err := url.Parse(str)
		if err != nil {
			return err
		}
		v.Set(reflect.ValueOf(*u))
	case Regex:
		var str string
		f := reflect.ValueOf(&str).Elem()
		if err := unmarshal(f, buf); err != nil {
			return err
		}
		r, err := regexp.Compile(str)
		if err != nil {
			return err
		}
		v.Set(reflect.ValueOf(*r))
	default:
		return UnassignedTagErr(info)
	}
	return nil
}

func decodeOther(v reflect.Value, info byte, buf *bytes.Buffer) error {
	var value interface{}
	switch info {
	case Nil, Undefined:
		return nil
	case True:
		value = true
	case False:
		value = false
	case Float32:
		var u uint32
		if err := binary.Read(buf, binary.BigEndian, &u); err != nil {
			return err
		}
		value = math.Float32frombits(u)
	case Float64:
		var u uint64
		if err := binary.Read(buf, binary.BigEndian, &u); err != nil {
			return err
		}
		value = math.Float64frombits(u)
	default:
		return UnassignedTagErr(info)
	}
	var f reflect.Value
	switch k := v.Kind(); {
	case k == v.Kind():
		f = reflect.ValueOf(value)
	case k == reflect.Interface:
		f = reflect.New(reflect.TypeOf(value)).Elem()
		f.Set(reflect.ValueOf(value))
	default:
		return UnsupportedTypeErr(k)
	}
	v.Set(f)

	return nil
}

func decodeString(v reflect.Value, info byte, buf *bytes.Buffer) error {
	var size int
	switch info {
	case Len1:
		b, err := buf.ReadByte()
		if err != nil {
			return err
		}
		size = int(b)
	case Len2:
		size = int(binary.BigEndian.Uint16(buf.Next(2)))
	case Len4:
		size = int(binary.BigEndian.Uint32(buf.Next(4)))
	case Len8:
		size = int(binary.BigEndian.Uint64(buf.Next(8)))
	default:
		size = int(info)
	}
	value := string(buf.Next(size))

	var f reflect.Value
	switch k := v.Kind(); k {
	case reflect.String:
		f = reflect.ValueOf(value)
	case reflect.Interface:
		f = reflect.New(reflect.TypeOf(value)).Elem()
		f.SetString(value)
	default:
		return UnsupportedTypeErr(k)
	}
	v.Set(f)

	return nil
}
