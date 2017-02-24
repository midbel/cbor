package cbor

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"testing"
)

var sample = []struct{
	In  interface{}
	Out string
} {
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
}

func TestMarshal(t *testing.T) {
	for i, s := range sample {
		buf, err := Marshal(s.In)
		if err != nil {
			t.Errorf("%03d) fail to marshal %v (%s)", i, s.In, err)
			continue
		}
		other, _ := hex.DecodeString(s.Out)
		if !bytes.Equal(buf, other) {
			t.Errorf("%03d) want: %x, got: %x", i, other, buf)
		}
	}
}

func TestUnmarshal(t *testing.T) {
	for i, s := range sample {
		buf, _ := hex.DecodeString(s.Out)
		
		var other interface{}
		if _, err := Unmarshal(buf, &other); err != nil {
			t.Errorf("%03d) fail to unmarshal %v (%s)", i, s.Out, err)
			continue
		}
		if fmt.Sprintf("%v", other) != fmt.Sprintf("%v", s.In) {
			t.Errorf("%03d) want: %v (%[1]T), got: %v (%[2]T)", i, s.In, other)
		}
	}
}
