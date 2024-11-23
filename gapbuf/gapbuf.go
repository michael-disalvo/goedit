package gapbuf

import (
	"fmt"
	"slices"
)

const growAmount = 10

type GapBuffer struct {
	arr      []rune
	gapStart int
	gapLen   int
}

func (self *GapBuffer) grow() {
	amount := growAmount
	self.arr = slices.Grow(self.arr, amount)
	for _ = range amount {
		self.arr = append(self.arr, '_')
	}
	self.gapStart = len(self.arr) - amount
	self.gapLen = amount
}

func (self *GapBuffer) innerLen() int {
	return len(self.arr) - self.gapLen
}

func (self *GapBuffer) mapIndex(idx int) int {
	if idx < self.gapStart {
		return idx
	}
	return idx + self.gapLen
}

func (self *GapBuffer) shiftGapLeft(times int) {
	for i := 0; i < times && self.gapStart > 0; i++ {
		self.gapStart--
		self.arr[self.gapStart+self.gapLen] = self.arr[self.gapStart]
	}
}

func (self *GapBuffer) shiftGapRight(times int) {
	for i := 0; i < times && self.gapStart < self.innerLen(); i++ {
		self.arr[self.gapStart] = self.arr[self.gapStart+self.gapLen]
		self.gapStart++
	}
}

func (self *GapBuffer) moveGap(idx int) {
	if self.gapStart < idx {
		self.shiftGapRight(idx - self.gapStart)
	} else if idx < self.gapStart {
		self.shiftGapLeft(self.gapStart - idx)
	}
}

func NewGapBuffer() GapBuffer {
	buf := GapBuffer{
		arr:      make([]rune, 0),
		gapStart: 0,
		gapLen:   0,
	}
	return buf
}

func GapBufferFrom(str string) GapBuffer {
	return GapBuffer{
		arr: []rune(str),
	}
}

func (self *GapBuffer) Len() int {
	return self.innerLen()
}

func (self *GapBuffer) Get(idx int) rune {
	if idx < 0 || idx >= self.innerLen() {
		panic(fmt.Sprintf("index %v out of bounds of gap buffer", idx))
	}
	return self.arr[self.mapIndex(idx)]
}

func (self *GapBuffer) Push(ch rune) {
	self.arr = append(self.arr, ch)
}

func (self *GapBuffer) Set(idx int, ch rune) {
	if idx < 0 || idx >= self.innerLen() {
		panic(fmt.Sprintf("cannot set at index %v of gap buffer", idx))
	}
	self.arr[self.mapIndex(idx)] = ch
}

func (self *GapBuffer) Insert(idx int, ch rune) {
	if idx < 0 || idx > self.innerLen() {
		panic(fmt.Sprintf("cannot insert at index %v of gap buffer", idx))
	}
	if self.gapLen == 0 {
		self.grow()
	}
	self.moveGap(idx)
	self.arr[self.gapStart] = ch
	self.gapLen--
	self.gapStart++
}

func (self *GapBuffer) Remove(idx int) (ch rune) {
	if idx < 0 || idx >= self.innerLen() {
		panic(fmt.Sprintf("cannot remove index %v of gap buffer", idx))
	}

	ch = self.arr[self.mapIndex(idx)]

	if idx == self.innerLen() && self.gapLen == 0 {
		self.gapStart = self.innerLen()
		self.gapLen++
		return
	}

	self.moveGap(idx + 1)
	self.gapStart--
	self.gapLen++
	return
}
