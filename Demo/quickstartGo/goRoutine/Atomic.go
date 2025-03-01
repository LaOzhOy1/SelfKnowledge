package main

import (
	"fmt"
	"sync"
)

/*
使用Mute 与 defer
（defer 语句正好是在函数退出时执行的语句，所以使用 defer 能非常方便地处理资源释放问题）
（使用匿名函数实现代码区）
实现go 中 的 互斥量
用race 检测效果
*/

type atomicInt int

var lock sync.Mutex

func (atomic *atomicInt) add() {
	func() {
		lock.Lock()
		defer lock.Unlock()
		*atomic++
	}()
}
func main() {
	var number atomicInt

	number.add()
	go func() {
		number.add()
	}()
	lock.Lock()
	defer lock.Unlock()
	fmt.Print(number)

	//time.Sleep(time.Second)

}
