package main

import (
	"fmt"
	"unicode/utf8"
)

func main() {
	//从字符串字面值看len(str)的结果应该是8,但在Golang中string类型的底层是通过byte数组实现的,在unicode编码中,中文字符占两个字节,而在utf-8编码中,中文字符占三个字节而Golang的默认编码正是utf-8.
	var str = "hello 世界"
	fmt.Println("len(str):", len(str))
	fmt.Println("RuneCountInString:", utf8.RuneCountInString(str))

	fmt.Println("rune:", len([]rune(str)))
}
