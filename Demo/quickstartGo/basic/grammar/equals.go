package main

import (
	"fmt"
	"net"
	"reflect"
)

/*
 介绍 =:  = 的区别
 介绍struct 等号判断
*/

func main() {
	/*****
	  	整型和浮点型变量的默认值为 0，如var a int，默认a=0
	    字符串变量的默认值为空字符串
	    布尔型变量默认为 bool
	    切片、函数、指针变量的默认为 nil
	*/
	var empty int

	empty = 1
	empty = empty

	// ：= 会确定属性
	a := "1"
	a = a
	// 也会确定属性
	var b = "2"
	var c = b
	c = c
	d := b
	d = d
	// 会做属性检查
	//d = empty
	//No new variables on left side of :=
	//d := d
	var conn net.Conn
	var err error
	conn, err = net.Dial("tcp", "127.0.0.1:8080")
	conn, err = net.Dial("tcp", "127.0.0.1:8080")

	conn = conn
	err = err

	//虽然err重复声明了，但是conn和conn2没有重复声明，只要有一个新声明，不会报错  元组
	_, err = net.Dial("tcp", "127.0.0.1:8080")
	_, err = net.Dial("tcp", "127.0.0.1:8080")

	////重复声明了
	//conn, err := net.Dial("tcp", "127.0.0.1:8080")
	//conn, err := net.Dial("tcp", "127.0.0.1:8080")

	// 判断 slice 是否相等 https://www.cnblogs.com/apocelipes/p/11116725.html  利用了反射
	as := []int{1, 2, 3, 4}
	bs := []int{1, 3, 2, 4}
	cs := []int{1, 2, 3, 4}
	fmt.Println(reflect.DeepEqual(as, bs))
	fmt.Println(reflect.DeepEqual(as, cs))
	// 判断方法是不是一样
	_ = reflect.TypeOf(a).Kind() == reflect.TypeOf(b).Kind()
	var g, f I
	reflect.DeepEqual(g, f)
	reflect.ValueOf(a).Elem().Set(reflect.ValueOf(b))

	type A interface{}
	type B struct{}
	// 先当于一个 类型挂载了nil值
	var s *B
	// 值比较
	print(s == nil)       //true
	print(s == (*B)(nil)) //true

	//当interface{} （泛型）与一个值做比较的时候，会同时比较type和value，都相等的时候，才会认为相等。下方中的(*B)(nil)的type与(A)(a)相同，都是*B，值也相同，都是nil，所以他们就相等了。
	print((A)(s) == (*B)(nil)) //true
	//
	print((A)(s) == nil) //false  nil == (int*)nil

}

type I interface{}

// 自定义实现
func testEq(a, b []int) bool {
	// If one is nil, the other must also be nil.
	if (a == nil) != (b == nil) {
		return false
	}

	if len(a) != len(b) {
		return false
	}

	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}

	return true
}
