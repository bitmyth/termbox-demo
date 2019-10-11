package main

import (
	"bufio"
	"flag"
	"github.com/nsf/termbox-go"
	"log"
	"os"
	"os/exec"
	"unicode/utf8"
)

const edit_box_width = 30

type IBox struct {
	text          [][]byte
	filteredText  [][]byte
	scrollY       int
	cursorXOffset int
	cursorYOffset int
	width         int
	height        int
}

const coldef = termbox.ColorDefault

var file string

func main() {
	flag.StringVar(&file, "f", "/Users/gsh/.bash_profile", "file path")
	flag.Parse()

	err := termbox.Init()
	if err != nil {
		panic(err)
	}
	defer termbox.Close()
	termbox.SetInputMode(termbox.InputEsc)

	termbox.Clear(coldef, coldef)

	w, h := termbox.Size()

	box := IBox{width: w, height: h}

	file, err := os.Open(file)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		box.text = append(box.text, scanner.Bytes())
		//fmt.Println(scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	box.refresh()
	for {
		switch ev := termbox.PollEvent(); ev.Type {
		case termbox.EventKey:
			switch ev.Key {
			case termbox.KeyEsc:
				return

			case termbox.KeyCtrlK:
				box.moveCursorUp()

			case termbox.KeyCtrlJ:
				// hit the bottom of whole document
				if box.scrollY+box.cursorYOffset == len(box.text)-1 {
					break
				}
				// hit the bottom of screen
				box.moveCursorDown()

			case termbox.KeyCtrlF:
				box.pageDown()

			case termbox.KeyCtrlB:
				box.pageUp()

			case termbox.KeyEnter:
				box.executeHook()
			}
		}
		box.refresh()
	}
}

func (box *IBox) executeHook() {
	selected := string(box.text[box.scrollY+box.cursorYOffset])
	//firstAppear := strings.Index(selected, "=")
	cmd := exec.Command("open", selected)
	_ = cmd.Run()
	//if firstAppear > 0 {
	//	println(selected[firstAppear+1:])
	//cmd := exec.Command("bash", "-c", selected[firstAppear+1:])
	//cmd := exec.Command("open", selected)
	//buf, err := cmd.Output()
	//println(string(buf))
	//if err != nil {
	//	println(err.Error())
	//}
	//}
}

func (box *IBox) pageUp() {
	scrollY := box.scrollY - box.height
	if scrollY < 0 {
		scrollY = 0
		box.cursorYOffset = 0
	}
	box.scrollY = scrollY
}

func (box *IBox) pageDown() {
	scrollY := box.scrollY + box.height
	if scrollY >= len(box.text) {
		scrollY = box.scrollY
		box.cursorYOffset = (len(box.text)-scrollY)%box.height - 1
	}
	box.scrollY = scrollY
}

func (box *IBox) moveCursorUp() {
	if box.cursorYOffset > 0 {
		box.cursorYOffset--
	} else if box.cursorYOffset == 0 && box.scrollY > 0 {
		box.scrollY--
	}
}

func (box *IBox) moveCursorDown() {
	if box.cursorYOffset > box.height-2 {
		box.scrollY++
	} else {
		box.cursorYOffset++
	}
}

func (box IBox) refresh() {
	y := 0

	// loop through all lines
	for index, line := range box.text[box.scrollY:] {
		bg := termbox.ColorDefault
		fg := termbox.ColorDefault

		if box.cursorYOffset == index {
			bg = termbox.ColorBlue
			fg = termbox.ColorWhite
		}

		text := line

		// print line
		x := 0
		for {
			if x == len(line) {
				if x < box.width {
					for remain := box.width - x; remain > 0; remain-- {
						termbox.SetCell(x, y, ' ', fg, bg)
						x++
					}
				}
				break
			}

			r, size := utf8.DecodeRune(text)

			termbox.SetCell(x, y, r, fg, bg)

			x++
			text = text[size:]
		}

		y++
	}

	if y < box.height {
		for remain := box.height - y; remain > 0; remain-- {
			for x := 0; x < box.width; x++ {
				termbox.SetCell(x, y, ' ', termbox.ColorDefault, termbox.ColorDefault)
			}
			y++
		}
	}

	termbox.Flush()
}
