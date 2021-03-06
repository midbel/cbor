package cbor

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math"
)

var (
	ErrTooLarge   = errors.New("cbor: too large")
	ErrOutOfRange = errors.New("cbor: out of range")
)

type UnsupportedError string

func (u UnsupportedError) Error() string {
	return fmt.Sprintf("cbor: unsupported data type %q", string(u))
}

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
	Simple
)

const (
	Float16 byte = iota + 25
	Float32
	Float64
)

const (
	TagRFC3339 = 0x00
	TagUnix    = 0x01
	TagItem    = 0x18
	TagURI     = 0x20
	TagRegex   = 0x23
)

const (
	Len1 byte = iota + 24
	Len2
	Len4
	Len8
)

func Debug(w io.Writer, bs []byte) error {
	return DebugReader(w, bytes.NewReader(bs))
}

func DebugReader(w io.Writer, r io.Reader) error {
	return debugReader(w, bufio.NewReader(r), true)
}

func debugReader(w io.Writer, r *bufio.Reader, nl bool) error {
	b, err := r.ReadByte()
	if err != nil {
		return err
	}

	switch m, a := b&0xE0, b&0x1F; m {
	case Uint:
		err = debugUint(w, r, a)
	case Int:
		err = debugInt(w, r, a)
	case Bin:
	case String:
		err = debugString(w, r, a)
	case Array:
		err = debugArray(w, r, a)
	case Map:
		err = debugMap(w, r, a)
	case Tag:
	case Other:
		err = debugOther(w, r, a)
	}
	if nl {
		fmt.Fprintln(w)
	}
	return err
}

func debugArray(w io.Writer, r io.Reader, a byte) error {
	size, err := sizeof(r, a)
	if err != nil {
		return err
	}
	var buf bytes.Buffer
	buf.WriteString("[")

	rs := bufio.NewReader(r)
	for i := 0; i < size; i++ {
		if err := debugReader(&buf, rs, false); err != nil {
			return err
		}
		if i < size-1 {
			buf.WriteString(", ")
		}
	}
	buf.WriteString("]")
	io.Copy(w, &buf)
	return nil
}

func debugMap(w io.Writer, r io.Reader, a byte) error {
	size, err := sizeof(r, a)
	if err != nil {
		return err
	}
	var buf bytes.Buffer
	buf.WriteString("{")

	rs := bufio.NewReader(r)
	for i := 0; i < size; i++ {
		if err := debugReader(&buf, rs, false); err != nil {
			return err
		}
		buf.WriteString(": ")
		if err := debugReader(&buf, rs, false); err != nil {
			return err
		}
		if i < size-1 {
			buf.WriteString(", ")
		}
	}
	buf.WriteString("}")
	io.Copy(w, &buf)
	return nil
}

func debugOther(w io.Writer, r io.Reader, a byte) error {
	if a < False {
		fmt.Fprintf(w, fmt.Sprintf("simple(%d)", a))
		return nil
	}
	switch a {
	case Simple:
		var v uint8
		binary.Read(r, binary.BigEndian, &v)
		fmt.Fprintf(w, fmt.Sprintf("simple(%d)", v))
	case False:
		fmt.Fprint(w, "false")
	case True:
		fmt.Fprint(w, "true")
	case Nil:
		fmt.Fprint(w, "null")
	case Undefined:
		fmt.Fprint(w, "undefined")
	case Float16:
	case Float32:
		var v uint32
		binary.Read(r, binary.BigEndian, &v)
		fmt.Fprintf(w, "%g", math.Float32frombits(v))
	case Float64:
		var v uint64
		binary.Read(r, binary.BigEndian, &v)
		fmt.Fprintf(w, "%g", math.Float64frombits(v))
	}
	return nil
}

func debugString(w io.Writer, r io.Reader, a byte) error {
	size, err := sizeof(r, a)
	if err != nil {
		return err
	}
	bs := make([]byte, size)
	if _, err := io.ReadFull(r, bs); err != nil {
		return err
	}
	fmt.Fprintf(w, "%q", string(bs))
	return nil
}

func debugUint(w io.Writer, r io.Reader, a byte) error {
	var (
		err  error
		item uint
	)
	switch a {
	case Len1:
		var v uint8
		err = binary.Read(r, binary.BigEndian, &v)
		item = uint(v)
	case Len2:
		var v uint16
		err = binary.Read(r, binary.BigEndian, &v)
		item = uint(v)
	case Len4:
		var v uint32
		err = binary.Read(r, binary.BigEndian, &v)
		item = uint(v)
	case Len8:
		var v uint64
		err = binary.Read(r, binary.BigEndian, &v)
		item = uint(v)
	default:
		item = uint(a)
	}
	if err != nil {
		return err
	}
	fmt.Fprintf(w, "%d", item)
	return nil
}

func debugInt(w io.Writer, r io.Reader, a byte) error {
	var (
		err  error
		item int
	)
	switch a {
	case Len1:
		var v int8
		err = binary.Read(r, binary.BigEndian, &v)
		item = int(v)
	case Len2:
		var v int16
		err = binary.Read(r, binary.BigEndian, &v)
		item = int(v)
	case Len4:
		var v int32
		err = binary.Read(r, binary.BigEndian, &v)
		item = int(v)
	case Len8:
		var v int64
		err = binary.Read(r, binary.BigEndian, &v)
		item = int(v)
	default:
		item = int(a)
	}
	if err != nil {
		return err
	}
	fmt.Fprintf(w, "%d", -1-item)
	return nil
}

func sizeof(r io.Reader, a byte) (int, error) {
	var (
		size int
		err  error
	)
	switch a {
	case Len1:
		var v uint8
		err = binary.Read(r, binary.BigEndian, &v)
		size = int(v)
	case Len2:
		var v uint16
		err = binary.Read(r, binary.BigEndian, &v)
		size = int(v)
	case Len4:
		var v uint32
		err = binary.Read(r, binary.BigEndian, &v)
		size = int(v)
	case Len8:
		var v uint64
		err = binary.Read(r, binary.BigEndian, &v)
		size = int(v)
	default:
		size = int(a)
	}
	return size, err
}
