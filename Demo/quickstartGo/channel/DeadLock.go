package main

import "fmt"

func produce(c chan string) {

	c <- "channelValue"
}

func consume(c chan string) {
	fmt.Println(<-c)
}

func produce2(c chan string, c2 chan string) {

}

// deadLock1  主协程等待通道写又等待通道读
// deadLock2  主携程等待通道1 写入 子携程等待通道2读取/写入
func main() {
	//deadLock1()
}

// 子携程必须先启动
func deadLock1() {
	ch := make(chan string)
	consume(ch)
	produce(ch)
	// 父携程 生产 子携程 消费
	//go consume(ch)
	//produce(ch)
	// 父携程 消费  子携程   生产
	//go produce(ch)
	//consume(ch)
}

//func deadLock2() {
//	ch := make(chan  string)
//	ch2 := make(chan string)
//	//consume(ch)
//	go consume(ch)
//
//	produce()
//}

func deadLock3() {
	chs := make(chan string, 2)
	chs <- "first"
	chs <- "second"

	for ch := range chs {
		fmt.Println(ch)
	}
}
