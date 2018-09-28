package cbor

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
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

func Debug(w io.Writer, bs []byte) error {
	return DebugReader(w, bytes.NewReader(bs))
}

func DebugReader(w io.Writer, r io.Reader) error {
	return debugReader(w, bufio.NewReader(r))
}

func debugReader(w io.Writer, r *bufio.Reader) error {
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
	return nil
}

func debugArray(w io.Writer, r io.Reader, a byte) error {
	return nil
}

func debugMap(w io.Writer, r io.Reader, a byte) error {
	return nil
}

func debugOther(w io.Writer, r io.Reader, a byte) error {
	return nil
}

func debugString(w io.Writer, r io.Reader, a byte) error {
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
	if err != nil {
		return err
	}
	bs := make([]byte, size)
	if _, err := io.ReadFull(r, bs); err != nil {
		return err
	}
	fmt.Fprintf(w, "string (%03b_%05b): %s\n", String>>5, a, string(bs))
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
	fmt.Fprintf(w, "uint (%03b_%05b): %d\n", Uint>>5, a, item)
	return nil
}

func debugInt(w io.Writer, r io.Reader, a byte) error {
	fmt.Fprintf(w, "int (%03b_%05b): %d\n", Int>>5, a, 0)
	return nil
}
