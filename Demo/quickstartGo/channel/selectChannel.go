package main

import (
	"fmt"
	"time"
)

/*
解决deadLock2 情况

通过 select 中实现 多个channel 进行无锁资源竞争
并发指的是一个CPU上运行多个线程 并行则是n个cpu运行n个线程
Go 语言通过 Go 实现并发编程

G 代表Goroutine 的 控制结构，是对GoRoutine的抽象  寄存器（上下文，当前的M，阻塞的时间）

M 简单理解为系统线程（具有CPU资源），存（当前执行的G，带有调度栈的G）

P 维护一个队列（存放G，全局队列，本地队列）
当有任务时会调用M来执行任务，这个队列长度决定了并发任务的数量
可以通过runtime.GOMAXPROCS进行指定。在Go1.5之后GOMAXPROCS被默认设置可用的核数

# PM需要进行绑定 构成一个执行单元，才能进行线程循环调度，意味着没有P的M不会执行

本地队列： 当前P的队列，本地队列是Lock-Free，没有数据竞争问题，无需加锁处理，可以提升处理速度。
全局队列：全局队列为了保证多个P之间任务的平衡。所有M共享P全局队列，为保证数据竞争问题，需要加锁处理。相比本地队列处理速度要低于全局队列

一个Goroutine 的运行过程

新建一个G 对象 ， 挂到P的队列中
如果有P是空闲的，调度器就将GoRoutine 对象与 Machine对象绑定
M 调度执行G对象，
携程抢占
当PM进行绑定时间超过一定时间或执行G的时候发生系统调用，系统会将P转移到新的M上继续执行新的G内容。
当发生上下文切换时，Machine对象会将内部的调度栈信息放入当前执行的Goroutine对象中进行保存。
当下次获得Cpu资源时，将调用栈信息从G中取出，继续完成。

清理线程（与P解除绑定）后，继续等待新的G执行。
*/
func createWorker(id int) chan int {
	out := make(chan int)
	go func() {
		// 没有值会阻塞
		fmt.Println(id, <-out)
	}()
	return out
}

func consumers(c4 <-chan time.Time, chs ...chan int) {

	// nil 在case 中会阻塞

	worker := createWorker(3)
	var result int
	for true {
		if result != 0 {
			chs[2] = worker
		}
		fmt.Println("next iterator")
		select {

		/*
				case 1 : 1
			case 3 :
			case 2 : 2

			1 一定比 2 快， 因为他在父携程中是有序发送的

			2 和 3 无法确定 因为 当1 收到值后
				同一时刻chs[2] 也可以进行写入 chs[1]可以进行读取
				c4 也可以读值， 当c4 获取到 值后 子携程退出 父携程因为c2写入阻塞 而无法结束
			每一次select 执行一个case
		*/

		case n := <-chs[1]:
			fmt.Println("case 2 :", n)
			//time.Sleep(time.Second*1)
			result = n
		case n := <-chs[0]:
			fmt.Println("case 1 :", n)
			result = n
			//time.Sleep(time.Second*5)
		// nil 会阻塞  信号传递
		case chs[2] <- result:
			fmt.Println("case 3 :")
			//time.Sleep(time.Second*5)
		case <-c4:
			fmt.Println("boom!!")
			//return
		// 为什么c4 只执行一次 这个case 执行多次？？？ 每次阻塞后会返回一个新的channel给case  并且全体阻塞超过5秒后 这个case 就会收到值
		// 如果在5秒内其他case运行，则该case重新获取新的channel
		case <-time.After(time.Second * 5):
			fmt.Println("timeout!!")

			//default:
			//	fmt.Println("nil")
		}
	}
}

func main() {
	c1 := make(chan int)
	c2 := make(chan int)
	var c3 chan int
	c4 := time.After(time.Second * 1)
	//
	go consumers(c4, c1, c2, c3)

	//for i := 0; i < 10; i++ {
	//	c1 <-i
	//	time.Sleep(time.Second*1)
	//}
	for i := 0; i < 10; i++ {

		time.Sleep(time.Second * 1)
		c2 <- 2 * i
	}
	close(c1)
	close(c2)
}
