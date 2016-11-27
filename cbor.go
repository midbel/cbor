package cbor

import (
	"errors"
	"fmt"
	"reflect"
)

const (
	Uint byte = iota << 5
	Int
	Bin
	String
	Slice
	Map
	Tag
	Other
)

const (
	Len1 byte = iota + 24
	Len2
	Len4
	Len8
)

const (
	False = Other | (20 + iota)
	True
	Nil
	Undefined
)

const (
	Float16 = Other | (25 + iota)
	Float32
	Float64
)

const Indef = 31

const Stop = 0xFF

const maskTag = 0x1F

type InvalidTagErr byte

func (i InvalidTagErr) Error() string {
	return fmt.Sprintf("invalid tag found %#02x", byte(i))
}

type UnsupportedTypeErr reflect.Kind

func (u UnsupportedTypeErr) Error() string {
	return fmt.Sprintf("cbor: unsupported type: %s", reflect.Kind(u))
}

var (
	TooManyValuesErr = errors.New("cbor: too many values")
	TooFewValuesErr  = errors.New("cbor: too few values")
)
