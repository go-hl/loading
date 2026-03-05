package loading

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"golang.org/x/term"
)

// Bar represents a progress bar.
type Bar struct {
	stepsTotal int64
	stepsCount int64

	termCols int64
	termRows int64

	quit  atomic.Bool
	check chan int
	done  chan struct{}
}

// NewBar creates a new [Bar].
func NewBar(steps int64) *Bar {
	return &Bar{
		stepsTotal: steps,
		stepsCount: 1,
		check:      make(chan int, 1),
		done:       make(chan struct{}, 1),
	}
}

func (b *Bar) termSizeUpdate() {
	fileDescriptor := int(os.Stdin.Fd())
	cols, rows, err := term.GetSize(fileDescriptor)
	if err != nil {
		log.Println("error geting terminal size:", err)
		return
	}

	atomic.SwapInt64(&b.termCols, int64(cols))
	atomic.SwapInt64(&b.termRows, int64(rows))
}

func (b *Bar) percentage() int {
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

func (b *Bar) print() {
	count := int(atomic.LoadInt64(&b.stepsCount))
	total := int(atomic.LoadInt64(&b.stepsTotal))
	length := len(strconv.Itoa(total))
	marker := strings.Repeat("0", length)
	layout := len(fmt.Sprintf("[] %s/%s 000%%", marker, marker))

	cols := int(atomic.LoadInt64(&b.termCols))
	chars := map[bool]int{true: cols - layout}[cols >= layout]
	repeat := (chars * b.percentage()) / 100

	fmt.Printf(
		"[%s%s] %*d/%*d %3d%%",
		strings.Repeat("#", repeat), strings.Repeat(".", chars-repeat),
		length, count, length, total, b.percentage(),
	)
}

func (*Bar) clear() {
	time.Sleep(time.Millisecond * 500)
	ansiClearNexts()
}

func (b *Bar) draw() {
	b.termSizeUpdate()

	ansiNewLine()
	ansiCursorUp()

	ansiCursorSave()
	ansiClearNexts()

	ansiCursorEnd(b)
	b.print()

	ansiCursorRestore()
}

// Done simples call this after call [Bar.Render] to wait [Bar] completion.
func (b *Bar) Done() {
	<-b.done
}

// Step receives the count of steps to progress, one or more per time.
func (b *Bar) Step(count int) {
	if !b.quit.Load() {
		select {
		case b.check <- count:
		default:
		}
	}
}

// Render stats the progress bar showing.
// It is finishd when the count of steps the greater than or equal the total.
// Also can fast cancel with cancel function return value.
func (b *Bar) Render() context.CancelFunc {
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
