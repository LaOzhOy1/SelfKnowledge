package main

import "fmt"

/*
介绍invoker 的多种情况
*/

type PointContent struct {
	a int
}

func (receiver PointContent) fixAWithRef() {
	receiver.a = 3
}

func (receiver *PointContent) fixAWithPointer() {
	receiver.a = 4
}

type ArrayContent struct {
	a []int
}

// 测试数组类型  0 7 8
func (receiver ArrayContent) fixArrWithRef() {
	receiver.a[0] = 7
}

func (receiver *ArrayContent) fixArrWithPointer() {
	receiver.a[0] = 8
}

// 测试改变引用
// 预测 8 11  正确
// 原因是 传递类过来 是copy了一个临时引用指向对应内存数据
// 方法体内赋值 出去方法体后 临时引用就销毁了
// 但是 return 确实可以把临时变量 传回去给 调用函数方使用  原因归结为 底层方法栈和返回值栈  不回回收局部变量的内存
//
//	意思是说go语言编译器会自动决定把一个变量放在栈还是放在堆，编译器会做逃逸分析(escape analysis)，当发现变量的作用域没有跑出函数范围，就可以在栈上，反之则必须分配在堆。
func (receiver ArrayContent) fixArrWithRefReplace() {
	tmp := make([]int, 3)
	tmp[0] = 10
	receiver.a = tmp
}
func (receiver *ArrayContent) fixArrWithPointerReplace() {
	tmp := make([]int, 3)
	tmp[0] = 11
	receiver.a = tmp
}

type testObject struct {
	b int
}

type PointContent2 struct {
	a testObject
}

func (receiver PointContent2) fixArrWithRefReplace() {
	receiver.a.b = 4
}
func (receiver *PointContent2) fixArrWithPointerReplace() {
	receiver.a.b = 8
}
func (receiver PointContent3) fixArrWithRefReplace() {
	receiver.a["1"] = "4"
}
func (receiver *PointContent3) fixArrWithPointerReplace() {
	receiver.a["1"] = "8"
}

type PointContent3 struct {
	a map[string]string
}
type testObject2 func(int) int
type PointContent4 struct {
	add testObject2
}

func (receiver PointContent4) fixArrWithRefReplace() {
	receiver.add = func(i int) int {
		return i - 1
	}
}
func (receiver *PointContent4) fixArrWithPointerReplace() {
	receiver.add = func(i int) int {
		return i - 1
	}
}

func main() {
	// 作为一个调用方  简单数据
	fmt.Print("p :  ")
	p := PointContent{
		a: 2,
	}

	fmt.Print(p.a)

	p.fixAWithRef()
	fmt.Print(p.a)

	p.fixAWithPointer()
	// 发送改变
	fmt.Println(p.a)

	fmt.Print("p2 :  ")
	p2 := PointContent2{
		a: testObject{
			b: 2,
		},
	}
	fmt.Print(p2.a.b)
	p2.fixArrWithRefReplace()
	// 不改变 发生深拷贝
	fmt.Print(p2.a.b)

	p2.fixArrWithPointerReplace()
	fmt.Println(p2.a.b)

	fmt.Print("p3 :  ")
	p3 := PointContent3{
		a: map[string]string{"1": "1,", "2": "2"},
	}
	fmt.Print(p3.a["1"])
	p3.fixArrWithRefReplace()
	// 改变 发生深拷贝
	fmt.Print(p3.a["1"])

	p3.fixArrWithPointerReplace()
	fmt.Println(p3.a["1"])

	fmt.Print("p4 :  ")
	p4 := PointContent4{
		add: func(i int) int {
			return i + 1
		},
	}
	fmt.Print(p4.add(1))
	p4.fixArrWithRefReplace()
	// 改变 发生深拷贝
	fmt.Print(p4.add(1))

	p4.fixArrWithPointerReplace()
	fmt.Println(p4.add(1))

	fmt.Print("arr :  ")
	// 切片的特殊性
	arr := ArrayContent{
		a: make([]int, 2),
	}
	fmt.Print(arr.a[0])
	// 改变
	arr.fixArrWithRef()
	fmt.Print(arr.a[0])
	// 改变  会自动取他的pointer
	arr.fixArrWithPointer()
	fmt.Println(arr.a[0])

	arr.fixArrWithRefReplace()
	// 不改变  在 局部方法中 修改了引用  没有改动到真实的数据
	fmt.Print(arr.a[0])
	arr.fixArrWithPointerReplace()
	// 改变  引用被替换了
	fmt.Println(arr.a[0])

	arr2 := ArrayContent{
		a: []int{0, 2, 3, 4},
	}
	fmt.Print("arr2 :  ")
	fmt.Print(arr2.a[0])
	// 改变
	arr2.fixArrWithRef()
	fmt.Print(arr2.a[0])
	// 改变  会自动取他的pointer
	arr2.fixArrWithPointer()
	fmt.Println(arr2.a[0])

	arr2.fixArrWithRefReplace()
	// 不改变  在 局部方法中 修改了引用  没有改动到真实的数据
	fmt.Print(arr2.a[0])
	arr2.fixArrWithPointerReplace()
	// 改变
	fmt.Println(arr2.a[0])

}
