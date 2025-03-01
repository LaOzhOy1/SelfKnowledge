package main

import "fmt"

type student struct {
	id   int32
	name string
}

/*
**
1. %v    只输出所有的值

2. %+v 先输出字段类型，再输出该字段的值

3. %#v 先输出结构体名字值，再输出结构体（字段类型+字段的值）
*/
func main() {
	a := &student{id: 1, name: "xiaoming"}
	change(*a)

	fmt.Printf("a=%v	\n", a)
	fmt.Printf("a=%+v	\n", a)
	fmt.Printf("a=%#v	\n", a)
}

func change(student2 student) {

	student2.id = 2
}
