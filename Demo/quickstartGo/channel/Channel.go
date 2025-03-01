package main

import (
	"fmt"
	"log"
	"time"
)

/**
主要练习 channel的创建与使用

SingleWorker 通过一条channel进行通信
1。 如果没有缓冲区的channel 如果新数据到来就旧数据没有取出就会DeadLock
MultiWorker 通过channel数组进行通信

range 判断父routine的关闭事件

缺少buffer channel的使用样例

等待携程结束是通过 time sleep 完成的 待改善 --》 使用 *waitGroup
		-》 add 添加任务书， wait 任务 done 结束任务
4.26
*/

type SingleWorker interface {
	do()
}
type SingleAction func(ch chan interface{})
type SingleWorkerImpl struct {
	ch   chan interface{}
	task SingleAction
}

// 不实现do 方法 通过task 动态调用方法
func (receiver *SingleWorkerImpl) do() {
	if receiver.task != nil {
		receiver.task(receiver.ch)
	} else {
		log.Print("not implement!")
	}
}

type MultiWorker interface {
	doForAll()
}

type MultiAction func(ch []chan interface{})

type MultiWorkerImpl struct {
	chs   []chan interface{}
	tasks MultiAction
}

// 不实现do 方法 通过task 动态调用方法
func (receiver *MultiWorkerImpl) doForAll() {
	if receiver.tasks != nil {
		receiver.tasks(receiver.chs)
	} else {
		log.Print("not implement!")
	}
}
func main() {

	singleWorkerChannel := make(chan interface{})
	singleWorker := SingleWorkerImpl{
		// 规定work与producer之间的协程
		ch: singleWorkerChannel,
		// 规定worker 要做的事情
		task: func(ch chan interface{}) {
			go func() {
				// 没有新数据将会自动阻塞
				//for value := range ch {
				//	fmt.Println("test single worker",value)
				//}
				for true {
					if value := <-ch; true {
						fmt.Println(value)
					}
				}
			}()
		},
	}

	singleWorker.do()

	// 思考 如何让其有序的打印 （一条新的routine）
	singleWorkerChannel <- 2
	singleWorkerChannel <- 3
	singleWorkerChannel <- 4

	//close(singleWorkerChannel)

	multiWorkerChannel := make([]chan interface{}, 10)
	// 固定数量
	//var multiWorkerChannel [] chan interface{}

	for i := 0; i < 10; i++ {
		multiWorkerChannel[i] = make(chan interface{})
	}

	multiWorker := MultiWorkerImpl{
		chs: multiWorkerChannel,
		tasks: func(ch []chan interface{}) {
			go func() {
				// 识别close 后的事件
				for _, value := range ch {

					fmt.Println(<-value)
				}

			}()
		},
	}
	multiWorker.doForAll()

	for i := 0; i < 10; i++ {
		multiWorkerChannel[i] <- i
	}
	// 思考 如果取消time来等待子携程 完成 （通过一条新携程）（waitGroup）
	time.Sleep(1000)
	for i := 0; i < 10; i++ {

		close(multiWorkerChannel[i])
	}

}

//func main() {
//	// 方法一：channel的创建赋值
//	// var ch chan int;
//	// ch = make(chan int);
//
//	// 方法二：短写法
//	// ch:=make(chan int);
//
//	// 方法三：综合写法:全局写法！！！！
//	var ch = make(chan int);
//
//	go write(2,ch);
//	time.Sleep(10);
//
//	c:=<-ch;
//	fmt.Printf("%v\n",c)
//}
//func write(a int,ch chan int){
//	ch<-a;
//}
