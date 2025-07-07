package gapbuf

import (
	"testing"
)

func (self *GapBuffer) zeroGap() {
	for i := self.gapStart; i < self.gapStart+self.gapLen; i++ {
		self.arr[i] = '_'
	}
}

func testGapBuffer() GapBuffer {
	return GapBuffer{
		arr:      []rune{'h', 'e', '_', '_', '_', 'l', 'l', 'o'},
		gapStart: 2,
		gapLen:   3,
	}
}

func TestNewGapBuffer(t *testing.T) {
	buf := NewGapBuffer()
	if len(buf.arr) != 0 {
		t.Errorf("new gap buffer should not allocate")
	}
	if buf.gapLen != 0 {
		t.Errorf("new gap buffer should have 0 size")
	}
	if buf.gapStart != 0 {
		t.Errorf("new gap buffer should have 0 gapStart")
	}
}

func TestGapBufferPush(t *testing.T) {
	buf := NewGapBuffer()
	runes := []rune{'1', '2', '3', '4'}
	for i, ch := range runes {
		buf.Push(ch)
		if buf.Get(i) != ch {
			t.Errorf("buf.Get(%v) expected %v, got %v", i, ch, buf.Get(i))
		}
	}
}

func TestGapBufferGet(t *testing.T) {
	buf := testGapBuffer()

	inputs := []int{0, 2}
	expected := []rune{'h', 'l'}
	for i := range len(inputs) {
		in := inputs[i]
		exp := expected[i]
		if ch := buf.Get(in); ch != exp {
			t.Errorf("buf.Get(%v) expected %v, got %v", in, string(exp), string(ch))
		}
	}

	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Expected buf.Get(5) to panic")
		}
	}()
	buf.Get(5)
}

func TestGapBufferFrom(t *testing.T) {
	buf := GapBufferFrom("hello")
	if s := string(buf.arr); s != "hello" {
		t.Errorf("GapBufferFrom() return arr of %v", string(buf.arr))
	}
}

func TestGapBufferBasics(t *testing.T) {
	buf := testGapBuffer()
	if l := buf.Len(); l != 5 {
		t.Errorf("buf.Len() expected 5, got %v", l)
	}

	buf.Set(4, 'X')
	if ch := buf.Get(4); ch != 'X' {
		t.Errorf("buf.Get(4) expected `X` got `%v`", string(ch))
	}
}

func TestGapBufferInsert(t *testing.T) {
	buf := testGapBuffer()

	expected := []string{
		"he1__llo",
		"h2_e1llo",
		"3h2e1llo",
		"4_________3h2e1llo",
		"43h2e1llo5________",
		"43h26_______e1llo5",
	}

	idxs := []int{2, 1, 0, 0, 9, 4}
	chs := []rune{'1', '2', '3', '4', '5', '6'}

	for i := range len(expected) {
		exp := expected[i]
		idx := idxs[i]
		ch := chs[i]

		buf.Insert(idx, ch)
		buf.zeroGap()
		if exp != string(buf.arr) {
			t.Errorf("buf.Insert(%v, %v) expected %v, got %v", idx, string(ch), exp, string(buf.arr))
		}
	}
}

func TestZeroGap(t *testing.T) {
	buf := GapBuffer{
		arr:      []rune{'t', 'e', 's', 't'},
		gapStart: 2,
		gapLen:   0,
	}
	buf.moveGap(0)
	if buf.gapStart != 0 {
		t.Errorf("Expected Gap to be at 0, at %v", buf.gapStart)
	}

	buf.moveGap(3)
	if buf.gapStart != 3 {
		t.Errorf("Expected Gap to be at 3, at %v", buf.gapStart)
	}
}

func TestRemove(t *testing.T) {
	buf := testGapBuffer()

	ch := buf.Remove(1)
	if ch != rune('e') {
		t.Errorf("Expected `e` to be returned from Remove(1), got %v", ch)
	}

	if buf.gapStart != 1 {
		t.Errorf("Expected gap to be at location of Remove, got %v", buf.gapStart)
	}

	if buf.gapLen != 4 {
		t.Errorf("Expected gap len to be 4, got %v", buf.gapLen)
	}

	buf.Insert(1, 'e')
	ch = buf.Remove(4)
	if ch != rune('o') {
		t.Errorf("Expected `l`, got %v", ch)
	}

	if buf.gapStart != 4 {
		t.Errorf("Expected gap start to be 4, got %v", buf.gapStart)
	}

	buf = GapBuffer{
		arr:      []rune{'t', 'e', 's', 't'},
		gapStart: 0,
		gapLen:   0,
	}
	ch = buf.Remove(3)

	if ch != rune('t') {
		t.Errorf("Expected ch to be `t`, got %v", ch)
	}

	if buf.gapStart != 3 {
		t.Errorf("Expected gap start to be 43, got %v", buf.gapStart)
	}

	if buf.gapLen != 1 {
		t.Errorf("Expected gap to be len 1, got %v", buf.gapLen)
	}

}
