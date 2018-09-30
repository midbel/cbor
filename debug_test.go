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
		{Raw: "80", Want: "array[]\n"},
		{Raw: "83010203", Want: "array[1, 2, 3]\n"},
	}
	testDebug(t, data)
}

func TestDebugMap(t *testing.T) {
	data := []debugunit{
		{Raw: "a0", Want: "map[]"},
		{Raw: "a201020304", Want: "map[]"},
	}
	testDebug(t, data)
}

func TestDebugPositive(t *testing.T) {
	data := []debugunit{
		{Raw: "01", Want: "positive(1)\n"},
		{Raw: "0a", Want: "positive(10)\n"},
		{Raw: "17", Want: "positive(24)\n"},
		{Raw: "1818", Want: "positive(25)\n"},
		{Raw: "1819", Want: "positive(25)\n"},
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
