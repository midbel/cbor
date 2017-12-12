package cbor

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"net/url"
	"reflect"
	"regexp"
	"time"
	"unicode/utf8"
)

const (
	Uint byte = iota << 5
	Int
	Bin
	String
	Array
	Map
	Tag
	Other
)

const (
	False byte = iota + 20
	True
	Nil
	Undefined
)

const (
	Float16 byte = iota + 25
	Float32
	Float64
)

const (
	IsoTime  byte = 0x00
	UnixTime byte = 0x01
	Item     byte = 0x18
	URI      byte = 0x20
	Regex    byte = 0x23
)

const (
	Len1 byte = iota + 24
	Len2
	Len4
	Len8
)

func Marshal(v interface{}) ([]byte, error) {
	b := new(bytes.Buffer)
	if err := marshal(b, reflect.ValueOf(v)); err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

func Unmarshal(bs []byte, v interface{}) error {
	return unmarshal(bytes.NewReader(bs), reflect.ValueOf(v).Elem())
}

func unmarshal(b *bytes.Reader, v reflect.Value) error {
	return nil
}

func marshal(b *bytes.Buffer, v reflect.Value) error {
	switch v.Kind() {
	default:
		return fmt.Errorf("cbor: %s unsupported data type", v.Kind())
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
		if ok, err := encodeItem(b, v.Interface()); ok || err != nil {
			return err
		}
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

func encodeItem(w io.Writer, v interface{}) (bool, error) {
	var s string
	switch v := v.(type) {
	default:
		return false, nil
	case time.Time:
		binary.Write(w, binary.BigEndian, Tag|IsoTime)
		s = v.UTC().Format(time.RFC3339)
	case url.URL:
		binary.Write(w, binary.BigEndian, Tag|Item)
		binary.Write(w, binary.BigEndian, URI)
		s = v.String()
	case regexp.Regexp:
		binary.Write(w, binary.BigEndian, Tag|Item)
		binary.Write(w, binary.BigEndian, Regex)
		s = v.String()
	}
	return true, encodeString(w, String, s)
}

func encodeLength(w io.Writer, t byte, z uint64) error {
	switch {
	default:
		return fmt.Errorf("cbor: too large")
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
		return fmt.Errorf("cbor: out of range")
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
