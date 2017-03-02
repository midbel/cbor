package cbor

import (
	"errors"
	"fmt"
	"reflect"
)

//Major types
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

//Optional Tagging items
const (
	IsoTime  = 0x00
	UnixTime = 0x01
	Item     = 0x18
	URI      = 0x20
)

const (
	Len1 byte = iota + 24
	Len2
	Len4
	Len8
)

const (
	False = 20 + iota
	True
	Nil
	Undefined
)

const (
	Float16 = 25 + iota
	Float32
	Float64
)

const (
	Indef = 0x1F
	Stop  = 0xFF
)

const mask = 0x1F

type InvalidTagErr byte

func (i InvalidTagErr) Error() string {
	return fmt.Sprintf("invalid tag found %#02x", byte(i))
}

type UnassignedTagErr byte

func (u UnassignedTagErr) Error() string {
	return fmt.Sprintf("unassigned tag %#02x", byte(u))
}

type UnsupportedTypeErr reflect.Kind

func (u UnsupportedTypeErr) Error() string {
	return fmt.Sprintf("cbor: %s not supported", reflect.Kind(u))
}

var (
	ErrTooManyValues = errors.New("cbor: too many values")
	ErrTooFewValues  = errors.New("cbor: too few values")
)
