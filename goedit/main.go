package main

import (
	"fmt"
	"os"
	"time"

	runewidth "github.com/mattn/go-runewidth"
	termbox "github.com/nsf/termbox-go"
)

func tbprint(x, y int, msg string) {
	for _, c := range msg {
		termbox.SetCell(x, y, c, termbox.ColorRed, termbox.ColorDefault)
		x += runewidth.RuneWidth(c)
	}
}

func main() {
	err := termbox.Init()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	tbprint(2, 2, "Hello Termbox!")
	termbox.Flush()

	time.Sleep(5 * time.Second)
	termbox.Close()

}
