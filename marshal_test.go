package cbor

import (
	"fmt"
	"testing"
)

type testunit struct {
	Value interface{}
	Want  string
}

func TestMarshalInt(t *testing.T) {
	data := []testunit{
		{Value: int(-1), Want: "0x20"},
		{Value: int(-500), Want: "0x3901f3"},
		{Value: int(-10), Want: "0x29"},
		{Value: int(-100), Want: "0x3863"},
		{Value: int(-1000), Want: "0x3903e7"},
	}
	testMarshal(t, data)
}

func TestMarshalUint(t *testing.T) {
	data := []testunit{
		{Value: uint(0), Want: "0x00"},
		{Value: uint(1), Want: "0x01"},
		{Value: uint(10), Want: "0x0a"},
		{Value: uint(23), Want: "0x17"},
		{Value: uint(24), Want: "0x1818"},
		{Value: uint(25), Want: "0x1819"},
		{Value: uint(100), Want: "0x1864"},
		{Value: uint(1000), Want: "0x1903e8"},
		{Value: uint(1000000), Want: "0x1a000f4240"},
	}
	testMarshal(t, data)
}

func TestMarshalBool(t *testing.T) {
	data := []testunit{
		{Value: true, Want: "0xf5"},
		{Value: false, Want: "0xf4"},
	}
	testMarshal(t, data)
}

func TestMarshalString(t *testing.T) {
	data := []testunit {
		{Value: "", Want: "0x60"},
		{Value: "a", Want: "0x6161"},
		{Value: "IETF", Want: "0x6449455446"},
	}
	testMarshal(t, data)
}

func TestMarshalFloat(t *testing.T) {
	data := []testunit {
		{Value: float32(0.0), Want: "0xf90000"},
		{Value: float32(-0.0), Want: "0xf98000"},
		{Value: float32(1.0), Want: "0xf93c00"},
		{Value: float64(1.1), Want: "0xfb3ff199999999999a"},
		{Value: float32(1.5), Want: "0xf93e00"},
		{Value: float32(65504.0), Want: "0xf97bff"},
		{Value: float32(100000.0), Want: "0xfa47c35000"},
	}
	testMarshal(t, data)
}

func TestMarshalArray(t *testing.T) {
	data := []testunit {
		{Value: []int{}, Want: "0x80"},
		{Value: []int{1, 2, 3}, Want: "0x83010203"},
		{Value: []interface{}{"a", map[string]string{"b": "c"}}, Want: "0x826161a161626163"},
	}
	testMarshal(t, data)
}

func TestMarshalMap(t *testing.T) {
	data := []testunit {
		{Value: map[string]interface{}{"a": 1, "b": []int{2, 3}}, Want: "0xa26161016162820203"},
	}
	testMarshal(t, data)
}

func testMarshal(t *testing.T, data []testunit) {
	for i, d := range data {
		got, err := Marshal(d.Value)
		if err != nil {
			t.Error(err)
			continue
		}
		if s := fmt.Sprintf("%#x", got); d.Want != s {
			r := "%3d: %v => want: %s, got: %s"
			if _, ok := d.Value.(fmt.Stringer); ok {
				r = "%3d: %s => want: %s, got: %s"
			}
			t.Errorf(r, i+1, d.Value, d.Want, s)
		}
	}
}
