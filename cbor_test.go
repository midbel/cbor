package cbor

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"testing"
)

var sample = []struct {
	In  interface{}
	Out string
}{
	{false, "f4"},
	{true, "f5"},
	{nil, "f7"},
	{"", "60"},
	{"a", "6161"},
	{"IETF", "6449455446"},
	{"\"\\", "62225c"},
	{0, "00"},
	{1, "01"},
	{10, "0a"},
	{23, "17"},
	{24, "1818"},
	{25, "1819"},
	{100, "1864"},
	{1000, "1903e8"},
	{1000000, "1a000f4240"},
	{1000000000000, "1b000000e8d4a51000"},
	{-1, "20"},
	{-10, "29"},
	{-100, "3863"},
	{-1000, "3903e7"},
	{0.0, "f90000"},
	{-0.0, "f98000"},
	{1.0, "f93c00"},
	{1.1, "fb3ff199999999999a"},
	{1.5, "f93e00"},
	{65504.0, "f97bff"},
	{100000.0, "fa47c35000"},
	{3.4028234663852886e+38, "fa7f7fffff"},
	{1.0e+300, "fb7e37e43c8800759c"},
	{5.960464477539063e-8, "f90001"},
	{0.00006103515625, "f90400"},
	{-4.0, "f9c400"},
	{-4.1, "fbc010666666666666"},
}

func TestMarshal(t *testing.T) {
	for i, s := range sample {
		buf, err := Marshal(s.In)
		if err != nil {
			t.Errorf("%03d) fail to marshal %v (%s)", i+1, s.In, err)
			continue
		}
		other, _ := hex.DecodeString(s.Out)
		if !bytes.Equal(buf, other) {
			t.Errorf("%03d) want: %x, got: %x", i+1, other, buf)
		}
	}
}

func TestUnmarshal(t *testing.T) {
	for i, s := range sample {
		buf, _ := hex.DecodeString(s.Out)

		var other interface{}
		if _, err := Unmarshal(buf, &other); err != nil {
			t.Errorf("%03d) fail to unmarshal %#x, should be %v (%s)", i+1, buf, s.In, err)
			continue
		}
		if fmt.Sprintf("%v", other) != fmt.Sprintf("%v", s.In) {
			t.Errorf("%03d) want: %v, got: %v", i+1, s.In, other)
		}
	}
}
