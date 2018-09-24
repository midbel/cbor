package cbor

import (
	"errors"
	"fmt"
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
