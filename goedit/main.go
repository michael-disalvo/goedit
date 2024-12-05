package main

import (
	"bufio"
	"fmt"
	"io"
	"os"

	"github.com/mattn/go-runewidth"
	"github.com/michael-disalvo/gapbuf"
	"github.com/nsf/termbox-go"
)

const cTabWidth int = 4

type EditSession struct {
	buf      gapbuf.GapBuffer // the actual text backing
	lineLens []int            // the length of each line, used to map (x, y) -> logical index
	filename string           // name of the file we are editing
}

func runeWidth(ch rune) int {
	switch ch {
	case '\t':
		return cTabWidth
	default:
		return runewidth.RuneWidth(ch)
	}
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
}

func buildEditSession(filename string) (editSession EditSession, err error) {
	file, err := os.Open(filename)
	if err != nil {
		return
	}
	defer file.Close()

	file_reader := bufio.NewReader(file)
	buf := gapbuf.NewGapBuffer()
	lineLens := make([]int, 0)
	currentLineLen := 0
	for {
		ch, _, ioErr := file_reader.ReadRune()
		if ioErr != nil {
			if ioErr == io.EOF {
				lineLens = append(lineLens, currentLineLen)
				break
			} else {
				err = fmt.Errorf("Error reading file: %w", ioErr)
				return
			}
		}
		currentLineLen += 1
		if ch == '\n' {
			lineLens = append(lineLens, currentLineLen)
			currentLineLen = 0
		}
		buf.Push(ch)
	}

	editSession = EditSession{
		buf,
		lineLens,
		filename,
	}
	return
}

func main() {
	if len(os.Args) < 2 {
		fmt.Printf("Usage: %s <file>\n", os.Args[0])
		os.Exit(1)
	}

	session, err := buildEditSession(os.Args[1])
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	err = termbox.Init()
	if err != nil {
		fmt.Printf("Error initializing termbox: %w\n", err)
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
			}
		}
	}
}
