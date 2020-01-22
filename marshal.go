package cbor

import (
	"bytes"
	"encoding/binary"
	"io"
	"math"
	"reflect"
	"unicode/utf8"
)

func Marshal(v interface{}) ([]byte, error) {
	var b bytes.Buffer
	if err := marshal(&b, reflect.ValueOf(v)); err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

func marshal(b *bytes.Buffer, v reflect.Value) error {
	switch v.Kind() {
	default:
		return UnsupportedError(v.Kind().String())
	case reflect.Invalid:
		binary.Write(b, binary.BigEndian, Other|Undefined)
	case reflect.Ptr:
		if v.IsNil() {
			binary.Write(b, binary.BigEndian, Other|Nil)
		} else {
			return marshal(b, v.Elem())
		}
	case reflect.Interface:
		return marshal(b, reflect.ValueOf(v.Interface()))
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return encodeNumber(b, Uint, v.Uint())
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if i := v.Int(); i >= 0 {
			return encodeNumber(b, Uint, uint64(i))
		} else {
			return encodeNumber(b, Int, uint64(-i-1))
		}
	case reflect.Float32:
		v := math.Float32bits(float32(v.Float()))
		binary.Write(b, binary.BigEndian, Other|Float32)
		binary.Write(b, binary.BigEndian, v)
	case reflect.Float64:
		v := math.Float64bits(v.Float())
		binary.Write(b, binary.BigEndian, Other|Float64)
		binary.Write(b, binary.BigEndian, v)
	case reflect.Bool:
		i := byte(Other | False)
		if v.Bool() {
			i = byte(Other | True)
		}
		binary.Write(b, binary.BigEndian, i)
	case reflect.String:
		s, t := v.String(), String
		if !utf8.ValidString(s) {
			t = Bin
		}
		if err := encodeString(b, t, s); err != nil {
			return err
		}
	case reflect.Slice, reflect.Array:
		z := v.Len()
		if err := encodeLength(b, Array, uint64(z)); err != nil {
			return err
		}
		for i := 0; i < z; i++ {
			if err := marshal(b, v.Index(i)); err != nil {
				return err
			}
		}
	case reflect.Map:
		z := v.Len()
		if err := encodeLength(b, Map, uint64(z)); err != nil {
			return err
		}
		for i, vs := 0, v.MapKeys(); i < z; i++ {
			if err := marshal(b, vs[i]); err != nil {
				return err
			}
			if err := marshal(b, v.MapIndex(vs[i])); err != nil {
				return err
			}
		}
	case reflect.Struct:
		z, t := v.NumField(), v.Type()
		if err := encodeLength(b, Map, uint64(z)); err != nil {
			return err
		}
		for i := 0; i < z; i++ {
			f := t.Field(i)
			if len(f.PkgPath) > 0 {
				continue
			}
			n := f.Name
			if t := f.Tag.Get("cbor"); len(t) > 0 {
				n = t
			}
			if err := encodeString(b, String, n); err != nil {
				return err
			}
			if err := marshal(b, v.Field(i)); err != nil {
				return err
			}
		}
	}
	return nil
}

func encodeLength(w io.Writer, t byte, z uint64) error {
	switch {
	default:
		return ErrTooLarge
	case z < uint64(Len1):
		binary.Write(w, binary.BigEndian, t|byte(z))
	case z <= math.MaxUint8:
		binary.Write(w, binary.BigEndian, t|Len1)
		binary.Write(w, binary.BigEndian, uint8(z))
	case z <= math.MaxUint16:
		binary.Write(w, binary.BigEndian, t|Len2)
		binary.Write(w, binary.BigEndian, uint16(z))
	case z <= math.MaxUint32:
		binary.Write(w, binary.BigEndian, t|Len4)
		binary.Write(w, binary.BigEndian, uint32(z))
	case z <= math.MaxUint64:
		binary.Write(w, binary.BigEndian, t|Len8)
		binary.Write(w, binary.BigEndian, uint64(z))
	}
	return nil
}

func encodeString(w io.Writer, t byte, v string) error {
	r := []byte(v)
	if err := encodeLength(w, t, uint64(len(r))); err != nil {
		return err
	}
	if len(r) > 0 {
		_, err := w.Write(r)
		return err
	}
	return nil
}

func encodeNumber(w io.Writer, t byte, v uint64) error {
	switch {
	default:
		return ErrOutOfRange
	case v < uint64(Len1):
		binary.Write(w, binary.BigEndian, t|byte(v))
	case v <= math.MaxUint8:
		binary.Write(w, binary.BigEndian, t|Len1)
		binary.Write(w, binary.BigEndian, uint8(v))
	case v <= math.MaxUint16:
		binary.Write(w, binary.BigEndian, t|Len2)
		binary.Write(w, binary.BigEndian, uint16(v))
	case v <= math.MaxUint32:
		binary.Write(w, binary.BigEndian, t|Len4)
		binary.Write(w, binary.BigEndian, uint32(v))
	case v <= math.MaxUint64:
		binary.Write(w, binary.BigEndian, t|Len8)
		binary.Write(w, binary.BigEndian, uint64(v))
	}
	return nil
}
