package loading

import (
	"fmt"
	"sync/atomic"
)

const (
	ansiESC = "\x1B"
	ansiCSI = ansiESC + "["
)

func ansiNewLine() {
	fmt.Print("\n") // create new line
}

func ansiCursorUp() {
	fmt.Print(ansiCSI + "A") // move cursor up
}

func ansiCursorEnd(b *Bar) {
	fmt.Printf(ansiCSI+"%d;H", atomic.LoadInt64(&b.termRows)) // move cursor to down
}

func ansiCursorSave() {
	fmt.Print(ansiESC + "7") // save current cursor position
}

func ansiCursorRestore() {
	fmt.Print(ansiESC + "8") // restore cursor to saved position
}

func ansiClearLine() {
	fmt.Print(ansiCSI + "2K") // clear cursor current line
}

func ansiClearNexts() {
	fmt.Print(ansiCSI + "J") // clear from cursor all next lines
}
