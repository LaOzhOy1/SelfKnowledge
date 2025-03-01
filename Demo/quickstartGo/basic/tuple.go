package main

import "fmt"

// 返回 X+Y 和 X*Y
func Computer2(X, Y int) (int, int) {
	return X + Y, X * Y
}

// 返回 X+Y 和 X*Y
func Computer(X, Y int) (add int, multiplied int) { //这里命令返回参数的变量add 和 multiplied
	add = X + Y
	multiplied = X * Y
	return
}

func main() {
	x := 10
	y := 20

	a, b := Computer(x, y)

	fmt.Printf("%d + %d = %d\n", x, y, a)
	fmt.Printf("%d * %d = %d\n", x, y, b)
}
