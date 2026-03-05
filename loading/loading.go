package loading

import (
	"fmt"
	"log"
	"os"
	"strings"

	"golang.org/x/term"
)

const (
	ansiESC = "\x1B"
	ansiCSI = ansiESC + "["
)

type bar struct {
	stepsTotal int
	stepsCount int

	termCols int
	termRows int
}

func NewBar(steps int) *bar {
	return &bar{
		stepsTotal: steps,
		stepsCount: 1,
	}
}

func (b bar) percentage() int {
	return (b.stepsCount * 100) / b.stepsTotal
}

func (b *bar) termSizeUpdate() {
	fileDescriptor := int(os.Stdin.Fd())
	cols, rows, err := term.GetSize(fileDescriptor)
	if err != nil {
		log.Println("error geting terminal size:", err)
		return
	}
	b.termCols = cols
	b.termRows = rows
}

func (b bar) print() {
	const layout = len("[] 000%")

	chars := map[bool]int{true: b.termCols - layout}[b.termCols >= layout]
	repeat := (chars * b.percentage()) / 100

	fmt.Printf("[%s%s] %3d%%", strings.Repeat("#", repeat), strings.Repeat(".", chars-repeat), b.percentage())
}

func (b *bar) Render() {
	b.termSizeUpdate()

	if b.stepsCount < b.stepsTotal {
		ansiNewLine()
		ansiCursorUp()
	}

	ansiCursorSave()
	ansiClearNexts()
	ansiCursorEnd(b)

	b.print()

	if b.stepsCount <= b.stepsTotal {
		ansiCursorRestore()
	}

	if b.stepsCount >= b.stepsTotal {
		ansiClearNexts()
	}

	b.stepsCount++
}

func ansiNewLine() {
	fmt.Print("\n") // create new line
}

func ansiCursorUp() {
	fmt.Print(ansiCSI + "A") // move cursor up
}

func ansiCursorEnd(b *bar) {
	fmt.Printf(ansiCSI+"%d;H", b.termRows) // move cursor to down
}

func ansiCursorSave() {
	fmt.Print(ansiESC + "7") // save current cursor position
}

func ansiCursorRestore() {
	fmt.Print(ansiESC + "8") // restore cursor to saved position
}

func ansiClearNexts() {
	fmt.Print(ansiCSI + "J") // clear from cursor all next lines
}
