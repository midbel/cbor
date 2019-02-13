package cbor

import (
	"encoding/hex"
	"strings"
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

func TestUnmarshalStrings(t *testing.T) {
	data := []struct {
		Raw  string
		Want string
	}{
		{Raw: "60", Want: ""},
		{Raw: "6161", Want: "a"},
		{Raw: "6449455446", Want: "IETF"},
		{Raw: "62c3bc", Want: "\u00fc"},
		{Raw: "63e6b0b4", Want: "\u6c34"},
	}
	for i, d := range data {
		bs, err := hex.DecodeString(d.Raw)
		if err != nil {
			t.Errorf("decode raw string fail (%d): %v", i+1, err)
			continue
		}
		var v string
		if err := Unmarshal(bs, &v); err != nil {
			t.Errorf("unmarshal fail (%d): %v", i+1, err)
			continue
		}
		if v != d.Want {
			t.Errorf("%d value badly decoded: want %s, got %s", i+1, d.Want, v)
		}
	}
}

func TestUnmarshalArrayStrings(t *testing.T) {
	data := []struct {
		Raw string
		Want []string
	} {
		{Raw: "826568656C6C6F65776F726C64", Want: []string{"hello", "world"}},
	}
	for i, d := range data {
		bs, err := hex.DecodeString(d.Raw)
		if err != nil {
			t.Errorf("decode raw string fail (%d): %v", i+1, err)
			continue
		}
		for i := 0; i <= len(d.Want); i++ {
			testUnmarshalArrayStrings(t, bs, d.Want, i)
		}
	}
}

func testUnmarshalArrayStrings(t *testing.T, bs []byte, values []string, n int) {
	var v []string
	if n >= 0 {
		v = make([]string, n)
	}
	if err := Unmarshal(bs, &v); err != nil {
		t.Errorf("unmarshal fail: %v", err)
		return
	}
	if strings.Join(v, ",") != strings.Join(values, ",") {
		t.Errorf("value badly decoded: want %v, got %v" , values, v)
	}
}
