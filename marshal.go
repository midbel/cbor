package cbor

import (
	"bytes"
	"encoding/binary"
	"io"
	"math"
	"reflect"
	"strings"
	"unicode/utf8"
)

type Encoder struct {
	w       io.Writer
	compact bool
	err     error
}

func NewEncoder(w io.Writer, compact bool) *Encoder {
	return &Encoder{w, compact, nil}
}

func (e *Encoder) Encode(v interface{}) error {
	if e.err != nil {
		return e.err
	}
	buf, err := runMarshal(v, e.compact)
	if err != nil {
		e.err = err
		return err
	}
	if _, err := e.w.Write(buf); err != nil {
		e.err = err
		return err
	}
	return nil
}

func Marshal(v interface{}) ([]byte, error) {
	return runMarshal(v, false)
}

func MarshalCompact(v interface{}) ([]byte, error) {
	return runMarshal(v, true)
}

func runMarshal(v interface{}, compact bool) ([]byte, error) {
	var buf bytes.Buffer
	if err := marshal(reflect.ValueOf(v), &buf, compact); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func marshal(v reflect.Value, buf *bytes.Buffer, compact bool) error {
	switch k := v.Kind(); k {
	case reflect.Struct:
		return marshalStruct(v, buf, compact)
	case reflect.Map:
		return marshalMap(v, buf, compact)
	case reflect.Slice:
		return marshalSlice(v, buf, compact)
	case reflect.Interface:
		return marshal(reflect.ValueOf(v.Interface()), buf, compact)
	case reflect.Ptr:
		return marshal(v.Elem(), buf, compact)
	default:
		return encode(v, buf)
	}
	return nil
}

func marshalStruct(v reflect.Value, buf *bytes.Buffer, compact bool) error {
	tag := Map
	if compact {
		tag = Slice
	}
	if err := encodeLength(tag, uint64(v.NumField()), buf); err != nil {
		return err
	}
	t := v.Type()
	for i := 0; i < v.NumField(); i++ {
		f := v.Field(i)
		if !f.CanSet() {
			continue
		}
		if !compact {
			name := strings.ToLower(t.Field(i).Name)
			if err := encode(reflect.ValueOf(name), buf); err != nil {
				return err
			}
		}
		if err := marshal(f, buf, compact); err != nil {
			return err
		}
	}
	return nil
}

func marshalSlice(v reflect.Value, buf *bytes.Buffer, compact bool) error {
	if err := encodeLength(Slice, uint64(v.Len()), buf); err != nil {
		return err
	}
	for i := 0; i < v.Len(); i++ {
		if err := marshal(v.Index(i), buf, compact); err != nil {
			return err
		}
	}
	return nil
}

func marshalMap(v reflect.Value, buf *bytes.Buffer, compact bool) error {
	if err := encodeLength(Map, uint64(v.Len()), buf); err != nil {
		return err
	}
	for _, k := range v.MapKeys() {
		if err := marshal(k, buf, compact); err != nil {
			return err
		}
		if err := marshal(v.MapIndex(k), buf, compact); err != nil {
			return err
		}
	}
	return nil
}

func encode(v reflect.Value, buf *bytes.Buffer) error {
	switch k := v.Kind(); k {
	case reflect.Invalid:
		buf.WriteByte(byte(Undefined))
	case reflect.String:
		tag := Bin
		if utf8.ValidString(v.String()) {
			tag = String
		}
		encodeLength(tag, uint64(len(v.String())), buf)
		buf.Write([]byte(v.String()))
	case reflect.Bool:
		if v.Bool() {
			buf.WriteByte(True)
		} else {
			buf.WriteByte(False)
		}
	case reflect.Float32:
		val := math.Float32bits(float32(v.Float()))

		buf.WriteByte(Float32)
		binary.Write(buf, binary.BigEndian, val)
	case reflect.Float64:
		val := math.Float64bits(v.Float())

		buf.WriteByte(Float64)
		binary.Write(buf, binary.BigEndian, val)
	case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Int:
		if val := v.Int(); val >= 0 {
			return encode(reflect.ValueOf(uint64(val)), buf)
		} else {
			tag := Int
			var datum interface{}
			switch val := -val - 1; {
			case val <= 24:
				tag |= byte(val)
			case val <= math.MaxInt8:
				tag |= Len1
				datum = int8(val)
			case val <= math.MaxInt16:
				tag |= Len2
				datum = int16(val)
			case val <= math.MaxInt32:
				tag |= Len4
				datum = int32(val)
			case val <= math.MaxInt64:
				tag |= Len8
				datum = int64(val)
			}
			buf.WriteByte(tag)
			if datum != nil {
				binary.Write(buf, binary.BigEndian, datum)
			}
		}
	case reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uint:
		tag := Uint

		var datum interface{}
		switch val := v.Uint(); {
		case val < 24:
			tag |= byte(val)
		case val <= math.MaxUint8:
			tag |= Len1
			datum = uint8(val)
		case val <= math.MaxUint16:
			tag |= Len2
			datum = uint16(val)
		case val <= math.MaxUint32:
			tag |= Len4
			datum = uint32(val)
		case val <= math.MaxUint64:
			tag |= Len8
			datum = uint64(val)
		}
		buf.WriteByte(tag)
		if datum != nil {
			binary.Write(buf, binary.BigEndian, datum)
		}
	default:
		return UnsupportedTypeErr(k)
	}
	return nil
}

func encodeLength(tag byte, length uint64, buf *bytes.Buffer) error {
	var size interface{}

	switch {
	case length < 1<<5:
		tag |= byte(length)
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
		return TooManyValuesErr
	}
	buf.WriteByte(tag)
	if size != nil {
		binary.Write(buf, binary.BigEndian, size)
	}
	return nil
}
