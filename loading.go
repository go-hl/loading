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
		check:      make(chan int, steps),
		done:       make(chan struct{}, 1),
	}
}

func (b *Bar) termSizeUpdate() {
	fileDescriptor := int(os.Stdin.Fd())
	cols, rows, err := term.GetSize(fileDescriptor)
	if err != nil {
		atomic.SwapInt64(&b.termCols, 0)
		atomic.SwapInt64(&b.termRows, 0)
		return
	}

	atomic.SwapInt64(&b.termCols, int64(cols))
	atomic.SwapInt64(&b.termRows, int64(rows))
}

func (b *Bar) percentage() int {
	count := int(atomic.LoadInt64(&b.stepsCount))
	percentage := (count * 100) / int(b.stepsTotal)
	overflow := percentage > 100

	if overflow {
		b.stop()
		return 100
	}

	return percentage
}

func (b *Bar) print() {
	count := int(atomic.LoadInt64(&b.stepsCount))
	length := len(strconv.Itoa(int(b.stepsTotal)))
	marker := strings.Repeat("0", length)
	layout := len(fmt.Sprintf("[] %s/%s 000%%", marker, marker))

	cols := int(atomic.LoadInt64(&b.termCols))
	chars := map[bool]int{true: cols - layout}[cols >= layout]
	repeat := (chars * b.percentage()) / 100

	fmt.Printf(
		"[%s%s] %*d/%*d %3d%%",
		strings.Repeat("#", repeat), strings.Repeat(".", chars-repeat),
		length, count, length, b.stepsTotal, b.percentage(),
	)
}

func (*Bar) clear() {
	time.Sleep(time.Second)
	ansiClearNexts()
}

func (b *Bar) draw() {
	b.termSizeUpdate()

	ansiNewLine()
	ansiCursorUp()

	ansiCursorSave()
	ansiClearNexts()

	ansiCursorEnd(b)
	ansiClearLine()
	b.print()

	ansiCursorRestore()
}

func (b *Bar) stop() {
	b.quit.Store(true)
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
			log.Println("missing step", atomic.LoadInt64(&b.stepsCount))
		}
	}
}

// Render stats the progress bar showing.
// It is finishd when the count of steps the greater than or equal the total.
// Also can fast cancel with cancel function return value.
func (b *Bar) Render() context.CancelFunc {
	ctx, cancel := context.WithCancel(context.Background())

	b.draw()
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
				b.clear()
				b.stop()
			case count := <-b.check:
				atomic.AddInt64(&b.stepsCount, int64(count))

				stepsCount := atomic.LoadInt64(&b.stepsCount)
				finished := stepsCount >= b.stepsTotal

				b.draw()
				if finished {
					b.clear()
					b.stop()
				}
			}
		}
	}()

	return cancel
}
