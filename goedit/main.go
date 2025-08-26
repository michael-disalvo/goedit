package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"slices"

	"github.com/mattn/go-runewidth"
	"github.com/nsf/termbox-go"
)

type Cursor struct {
	index            int
	targetCellOffset int
}

func LineOfIndex(buf *Buffer, index int) int {
	y, startsLine := slices.BinarySearch(buf.lineStarts, index)
	if !startsLine {
		y -= 1
	}
	return y
}

func IndexToGrid(buf *Buffer, index int) (int, int) {
	y := LineOfIndex(buf, index)

	x := 0
	lineStartIndex := buf.lineStarts[y]

	for i := lineStartIndex; i < index; i += 1 {
		ch := buf.text[i]
		x += runeWidth(ch)
	}

	return x, y
}

func NewCursor() *Cursor {
	return &Cursor{
		index:            0,
		targetCellOffset: 0,
	}
}

func DisplayCursor(cursor *Cursor, buf *Buffer) {
	x, y := IndexToGrid(buf, cursor.index)
	y -= buf.windowOffset
	termbox.SetCursor(x, y)
}

func RuneOffsetForCellOffset(buf *Buffer, y int, targetCellOffset int) (runeOffset int) {
	index := buf.lineStarts[y]
	currCellOffset := 0
	runesInLine := buf.NumRunesInLine(y)

	for currCellOffset < targetCellOffset && runeOffset < runesInLine {
		currCellOffset += runeWidth(buf.text[index])
		index += 1
		runeOffset += 1
	}

	return
}

func MoveCursorUp(cursor *Cursor, buf *Buffer) {
	y := LineOfIndex(buf, cursor.index)

	if y > 0 {
		prevLineStart := buf.lineStarts[y-1]

		runeOffset := RuneOffsetForCellOffset(buf, y-1, cursor.targetCellOffset)

		cursor.index = prevLineStart + runeOffset

		if y-1 < buf.windowOffset {
			buf.windowOffset -= 1
		}
	}
}

func MoveCursorDown(cursor *Cursor, buf *Buffer, height int) {
	y := LineOfIndex(buf, cursor.index)

	numLines := len(buf.lineStarts)
	if y+1 < numLines {
		nextLineStart := buf.lineStarts[y+1]
		runeOffset := RuneOffsetForCellOffset(buf, y+1, cursor.targetCellOffset)
		cursor.index = nextLineStart + runeOffset
		if y+1 >= buf.windowOffset+height {
			buf.windowOffset += 1
		}
	}

}

func MoveCursorRight(cursor *Cursor, buf *Buffer) {
	y := LineOfIndex(buf, cursor.index)
	numRunesInLine := buf.NumRunesInLine(y)
	indexInLine := cursor.index - buf.lineStarts[y]

	if indexInLine < numRunesInLine {
		cursor.targetCellOffset += runeWidth(buf.text[cursor.index])
		cursor.index += 1
	}
}

func MoveCursorLeft(cursor *Cursor, buf *Buffer) {
	y := LineOfIndex(buf, cursor.index)
	indexInLine := cursor.index - buf.lineStarts[y]
	if indexInLine > 0 {
		cursor.index -= 1
		cursor.targetCellOffset -= runeWidth(buf.text[cursor.index])
	}
}

type Buffer struct {
	text         []rune
	lineStarts   []int
	windowOffset int
}

func SetCell(x, y int, ch rune) {
	termbox.SetCell(x, y, ch, termbox.ColorDefault, termbox.ColorDefault)
}

func (buf *Buffer) String() string {
	return fmt.Sprintf("%v\n=====\nLineStarts:%v", string(buf.text[:len(buf.text)-1]), buf.lineStarts)
}

func (buf *Buffer) NumRunesInLine(y int) (numRunesInLine int) {
	lineStart := buf.lineStarts[y]
	if y+1 < len(buf.lineStarts) {
		numRunesInLine = buf.lineStarts[y+1] - lineStart - 1
	} else {
		numRunesInLine = len(buf.text) - lineStart - 1
	}
	return numRunesInLine
}

func (buf *Buffer) InsertRune(ch rune, index int) {
	newText := slices.Insert(buf.text, index, ch)
	newLineStarts := BuildLineStarts(newText)

	buf.text = newText
	buf.lineStarts = newLineStarts
}

func (buf *Buffer) RemoveRune(index int) {
	if index >= 0 {
		newText := slices.Delete(buf.text, index, index+1)
		newLineStarts := BuildLineStarts(newText)

		buf.text = newText
		buf.lineStarts = newLineStarts
	}
}

func (buf *Buffer) JustText() []byte {
	return []byte(string(buf.text[:len(buf.text)-1]))
}

func (buf *Buffer) Display(height int) {
	x := 0

	for y := buf.windowOffset; y < buf.windowOffset+height && y < len(buf.lineStarts); y += 1 {
		lineStart := buf.lineStarts[y]
		numRunesInLine := buf.NumRunesInLine(y)
		x = 0
		for i := lineStart; i < lineStart+numRunesInLine; i += 1 {
			ch := buf.text[i]
			SetCell(x, y-buf.windowOffset, ch)
			x += runeWidth(ch)
		}
	}

}

func isLineStart(text []rune, i int) bool {
	return i == 0 || text[i-1] == '\n'
}

func BuildLineStarts(text []rune) []int {

	lineStarts := make([]int, 0)
	for i := range text {
		if isLineStart(text, i) {
			lineStarts = append(lineStarts, i)
		}
	}

	return lineStarts
}

func NewBuffer(r io.Reader) (*Buffer, error) {
	bufReader := bufio.NewReader(r)

	text := make([]rune, 0)

	for {
		ch, _, err := bufReader.ReadRune()

		if err == io.EOF {
			break
		} else if err != nil {
			return nil, fmt.Errorf("error trying to build buffer: %w", err)
		}

		text = append(text, ch)
	}

	text = append(text, -1)

	lineStarts := BuildLineStarts(text)

	buf := &Buffer{
		text,
		lineStarts,
		0,
	}

	return buf, nil

}

func runeWidth(ch rune) int {
	if ch == '\t' {
		return 4
	} else {
		return runewidth.RuneWidth(ch)
	}
}

func main() {
	if len(os.Args) < 2 {
		fmt.Printf("Usage: %v FILE\n", os.Args[0])
		os.Exit(1)
	}

	filename := os.Args[1]
	file, err := os.Open(filename)
	if err != nil {
		log.Fatalln(err)
	}

	buf, err := NewBuffer(file)
	if err != nil {
		log.Fatalln(err)
	}

	err = termbox.Init()
	if err != nil {
		log.Fatalln(err)
	}
	defer termbox.Close()

	cursor := NewCursor()

mainloop:
	for {

		_, height := termbox.Size()

		termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
		buf.Display(height)
		DisplayCursor(cursor, buf)

		termbox.Flush()

		switch ev := termbox.PollEvent(); ev.Type {
		case termbox.EventKey:
			if ev.Ch != 0 {
				buf.InsertRune(ev.Ch, cursor.index)
				cursor.index += 1
			} else {
				switch ev.Key {
				case termbox.KeyEsc:
					break mainloop
				case termbox.KeyArrowRight:
					MoveCursorRight(cursor, buf)
				case termbox.KeyArrowLeft:
					MoveCursorLeft(cursor, buf)
				case termbox.KeyArrowDown:
					MoveCursorDown(cursor, buf, height)
				case termbox.KeyArrowUp:
					MoveCursorUp(cursor, buf)
				case termbox.KeyEnter:
					buf.InsertRune('\n', cursor.index)
					cursor.index += 1
				case termbox.KeySpace:
					buf.InsertRune(' ', cursor.index)
					cursor.index += 1
				case termbox.KeyBackspace2:
					buf.RemoveRune(cursor.index - 1)
					if cursor.index > 0 {
						cursor.index -= 1
					}
				case termbox.KeyTab:
					buf.InsertRune('\t', cursor.index)
					cursor.index += 1
				case termbox.KeyCtrlW:
					err := os.WriteFile(filename, buf.JustText(), 0666)
					if err != nil {
						log.Fatal(err)
					}
				}

			}
		case termbox.EventError:
			log.Fatalln(ev.Err)
		}

	}
}
