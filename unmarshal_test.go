package cbor

import (
	"encoding/hex"
	"testing"
)

func TestUnmarshalInt(t *testing.T) {
	data := []struct {
		Raw  string
		Want int
	}{
		{Raw: "20", Want: -1},
		{Raw: "29", Want: -10},
		{Raw: "3863", Want: -100},
		{Raw: "3903e7", Want: -1000},
	}
	for i, d := range data {
		bs, err := hex.DecodeString(d.Raw)
		if err != nil {
			t.Errorf("decode raw string fail (%d): %v", i+1, err)
			continue
		}
		var v int
		if err := Unmarshal(bs, &v); err != nil {
			t.Errorf("unmarshal fail (%d): %v", i+1, err)
			continue
		}
		if v != d.Want {
			t.Errorf("%d value badly decoded: want %d, got %d", i+1, d.Want, v)
		}
	}
}

func TestUnmarshalUint(t *testing.T) {
	data := []struct {
		Raw  string
		Want uint
	}{
		{Raw: "00", Want: 0},
		{Raw: "01", Want: 1},
		{Raw: "0a", Want: 10},
		{Raw: "17", Want: 23},
		{Raw: "1818", Want: 24},
		{Raw: "1819", Want: 25},
	}
	for i, d := range data {
		bs, err := hex.DecodeString(d.Raw)
		if err != nil {
			t.Errorf("decode raw string fail (%d): %v", i+1, err)
			continue
		}
		var v uint
		if err := Unmarshal(bs, &v); err != nil {
			t.Errorf("unmarshal fail (%d): %v", i+1, err)
			continue
		}
		if v != d.Want {
			t.Errorf("%d value badly decoded: want %d, got %d", i+1, d.Want, v)
		}
	}
}
