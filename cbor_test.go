package cbor

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"net/url"
	"testing"
	"time"
)

type Sample struct {
	In  interface{}
	Out string
}

var sampleTimes = []Sample{
	{time.Date(2013, 3, 21, 20, 4, 0, 0, time.UTC), "c074323031332d30332d32315432303a30343a30305a"},
	{time.Date(2017, 2, 28, 20, 38, 0, 0, time.UTC), "c074323031372d30322d32385432303a33383a30305a"},
}

var sampleURIs = []Sample{
	{
		In:  url.URL{Scheme: "http", Host: "www.example.com"},
		Out: "d82076687474703a2f2f7777772e6578616d706c652e636f6d",
	},
	{
		In:  url.URL{Scheme: "http", Host: "www.google.com", RawQuery: "q=golang"},
		Out: "d820781e687474703a2f2f7777772e676f6f676c652e636f6d3f713d676f6c616e67",
	},
}

var sampleOthers = []Sample{
	{false, "f4"},
	{true, "f5"},
	{nil, "f7"},
}

var sampleStrings = []Sample{
	{"", "60"},
	{"a", "6161"},
	{"IETF", "6449455446"},
	{"\"\\", "62225c"},
}

var sampleUints = []Sample{
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
}

var sampleInts = []Sample{
	{-1, "20"},
	{-10, "29"},
	{-100, "3863"},
	{-1000, "3903e7"},
}

var sampleFloats = []Sample{
	//{0.0, "f90000"},
	//{-0.0, "f98000"},
	//{1.0, "f93c00"},
	//{1.5, "f93e00"},
	//{65504.0, "f97bff"},
	//{5.960464477539063e-8, "f90001"},
	//{0.00006103515625, "f90400"},
	//{-4.0, "f9c400"},
	{float64(1.1), "fb3ff199999999999a"},
	{float32(100000.0), "fa47c35000"},
	{float32(3.4028234663852886e+38), "fa7f7fffff"},
	{float64(1.0e+300), "fb7e37e43c8800759c"},
	{float64(-4.1), "fbc010666666666666"},
}

var sampleArrays = []Sample{
	{[]int{1, 2, 3}, "83010203"},
	{[]interface{}{1, []int{2, 3}, []int{4, 5}}, "8301820203820405"},
	{[]interface{}{"a", map[string]string{"b": "c"}}, "826161bf61626163ff"},
}

var sampleMaps = []Sample{
	{map[string]interface{}{}, "a0"},
	{map[int]int{1: 2, 3: 4}, "a201020304"},
	{map[string]interface{}{"a": 1, "b": []int{2, 3}}, "a26161016162820203"},
	{map[string]interface{}{"Fun": true, "Amt": -2}, "bf6346756ef563416d7421ff"},
}

func TestMaps(t *testing.T) {
	runTests(t, sampleMaps)
}

func TestArrays(t *testing.T) {
	runTests(t, sampleArrays)
}

func TestTimes(t *testing.T) {
	runTests(t, sampleTimes)
}

func TestURIs(t *testing.T) {
	runTests(t, sampleURIs)
}

func TestStrings(t *testing.T) {
	runTests(t, sampleStrings)
}

func TestOthers(t *testing.T) {
	runTests(t, sampleOthers)
}

func TestInts(t *testing.T) {
	runTests(t, sampleInts)
}

func TestUints(t *testing.T) {
	runTests(t, sampleUints)
}

func TestFloats(t *testing.T) {
	runTests(t, sampleFloats)
}

func runTests(t *testing.T, s []Sample) {
	sample := map[string]func([]Sample, *testing.T){
		"marshal":   testMarshalSample,
		"unmarshal": testUnmarshalSample,
	}
	for n, f := range sample {
		t.Run(n, func(t *testing.T) {
			f(s, t)
		})
	}
}

func testMarshalSample(sample []Sample, t *testing.T) {
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

func testUnmarshalSample(sample []Sample, t *testing.T) {
	for i, s := range sample {
		buf, err := hex.DecodeString(s.Out)
		if err != nil {
			t.Errorf("invalid encoded cbor item %s: %s", s.Out, err)
			continue
		}

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
