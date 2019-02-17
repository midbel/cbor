package cbor

import (
	"encoding/hex"
	"reflect"
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
		{Raw: "f0", Want: 16},
		{Raw: "f818", Want: 24},
		{Raw: "f8ff", Want: 255},
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
		{Raw: "f0", Want: 16},
		{Raw: "f818", Want: 24},
		{Raw: "f8ff", Want: 255},
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
	t.Run("single-string", func(t *testing.T) {
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
	})
	t.Run("array-string", func(t *testing.T) {
		vs := []string{"hello", "world"}
		bs, err := hex.DecodeString("826568656C6C6F65776F726C64")
		if err != nil {
			t.Errorf("decode raw string fail: %v", err)
			return
		}
		for i := -1; i <= len(vs); i++ {
			var v []string
			if i >= 0 {
				v = make([]string, i)
			}
			if err := Unmarshal(bs, &v); err != nil {
				t.Errorf("unmarshal fail with length %d: %v", i, err)
				return
			}
			if !reflect.DeepEqual(v, vs) {
				t.Errorf("value badly decoded: want %v, got %v", vs, v)
				return
			}
		}
	})
}

func TestUnmarshalMap(t *testing.T) {
	bs, err := hex.DecodeString("a4616101616202616304616405")
	if err != nil {
		t.Errorf("fail to decode string: %v", err)
		return
	}
	var m map[string]int
	if err := Unmarshal(bs, &m); err != nil {
		t.Errorf("unmarshal fail: %v", err)
		return
	}
	v := map[string]int{
		"a": 1,
		"b": 2,
		"c": 4,
		"d": 5,
	}
	if !reflect.DeepEqual(m, v) {
		t.Errorf("values does not match: %+v != %+v", m, v)
	}
}

func TestUnmarshalStruct(t *testing.T) {
	type ab struct {
		A int   `cbor:"a"`
		B []int `cbor:"b"`
	}
	a := ab{A: 1, B: []int{2, 3}}
	b := ab{A: 4, B: []int{5, 6}}
	t.Run("single-struct", func(t *testing.T) {

		bs, err := hex.DecodeString("a26161016162820203")
		if err != nil {
			t.Errorf("fail to decode string: %v", err)
			return
		}
		var c ab
		if err := Unmarshal(bs, &c); err != nil {
			t.Errorf("unmarshal fail: %v", err)
			return
		}
		if !reflect.DeepEqual(a, c) {
			t.Errorf("values does not match: %+v != %+v", a, c)
		}
	})
	t.Run("array-struct", func(t *testing.T) {
		bs, err := hex.DecodeString("82a26161016162820203a26161046162820506")
		if err != nil {
			t.Errorf("fail to decode string: %v", err)
			return
		}
		var cs []ab
		if err := Unmarshal(bs, &cs); err != nil {
			t.Errorf("unmarshal fail: %v", err)
			return
		}
		vs := []ab{a, b}
		if !reflect.DeepEqual(cs, vs) {
			t.Errorf("values does not match: %+v != %+v", vs, cs)
		}
	})
}

func TestUnmarshalFloat(t *testing.T) {
	data := []struct {
		Raw  string
		Want float64
	}{
		{Raw: "fa47c35000", Want: 100000.0},
		{Raw: "fa7f7fffff", Want: 3.4028234663852886e+38},
		{Raw: "f9c400", Want: -4.0},
	}
	for i, d := range data {
		bs, err := hex.DecodeString(d.Raw)
		if err != nil {
			t.Errorf("decode raw string fail (%d): %v", i+1, err)
			continue
		}
		var v float64
		if err := Unmarshal(bs, &v); err != nil {
			t.Errorf("unmarshal fail (%d): %v", i+1, err)
			continue
		}
		if v != d.Want {
			t.Errorf("%d value badly decoded: want %f, got %f", i+1, d.Want, v)
		}
	}
}

func TestUnmarshalBool(t *testing.T) {
	data := []struct {
		Raw  string
		Want bool
	}{
		{Raw: "f5", Want: true},
		{Raw: "f4", Want: false},
	}
	for i, d := range data {
		bs, err := hex.DecodeString(d.Raw)
		if err != nil {
			t.Errorf("decode raw string fail (%d): %v", i+1, err)
			continue
		}
		var v bool
		if err := Unmarshal(bs, &v); err != nil {
			t.Errorf("unmarshal fail (%d): %v", i+1, err)
			continue
		}
		if v != d.Want {
			t.Errorf("%d value badly decoded: want %t, got %t", i+1, d.Want, v)
		}
	}
}
