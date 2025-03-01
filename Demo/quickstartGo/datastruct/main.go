package main

import (
	"fmt"
	"unsafe"
)

// 定义非空结构体
type S struct {
	a interface{} // prints 16
	b interface{} // prints 16
}

type SE struct {
	a uint16 //  prints 2,
	b int16  // prints2  -32768 through 32767
	c uint32 //  prints 6,
	d uint64 //  prints 8

}

func main() {
	var s SE
	fmt.Println(unsafe.Sizeof(s)) // not 6
	var s2 struct{}
	fmt.Println(unsafe.Sizeof(s2)) // prints 0
}
