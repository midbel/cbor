package cbor

import (
	"bytes"
	"encoding/binary"
	"math"
	"net/url"
	"reflect"
	"time"
	"unicode/utf8"
)

func Marshal(v interface{}) ([]byte, error) {
	switch v := v.(type) {
	case time.Time:
		buf, err := runMarshal(v.UTC().Format(time.RFC3339))
		if err != nil {
			return nil, err
		}
		return append([]byte{Tag | IsoTime}, buf...), nil
	case url.URL:
		buf, err := runMarshal(v.String())
		if err != nil {
			return nil, err
		}
		return append([]byte{Tag | Item, URI}, buf...), nil
	default:
		return runMarshal(v)
	}
}

func runMarshal(v interface{}) ([]byte, error) {
	var buf bytes.Buffer
	if err := marshal(reflect.ValueOf(v), &buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func marshal(v reflect.Value, buf *bytes.Buffer) error {
	switch k := v.Kind(); k {
	case reflect.Struct:
		if err := encodeLength(Map, uint64(v.NumField()), buf); err != nil {
			return err
		}
		for i := 0; i < v.NumField(); i++ {
			if !v.CanSet() {
				return nil
			}
			f := v.Field(i)
			if err := marshal(f, buf); err != nil {
				return err
			}
		}
	case reflect.Map:
		if err := encodeLength(Map, uint64(v.Len()), buf); err != nil {
			return err
		}
		for _, k := range v.MapKeys() {
			if err := marshal(k, buf); err != nil {
				return err
			}
			if err := marshal(v.MapIndex(k), buf); err != nil {
				return err
			}
		}
	case reflect.Slice:
		if err := encodeLength(Slice, uint64(v.Len()), buf); err != nil {
			return err
		}
		for i := 0; i < v.Len(); i++ {
			if err := marshal(v.Index(i), buf); err != nil {
				return err
			}
		}
	case reflect.Interface:
		return marshal(reflect.ValueOf(v.Interface()), buf)
	case reflect.Ptr:
		return marshal(v.Elem(), buf)
	default:
		return encode(v, buf)
	}
	return nil
}

func encode(v reflect.Value, buf *bytes.Buffer) error {
	switch k := v.Kind(); k {
	case reflect.Invalid:
		buf.WriteByte(byte(Other | Undefined))
	case reflect.String:
		tag := Bin
		if utf8.ValidString(v.String()) {
			tag = String
		}
		encodeLength(tag, uint64(len(v.String())), buf)
		buf.Write([]byte(v.String()))
	case reflect.Bool:
		if v.Bool() {
			buf.WriteByte(Other | True)
		} else {
			buf.WriteByte(Other | False)
		}
	case reflect.Float32:
		buf.WriteByte(Other | Float32)
		
		val := math.Float32bits(float32(v.Float()))
		binary.Write(buf, binary.BigEndian, val)
	case reflect.Float64:
		buf.WriteByte(Other | Float64)
		
		val := math.Float64bits(v.Float())
		binary.Write(buf, binary.BigEndian, val)
	case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Int:
		switch val := v.Int(); {
		case val >= 0:
			return encode(reflect.ValueOf(uint64(val)), buf)
		case val > -24:
			buf.WriteByte(Int | byte(val))
		case val >= math.MinInt8:
			buf.WriteByte(Int | Len1)
			binary.Write(buf, binary.BigEndian, int8(val))
		case val >= math.MinInt16:
			buf.WriteByte(Int | Len2)
			binary.Write(buf, binary.BigEndian, int16(val))
		case val >= math.MinInt32:
			buf.WriteByte(Int | Len4)
			binary.Write(buf, binary.BigEndian, int32(val))
		case val >= math.MinInt64:
			buf.WriteByte(Int | Len8)
			binary.Write(buf, binary.BigEndian, int64(val))
		}
	case reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uint:
		switch val := v.Uint(); {
		case val < 24:
			buf.WriteByte(Uint | byte(val))
		case val <= math.MaxUint8:
			buf.WriteByte(Uint | Len1)
			binary.Write(buf, binary.BigEndian, uint8(val))
		case val <= math.MaxUint16:
			buf.WriteByte(Uint | Len2)
			binary.Write(buf, binary.BigEndian, uint16(val))
		case val <= math.MaxUint32:
			buf.WriteByte(Uint | Len4)
			binary.Write(buf, binary.BigEndian, uint32(val))
		case val <= math.MaxUint64:
			buf.WriteByte(Uint | Len8)
			binary.Write(buf, binary.BigEndian, uint64(val))
		}
	default:
		return UnsupportedTypeErr(k)
	}
	return nil
}

func encodeLength(tag byte, length uint64, buf *bytes.Buffer) error {
	var size interface{}

	switch {
	case length < uint64(Len1):
		buf.WriteByte(tag | byte(length))
		return nil
	case length <= math.MaxUint8:
		tag |= Len1
		size = uint8(length)
	case length <= math.MaxUint16:
		tag |= Len2
		size = uint16(length)
	case length <= math.MaxUint32:
		tag |= Len4
		size = uint32(length)
	case length <= math.MaxUint64:
		tag |= Len8
		size = uint64(length)
	default:
		switch tag {
		case Slice, Map, Bin, String:
			tag |= Indef
		default:
			return InvalidTagErr(tag)
		}
		return ErrTooManyValues
	}
	buf.WriteByte(tag)
	if size != nil {
		binary.Write(buf, binary.BigEndian, size)
	}
	return nil
}
