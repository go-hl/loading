package loading

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
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

	mu  sync.Mutex
	out io.Writer

	length int
	layout int
}

// NewBar creates a new [Bar].
func NewBar(steps int64) *Bar {
	length := len(strconv.Itoa(int(steps)))
	marker := strings.Repeat("0", length)
	layout := len(fmt.Sprintf("000%% [] %s/%s", marker, marker))

	return &Bar{
		stepsTotal: steps,
		check:      make(chan int, steps),
		done:       make(chan struct{}, 1),
		out:        os.Stdout,
		length:     length,
		layout:     layout,
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

	if overflow := percentage > 100; overflow {
		b.stop()
		return 100
	}
	return percentage
}

func (b *Bar) clear() {
	time.Sleep(time.Second)

	b.mu.Lock()
	defer b.mu.Unlock()

	ansiCursorEnd(b.out, b)
	ansiClearLine(b.out)
}

func (b *Bar) print(w io.Writer) {
	cols := int(atomic.LoadInt64(&b.termCols))
	chars := map[bool]int{true: cols - b.layout}[cols >= b.layout]
	repeat := (chars * b.percentage()) / 100
	count := int(atomic.LoadInt64(&b.stepsCount))

	fmt.Fprintf(
		w, "%3d%% [%s%s] %*d/%*d", b.percentage(),
		strings.Repeat("#", repeat), strings.Repeat(".", chars-repeat),
		b.length, count, b.length, b.stepsTotal,
	)
}

func (b *Bar) draw(w io.Writer) {
	var buf bytes.Buffer

	b.termSizeUpdate()
	ansiCursorSave(&buf)
	ansiCursorEnd(&buf, b)
	ansiClearLine(&buf)
	b.print(&buf)
	ansiCursorRestore(&buf)

	_, err := w.Write(buf.Bytes())
	if err != nil {
		log.Println("not rendering:", err)
	}
}

func (b *Bar) display() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.draw(b.out)
}

func (b *Bar) stop() {
	b.quit.Store(true)
}

// Done must be called after call [Bar.Render] to wait [Bar] completion.
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

	b.display()
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

				b.display()
				if finished := stepsCount >= b.stepsTotal; finished {
					b.clear()
					b.stop()
				}
			}
		}
	}()

	return cancel
}

// Writer returns an [io.Writer] synchronized with the bar to use with
// other print methods like [fmt.Fprint] (and your variants) or [log.New].
// Every write will erase the bar, print the content, then redraw the bar,
// keeping the bar pinned to the last line at all times.
func (b *Bar) Writer() io.Writer {
	return &writer{
		bar: b,
	}
}
