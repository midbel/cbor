package cbor

import (
	"bytes"
	"encoding/hex"
	"strings"
	"testing"
)

type debugunit struct {
	Raw  string
	Want string
}

func TestDebugArray(t *testing.T) {
	data := []debugunit{
		{Raw: "80", Want: "[]\n"},
		{Raw: "83010203", Want: "[1, 2, 3]\n"},
	}
	testDebug(t, data)
}

func TestDebugMap(t *testing.T) {
	data := []debugunit{
		{Raw: "a0", Want: "{}\n"},
		{Raw: "a201020304", Want: "{1: 2, 3: 4}\n"},
		{Raw: "a26161016162820203", Want: "{\"a\": 1, \"b\": [2, 3]}\n"},
	}
	testDebug(t, data)
}

func TestDebugUnsigned(t *testing.T) {
	data := []debugunit{
		{Raw: "01", Want: "1\n"},
		{Raw: "0a", Want: "10\n"},
		{Raw: "17", Want: "23\n"},
		{Raw: "1818", Want: "24\n"},
		{Raw: "1819", Want: "25\n"},
	}
	testDebug(t, data)
}

func TestDebugSigned(t *testing.T) {
	data := []debugunit{
		{Raw: "20", Want: "-1\n"},
		{Raw: "29", Want: "-10\n"},
		{Raw: "3863", Want: "-100\n"},
		{Raw: "3903e7", Want: "-1000\n"},
	}
	testDebug(t, data)
}

func TestDebugFloat(t *testing.T) {
	data := []debugunit{
		{Raw: "fa47c35000", Want: "100000.0\n"},
		{Raw: "fa7f7fffff", Want: "3.4028234663852886e+38\n"},
		{Raw: "f9c400", Want: "-4.0\n"},
	}
	testDebug(t, data)
}

func TestDebugMisc(t *testing.T) {
	data := []debugunit{
		{Raw: "f4", Want: "false\n"},
		{Raw: "f5", Want: "true\n"},
		{Raw: "f6", Want: "null\n"},
		{Raw: "f7", Want: "undefined\n"},
	}
	testDebug(t, data)
}

func testDebug(t *testing.T, data []debugunit) {
	for i, d := range data {
		bs, err := hex.DecodeString(d.Raw)
		if err != nil {
			t.Errorf("%d: fail to decode hex string: %s (%s)", i+1, err, d.Raw)
			continue
		}
		var w bytes.Buffer
		if err := Debug(&w, bs); err != nil {
			t.Errorf("%d: failed to debug %s => %s", i+1, d.Raw, err)
			continue
		}
		if got := w.String(); got != d.Want {
			t.Errorf("%d: want %s, got %s", i+1, strings.TrimSpace(d.Want), got)
		}
	}
}
