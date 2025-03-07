package main

import (
	"strings"
	"testing"

	"github.com/michael-disalvo/gapbuf"
)

func gapBufFromStr(s string) gapbuf.GapBuffer {
	rdr := strings.NewReader(s)
	buf, err := gapBufFromReader(rdr)
	if err != nil {
		panic("could not build gap buf")
	}
	return buf
}

func TestNumCellsFromGapBuf(t *testing.T) {
	tests := []string{
		"abc\nd\n\ne\n\t\n",
		"abcd",
		"\n",
		"\na\n\n",
		"🦆abc\n",
	}

	expecteds := [][]int{
		{3, 1, 0, 1, 4, 0},
		{4},
		{0, 0},
		{0, 1, 0, 0},
		{4, 0},
	}

	for i, str := range tests {
		expected := expecteds[i]
		buf := gapBufFromStr(str)
		numCells := numCellsFromGapBuf(&buf)

		if len(numCells) != len(expected) {
			t.Errorf("expected %v num lines but got %v", len(expected), len(numCells))
		}

		for i, expectedNum := range numCells {
			if expectedNum != numCells[i] {
				t.Errorf("expected line %v to have %v cells but got %v", i, expectedNum, numCells[i])
			}
		}
	}
}

func TestGapBufferFromReader(t *testing.T) {
	test_strings := []string{"abc\nd\n\ne\tf", "", "abc", "\n\n\n", "\t\t"}
	for _, str := range test_strings {
		rdr := strings.NewReader(str)
		buf, err := gapBufFromReader(rdr)
		if err != nil {
			t.Errorf("unexpected err building gap buf: %v", err)
		}
		if len(str) != buf.Len() {
			t.Errorf("buf should have been len %v but was %v", len(str), buf.Len())
		}
		for idx, ch := range str {
			if ch != buf.Get(idx) {
				t.Errorf("index %v in buffer should have been %v, but got %v", idx, string(ch), string(buf.Get(idx)))
			}
		}
	}
}
