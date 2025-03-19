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

func newEditSession(s string) EditSession {
	rdr := strings.NewReader(s)
	session, err := buildEditSession(rdr, "testfile")
	if err != nil {
		panic("could not build edit session")
	}
	return session
}

func TestMoveRight(t *testing.T) {
	text := "first\tline\nsecond line\n\nfourth line"

	session := newEditSession(text)
	session.cursor = Cursor{10, 3, 34, true}

	expectedCursor := Cursor{11, 3, 34, false}
	session.moveCursorRight()
	if session.cursor != expectedCursor {
		t.Errorf("expected cursor to be %v but got %v", expectedCursor, session.cursor)
	}

	session = newEditSession(text)
	session.cursor = Cursor{0, 2, 22, false}
	expectedCursor = Cursor{0, 2, 22, false}
	session.moveCursorRight()
	if session.cursor != expectedCursor {
		t.Errorf("expected cursor to be %v but got %v", expectedCursor, session.cursor)
	}

	session = newEditSession(text)
	expectedCursor = Cursor{1, 0, 1, true}
	session.moveCursorRight()
	if session.cursor != expectedCursor {
		t.Errorf("expected cursor to be %v but got %v", expectedCursor, session.cursor)
	}

	expectedCursor = Cursor{2, 0, 2, true}
	session.moveCursorRight()
	if session.cursor != expectedCursor {
		t.Errorf("expected cursor to be %v but got %v", expectedCursor, session.cursor)
	}

	session.moveCursorRight() // s
	session.moveCursorRight() // t
	session.moveCursorRight() // \t
	session.moveCursorRight() // l

	expectedCursor = Cursor{9, 0, 6, true}
	if session.cursor != expectedCursor {
		t.Errorf("expected cursor to be %v but got %v", expectedCursor, session.cursor)
	}

	session.moveCursorRight() // i
	session.moveCursorRight() // n
	session.moveCursorRight() // e
	session.moveCursorRight() // <invalid>

	expectedCursor = Cursor{13, 0, 9, false}
	if session.cursor != expectedCursor {
		t.Errorf("expected cursor to be %v but got %v", expectedCursor, session.cursor)
	}
	session.moveCursorRight() // <invalid>
	expectedCursor = Cursor{13, 0, 9, false}
	if session.cursor != expectedCursor {
		t.Errorf("expected cursor to be %v but got %v", expectedCursor, session.cursor)
	}

}

func TestNewEditSession(t *testing.T) {
	text := "this is a nice file\nsecond line\n\nfourth line\n"
	session := newEditSession(text)

	if session.numLines() != 5 {
		t.Errorf("expected 5 lines but got %v", session.numLines())
	}

	firstLineUsedCells := session.numCells[0]
	if firstLineUsedCells != 19 {
		t.Errorf("expected 19 num cells in first line but got %v", firstLineUsedCells)
	}

	thirdLineUsedCells := session.numCells[2]
	if thirdLineUsedCells != 0 {
		t.Errorf("expected empty line to have 0 used cells but had %v", thirdLineUsedCells)
	}

	text2 := "no newline"
	session = newEditSession(text2)
	if session.numLines() != 1 {
		t.Errorf("expected 1 line with no newline but got %v", session.numLines())
	}

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
