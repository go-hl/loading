package main

import (
	"loading/loading"
	"time"
)

func main() {
	foo := []string{"0", "1", "2", "3", "4", "5", "6", "7", "8", "9", "a", "b", "c", "d", "e", "f"}
	// foo := []string{"0", "1", "2", "3", "4", "5", "6", "7", "8", "9", "a", "b", "c", "d", "e", "f", "0", "1", "2", "3", "4", "5", "6", "7", "8", "9", "a", "b", "c", "d", "e", "f"}
	bar := loading.NewBar(len(foo))

	for index, value := range foo {
		println(index+1, value)
		bar.Render()
		time.Sleep(time.Millisecond * 100)
	}
}
