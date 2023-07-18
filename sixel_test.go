package main

import (
	"strings"
	"testing"
)

func TestSixelSize(t *testing.T) {
	header := "\x1bP0;1;0q"
	colors := "#0;2;97;10;50#1;2;10;50;75"
	tests := []struct {
		si     string
		hi     int
		succ   bool
		wo, ho int
	}{
		{
			header + colors + "#0v!10~$#1!11~-!11~\x1b\\", 30,
			true, 11, 12,
		},
		{ // Size reported in raster attributes do not limit actual size
			header + `"1;1;1;6` + colors + "#0~~-#1~~-\x1b\\", 30,
			true, 2, 12,
		},
		{ // missing trailing '-' still reports correct height
			header + colors + "#0~-~\x1b\\", 30,
			true, 1, 12,
		},
		{ // oversized image is rejected
			header + colors + "#0~-~-\x1b\\", 11,
			false, 0, 0,
		},
		{ // repeat introducer
			header + colors + "#0!120~-!100~-\x1b\\", 30,
			true, 120, 12,
		},
		{ // '$' carriage return
			header + colors + "#0FFFF$#1ww-\x1b\\", 30,
			true, 4, 6,
		},
		{ // '$' carriage return
			header + colors + "#0FF$#1wwww-\x1b\\", 30,
			true, 4, 6,
		},
	}

	for i, test := range tests {
		reader := strings.NewReader(test.si)
		w, h, err := sixelSize(reader, test.hi)
		println(i, test.si)
		println()

		if !test.succ {
			if err == nil {
				t.Errorf("test #%d expected to fail", i)
			}
			continue
		} else if err != nil {
			t.Errorf("test #%d failed with error %s", i, err)
			continue
		}

		if w != test.wo || h != test.ho {
			t.Errorf("test #%d expected (%d, %d), got (%d, %d)", i, test.wo, test.ho, w, h)
		}
	}

}
