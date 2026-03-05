package main

import (
	"fmt"
	"loading/loading"
	"math/rand/v2"
	"sync"
	"time"

	"github.com/google/uuid"
)

func random() time.Duration {
	return time.Duration(rand.IntN(10) + 1)
}

func semConcorrencia() {
	const count = 10000
	bar := loading.NewBar(count)

	bar.Render()
	for index := range count {
		fmt.Println(index, uuid.NewString())
		bar.Step(1)
		time.Sleep(time.Millisecond)
	}
	bar.Done()

	fmt.Println("hello world")
}

func comConcorrencia() {
	const count = 10000
	bar := loading.NewBar(count)

	bar.Render()
	var wg sync.WaitGroup
	for index := range count {
		wg.Add(1)
		go func() {
			defer wg.Done()
			fmt.Println(index, uuid.NewString())
			time.Sleep(time.Second * random())
			bar.Step(1)
		}()
	}
	wg.Wait()
	bar.Done()

	fmt.Println("hello world")
}

func comConcorrenciaWorkers() {
	const count = 10000
	bar := loading.NewBar(count)
	data := make(chan string, count)

	bar.Render()
	var wg sync.WaitGroup
	for range 1000 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for data := range data {
				fmt.Println(data)
				time.Sleep(time.Second * random())
				bar.Step(1)
			}
		}()
	}

	for index := range count {
		data <- fmt.Sprint(index, uuid.NewString())
	}
	close(data)

	wg.Wait()
	bar.Done()

	fmt.Println("hello world")
}

func main() {
	// semConcorrencia()
	// comConcorrencia()
	// comConcorrenciaWorkers()
}
