package cbor

import (
	"encoding/hex"
	"testing"
)

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
	var v uint
	for i, d := range data {
		bs, err := hex.DecodeString(d.Raw)
		if err != nil {
			t.Errorf("decode raw string fail (%d): %v", i+1, err)
			continue
		}
		if err := Unmarshal(bs, &v); err != nil {
			t.Errorf("unmarshal fail (%d): %v", i+1, err)
			continue
		}
		if v != d.Want {
			t.Errorf("test %d value badly decoded: want %d, got %d", i+1, d.Want, v)
		}
	}
}
