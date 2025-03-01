package main

import (
	"fmt"
	"log"
	"sync"
	"time"
)

/*
通过waitGroup 来实现channel等待

单条携程 输入输出严格有序
*/
type SingleActionX func(ch chan interface{})
type SingleWaitGroupWorker interface {
	do()
}
type SingleWaitGroupWorkerImpl struct {
	ch   chan interface{}
	wg   *sync.WaitGroup
	task SingleActionX
}

// 不实现do 方法 通过task 动态调用方法
func (receiver *SingleWaitGroupWorkerImpl) do() {

	if receiver.task != nil {
		receiver.task(receiver.ch)

	} else {
		log.Print("not implement!")
	}

}

var wg sync.WaitGroup

func main() {
	ch := make(chan interface{})
	singleWaitGroupWorker := SingleWaitGroupWorkerImpl{
		ch: ch,
		wg: &wg,
		task: func(ch chan interface{}) {
			go func() {
				// 单条携程 严格有序 可以用range 替换
				for true {
					if s := <-ch; s != nil {
						wg.Done()
						fmt.Print(s)
						time.Sleep(100)
					}
				}
			}()
		},
	}
	wg.Add(10)
	singleWaitGroupWorker.do()
	for i := 0; i < 10; i++ {
		singleWaitGroupWorker.ch <- i

	}
	wg.Wait()
	close(ch)
}
