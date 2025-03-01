package main

import (
	"fmt"
)

var global = 0

func main() {

	/**
	由于变量是在方法体内定义，根据go的内存逃逸分析，闭包中的方法变量在调用完后分配在栈上，会被回收，所以每次调用会从0开始，而闭包内的改变不会影响到return的值，所以是深拷贝
	testWithMethodSend, f =  2
	test return is : 0
	*/
	for t := 0; t < 10; t++ {
		ret := testWithMethodSend()
		fmt.Printf("test return is : %d\n", ret)
	}

	/**
	testGlobal, global is 2
	testGlobal return is : 0, real global is :2
	全局变量分配在堆中
	*/
	//for t := 0; t< 10; t++ {
	//	ret := testGlobal()
	//	fmt.Printf("testGlobal return is : %d, real global is :%d\n", ret,global)
	//}
}

func test() int {
	var i = 0
	defer func() {
		i += 2 //defer里面对i增1
		fmt.Println("test, i = ", i)
	}()
	return i
}

func testGlobal() int {
	defer func() {
		global += 2 //defer里面对i增1
		fmt.Println("testGlobal, global is", global)
	}()
	return global
}

func testWithMethodSend() int {
	var i = 0
	defer func(f int) {
		f += 2 //defer里面对i增1
		fmt.Println("testWithMethodSend, f = ", f)
	}(i)
	return i
}
