package main

import "fmt"

type Lamp struct {
}

func (l Lamp) On() {
	println("On")

}
func (l Lamp) Off() {
	println("Off")
}

func worker(ch chan struct{}) {
	// Receive a message from the main program.
	<-ch
	println("roger")

	// Send a message to the main program.
	close(ch)
}
func main() {
	set := make(map[string]struct{})
	for _, value := range []string{"apple", "orange", "apple"} {
		set[value] = struct{}{}
	}
	fmt.Println(set)

	Lamp{}.On()
	ch := make(chan struct{})
	go worker(ch)

	// Send a message to a worker.
	ch <- struct{}{}

	// Receive a message from the worker.
	<-ch
	println("log")
}

// 递归 回溯 调用 最后调用main包的init
func init() {
	println("hey")
}
