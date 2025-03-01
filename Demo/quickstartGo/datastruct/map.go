package main

import (
	"fmt"
)

// go 中的 HashMap
func main() {

	var g1 map[int]string //默认值为nil

	g2 := map[int]string{}

	g3 := make(map[int]string)

	g4 := make(map[int]string, 10)

	fmt.Println(g1, g2, g3, g4, len(g4))

	//2、初始化

	var m1 map[int]string = map[int]string{1: "ck_god", 2: "god_girl"}

	m2 := map[int]string{1: "ck_god", 2: "god_girl"}

	fmt.Println(m1, m2)

	dict := map[string]int{"key1": 1, "key2": 2}
	value, ok := dict["key1"]
	if ok {
		fmt.Printf(string(rune(value)))
	} else {
		fmt.Println("key1 不存在")
	}
	m := make(map[string]int)
	d := make([]int, 0)
	if val := m["2"]; val != 0 {
		if val == 1 {

		}
	}
	d = append(d, 1)
	vale := byte(2)
	_ = string(vale) + string(rune(2))

	dict2 := map[string]int{"key1": 1, "key2": 2}
	if value, ok := dict2["key1"]; ok {
		fmt.Printf(string(rune(value)))
	} else {
		fmt.Println("key1 不存在")
	}
}
