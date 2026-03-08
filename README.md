# Loading

The package Loading provides a CLI loading bar that increments for each step.

## Usage

Procedural:
```go
package main

import (
	"fmt"

	"github.com/go-hl/loading"
)

const count = 1000

func main() {
	bar := loading.NewBar(count)
	writer := bar.Writer()

	bar.Render()
	for index := range count {
		fast(writer, index)
		bar.Step(1)
	}
	bar.Done()

	fmt.Println("hello world")
}
```

Concurrent:
```go
package main

import (
	"fmt"
	"sync"

	"github.com/go-hl/loading"
)

const count = 1000

func main() {
	bar := loading.NewBar(count)
	writer := bar.Writer()

	bar.Render()
	var wg sync.WaitGroup
	for index := range count {
		wg.Add(1)
		go func() {
			defer wg.Done()
			slow(writer, index)
			bar.Step(1)
		}()
	}
	wg.Wait()
	bar.Done()

	fmt.Println("hello world")
}
```

Parallelism:
```go
package main

import (
	"fmt"
	"sync"

	"github.com/go-hl/loading"
)

const count = 1000

func main() {
	data := make(chan int, count)
	bar := loading.NewBar(count)
	writer := bar.Writer()

	bar.Render()
	var wg sync.WaitGroup
	for range int(count * .5) {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for index := range data {
				slow(writer, index)
				bar.Step(1)
			}
		}()
	}

	for index := range count {
		data <- index
	}
	close(data)

	wg.Wait()
	bar.Done()

	fmt.Println("hello world")
}
```

Utilities:
```go
package main

import (
	"fmt"
	"io"
	"math/rand/v2"
	"time"

	"github.com/google/uuid"
)

func random() int {
	return rand.IntN(10) + 1
}

func half() int {
	return (rand.IntN(10) + 1) / 2
}

func duration() time.Duration {
	return time.Duration(random())
}

func fast(w io.Writer, index int) {
	for range random() {
		time.Sleep(time.Millisecond)
	}
	fmt.Fprintln(w, index, uuid.NewString())
}

func slow(w io.Writer, index int) {
	for range half() {
		time.Sleep(time.Second * duration())
	}
	fmt.Fprintln(w, index, uuid.NewString())
}
```
