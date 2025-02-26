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
	session.cursor.display()
}

func buildEditSession(filename string) (editSession EditSession, err error) {
	file, err := os.Open(filename)
	if err != nil {
		return
	}
	defer file.Close()

	file_reader := bufio.NewReader(file)
	buf := gapbuf.NewGapBuffer()

	for {
		ch, _, ioErr := file_reader.ReadRune()
		if ioErr != nil {
			if ioErr == io.EOF {
				break
			} else {
				err = fmt.Errorf("Error reading file: %w", ioErr)
				return
			}
		}
		buf.Push(ch)
	}

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

	cursor := newCursor(numCells[0])

	editSession = EditSession{
		buf,
		filename,
		numCells,
		cursor,
	}
	return
}

func (session *EditSession) moveCursorRight() {
	session.cursor = moveRight(session.cursor, &session.buf)
}

func (session *EditSession) moveCursorLeft() {
	session.cursor = moveLeft(session.cursor, &session.buf)
}

func moveLeft(cursor Cursor, buf *gapbuf.GapBuffer) Cursor {
	if cursor.x == 0 {
		return cursor
	}

	var newIdx int
	if !cursor.valid {
		newIdx = cursor.idx
	} else {
		newIdx = cursor.idx - 1
	}

	newX := cursor.x - runeWidth(buf.Get(newIdx))
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

func (cursor *Cursor) display() {
	termbox.SetCursor(cursor.x, cursor.y)
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

	session, err := buildEditSession(os.Args[1])
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
			}
			// TODO: move up and move down
		}
		termbox.Clear(termbox.ColorWhite, termbox.ColorDefault)
		session.display()
		termbox.Flush()
	}
}
