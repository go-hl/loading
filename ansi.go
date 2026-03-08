package loading

import (
	"fmt"
	"io"
	"sync/atomic"
)

const (
	ansiESC = "\x1B"
	ansiCSI = ansiESC + "["
)

func ansiCursorEnd(w io.Writer, b *Bar) {
	fmt.Fprintf(w, ansiCSI+"%d;H", atomic.LoadInt64(&b.termRows)) // move cursor to down
}

func ansiCursorSave(w io.Writer) {
	fmt.Fprint(w, ansiESC+"7") // save current cursor position
}

func ansiCursorRestore(w io.Writer) {
	fmt.Fprint(w, ansiESC+"8") // restore cursor to saved position
}

func ansiClearLine(w io.Writer) {
	fmt.Fprint(w, ansiCSI+"2K") // clear cursor current line
}
