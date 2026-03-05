package loading

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"sync/atomic"
	"time"

	"golang.org/x/term"
)

const (
	ansiESC = "\x1B"
	ansiCSI = ansiESC + "["
)

type bar struct {
	stepsTotal int64
	stepsCount int64

	termCols int64
	termRows int64

	quit  atomic.Bool
	check chan int
	done  chan struct{}
}

func NewBar(steps int64) *bar {
	return &bar{
		stepsTotal: steps,
		stepsCount: 1,
		check:      make(chan int, 1),
		done:       make(chan struct{}, 1),
	}
}

func (b *bar) print() {
	const layout = len("[] 000%")

	cols := int(atomic.LoadInt64(&b.termCols))
	chars := map[bool]int{true: cols - layout}[cols >= layout]
	repeat := (chars * b.percentage()) / 100

	fmt.Printf("[%s%s] %3d%%", strings.Repeat("#", repeat), strings.Repeat(".", chars-repeat), b.percentage())
}

func (b *bar) termSizeUpdate() {
	fileDescriptor := int(os.Stdin.Fd())
	cols, rows, err := term.GetSize(fileDescriptor)
	if err != nil {
		log.Println("error geting terminal size:", err)
		return
	}

	atomic.SwapInt64(&b.termCols, int64(cols))
	atomic.SwapInt64(&b.termRows, int64(rows))
}

func (b *bar) percentage() int {
	count := int(atomic.LoadInt64(&b.stepsCount))
	total := int(atomic.LoadInt64(&b.stepsTotal))
	percentage := (count * 100) / total
	overflow := percentage > 100

	if overflow {
		b.quit.Store(true)
		return 100
	}

	return percentage
}

func (*bar) clear() {
	time.Sleep(time.Millisecond * 500)
	ansiClearNexts()
}

func (b *bar) draw() {
	b.termSizeUpdate()

	ansiNewLine()
	ansiCursorUp()

	ansiCursorSave()
	ansiClearNexts()

	ansiCursorEnd(b)
	b.print()

	ansiCursorRestore()
}

func (b *bar) Done() {
	<-b.done
}

func (b *bar) Step(count int) {
	if !b.quit.Load() {
		select {
		case b.check <- count:
		default:
		}
	}
}

func (b *bar) Render() context.CancelFunc {
	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		for {
			if b.quit.Load() {
				b.done <- struct{}{}
				close(b.check)
				close(b.done)
				return
			}
			select {
			case <-ctx.Done():
				ansiCursorRestore()
				ansiClearNexts()
				b.quit.Store(true)
			case count := <-b.check:
				stepsCount := atomic.LoadInt64(&b.stepsCount)
				stepsTotal := atomic.LoadInt64(&b.stepsTotal)
				finished := stepsCount >= stepsTotal

				b.draw()
				if finished {
					b.clear()
					b.quit.Store(true)
					break
				}

				atomic.AddInt64(&b.stepsCount, int64(count))
			}
		}
	}()

	return cancel
}

func ansiNewLine() {
	fmt.Print("\n") // create new line
}

func ansiCursorUp() {
	fmt.Print(ansiCSI + "A") // move cursor up
}

func ansiCursorEnd(b *bar) {
	fmt.Printf(ansiCSI+"%d;H", atomic.LoadInt64(&b.termRows)) // move cursor to down
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
