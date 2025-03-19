package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/mattn/go-runewidth"
	"github.com/michael-disalvo/gapbuf"
	"github.com/nsf/termbox-go"
)

const cTabWidth int = 4

type Logger struct {
	file *os.File
}

func newLogger(filePath string) (*Logger, error) {
	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, err
	}
	return &Logger{file: file}, nil
}

func (l *Logger) logMessage(message string) error {
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	logEntry := fmt.Sprintf("%s: %s\n", timestamp, message)

	_, err := l.file.WriteString(logEntry)
	return err
}

func (l *Logger) Close() error {
	return l.file.Close()
}

func runeWidth(ch rune) int {
	switch ch {
	case '\t':
		return cTabWidth
	default:
		return runewidth.RuneWidth(ch)
	}
}

func maxInt(a, b int) int {
	if a > b {
		return a
	} else {
		return b
	}
}

type EditSession struct {
	buf      gapbuf.GapBuffer // the actual text backing
	filename string           // name of the file we are editing
	numCells []int            // the number of used cells in each line
	cursor   Cursor           // the cursor
}

func (session *EditSession) display() {
	x := 0
	y := 0
	for i := 0; i < session.buf.Len(); i++ {
		ch := session.buf.Get(i)
		if ch == '\n' {
			y += 1
			x = 0
			continue
		}
		termbox.SetCell(x, y, ch, termbox.ColorWhite, termbox.ColorDefault)
		x += runeWidth(ch)
	}
	session.cursor.display(session.numCells)
}

func gapBufFromReader(reader io.Reader) (gapbuf.GapBuffer, error) {
	bufReader := bufio.NewReader(reader)
	buf := gapbuf.NewGapBuffer()
	for {
		ch, _, ioErr := bufReader.ReadRune()
		if ioErr != nil {
			if ioErr == io.EOF {
				break
			} else {
				return buf, fmt.Errorf("Error reading rune: %w", ioErr)
			}
		}
		buf.Push(ch)
	}
	return buf, nil
}

func numCellsFromGapBuf(buf *gapbuf.GapBuffer) []int {
	numCells := make([]int, 0)
	currLineCells := 0
	for i := 0; i < buf.Len(); i++ {
		ch := buf.Get(i)
		if ch == '\n' {
			numCells = append(numCells, currLineCells)
			currLineCells = 0
		} else {
			currLineCells += runeWidth(ch)
		}
	}
	numCells = append(numCells, currLineCells)
	return numCells
}

func editSessionFromFile(filename string) (editSession EditSession, err error) {
	file, err := os.Open(filename)
	if err != nil {
		return
	}
	defer file.Close()
	return buildEditSession(file, filename)
}

func buildEditSession(rdr io.Reader, filename string) (editSession EditSession, err error) {
	buf, err := gapBufFromReader(rdr)
	if err != nil {
		return
	}

	numCells := numCellsFromGapBuf(&buf)

	cursor := newCursor(numCells[0])

	editSession = EditSession{
		buf,
		filename,
		numCells,
		cursor,
	}
	return
}

func (session *EditSession) numLines() int {
	return len(session.numCells)
}

func (session *EditSession) usedCellsInLine(line int) int {
	if line < 0 || line >= session.numLines() {
		return 0
	}
	return session.numCells[line]
}

func (session *EditSession) moveCursorRight() {
	session.cursor = moveRight(session.cursor, &session.buf)
}

func (session *EditSession) moveCursorLeft() {
	session.cursor = moveLeft(session.cursor, &session.buf, session.numCells)
}

func moveLeft(cursor Cursor, buf *gapbuf.GapBuffer, numCells []int) Cursor {
	if cursor.x == 0 {
		return cursor
	}

	var newX int
	var newIdx int
	if !cursor.valid {
		newIdx = cursor.idx
		newX = numCells[cursor.y] - (runeWidth(buf.Get(newIdx)))
	} else {
		newIdx = cursor.idx - 1
		newX = cursor.x - runeWidth(buf.Get(newIdx))
	}

	newValid := true
	return Cursor{
		x:     newX,
		y:     cursor.y,
		valid: newValid,
		idx:   newIdx,
	}
}

func moveRight(cursor Cursor, buf *gapbuf.GapBuffer) Cursor {
	if !cursor.valid {
		return cursor
	}

	ch := buf.Get(cursor.idx)
	newX := cursor.x + runeWidth(ch)
	newValid := cursor.idx+1 < buf.Len() && buf.Get(cursor.idx+1) != '\n'
	newIdx := cursor.idx
	if newValid {
		newIdx += 1
	}

	return Cursor{
		x:     newX,
		y:     cursor.y,
		idx:   newIdx,
		valid: newValid,
	}
}

func (session *EditSession) moveCursorDown() {
	cursor := session.cursor
	if cursor.y == session.numLines()-1 {
		return
	}

	newY := cursor.y + 1

	usedCells := session.usedCellsInLine(newY)
	newValid := cursor.x < usedCells

	// idx to next \n, not including current character
	idx := cursor.idx + 1
	for {
		if idx >= session.buf.Len() {
			break
		}
		if session.buf.Get(idx) == '\n' {
			break
		}
	}

	var newIdx int
	if usedCells == 0 {
		newIdx = idx
	} else if !newValid {
		// idx to ch before next \n, or end of buffer
		idx += 1
		for {
			if idx == session.buf.Len()-1 {
				newIdx = idx
				break
			}
			if session.buf.Get(idx+1) == '\n' {
				newIdx = idx
				break
			}
			idx += 1
		}
	} else {
		// idx to ch s.t. currX + runeWidth(ch) > cursor.x
		idx += 1
		currX := 0
		for {
			nextWidth := runeWidth(session.buf.Get(idx))
			if nextWidth+currX > cursor.x {
				newIdx = idx
				break
			}
			currX += nextWidth
			idx += 1
		}
	}

	newX := cursor.x

	newCursor := Cursor{
		x:     newX,
		y:     newY,
		valid: newValid,
		idx:   newIdx,
	}
	session.cursor = newCursor

}

func minInt(i1, i2 int) int {
	if i1 < i2 {
		return i1
	} else {
		return i2
	}
}

type Cursor struct {
	x     int
	y     int
	idx   int
	valid bool
}

func newCursor(numCellsInFirstLine int) Cursor {
	valid := numCellsInFirstLine > 0
	return Cursor{valid: valid}
}

func (cursor *Cursor) display(usedCells []int) {
	visualX := minInt(usedCells[cursor.y], cursor.x)
	termbox.SetCursor(visualX, cursor.y)
}

func main() {
	if len(os.Args) < 2 {
		fmt.Printf("Usage: %s <file>\n", os.Args[0])
		os.Exit(1)
	}

	logger, err := newLogger("goedit.out")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer logger.Close()

	session, err := editSessionFromFile(os.Args[1])
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	err = termbox.Init()
	if err != nil {
		fmt.Printf("Error initializing termbox: %v\n", err)
		os.Exit(1)
	}
	defer termbox.Close()

	session.display()
	termbox.Flush()

mainloop:
	for {
		switch ev := termbox.PollEvent(); ev.Type {
		case termbox.EventKey:
			switch ev.Key {
			case termbox.KeyEsc:
				break mainloop
			case termbox.KeyArrowRight:
				session.moveCursorRight()
				logger.logMessage(fmt.Sprintf("cursor at: %v", session.cursor))
			case termbox.KeyArrowLeft:
				session.moveCursorLeft()
				logger.logMessage(fmt.Sprintf("cursor at: %v", session.cursor))
			case termbox.KeyArrowDown:
				session.moveCursorDown()
				logger.logMessage(fmt.Sprintf("cursor at :%v", session.cursor))
			}
			// TODO: move up
		}
		termbox.Clear(termbox.ColorWhite, termbox.ColorDefault)
		session.display()
		termbox.Flush()
	}
}
