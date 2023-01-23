package main

import (
	"fmt"
	"runtime"
)

func squares(c chan int) {
	for i := 0; i < 4; i++ {
		num := <-c
		fmt.Println(num * num)
	}
}

func main() {

	c := make(chan int)
	go squares(c)

	fmt.Print(runtime.NumGoroutine())

	c <- 1
	c <- 2
	c <- 3
	c <- 4

	fmt.Print(runtime.NumGoroutine())
}
