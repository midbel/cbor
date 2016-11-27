package cbor

import (
	"bytes"
	"encoding/hex"
	"testing"
)

func TestMarshalString(t *testing.T) {
	data := []struct{ Text, Want string }{
		{"", "60"},
		{"a", "6161"},
		{"IETF", "6449455446"},
		{"\"\\", "62225c"},
	}
	for _, d := range data {
		buf, err := Marshal(d.Text)
		if err != nil {
			t.Errorf("fail to marshal %s: %s", d.Text, err)
			continue
		}
		other, _ := hex.DecodeString(d.Want)
		if !bytes.Equal(buf, other) {
			t.Errorf("want: %x, got: %x", other, buf)
		}
	}
}

func TestMarshalInt(t *testing.T) {
	data := []struct {
		Value int64
		Want  string
	}{
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

	for _, d := range data {
		buf, err := Marshal(d.Value)
		if err != nil {
			t.Errorf("fail to marshal %s: %s", d.Value, err)
			continue
		}
		other, _ := hex.DecodeString(d.Want)
		if !bytes.Equal(buf, other) {
			t.Errorf("marshal %d => want: %x, got: %x", d.Value, other, buf)
		}
	}
}
