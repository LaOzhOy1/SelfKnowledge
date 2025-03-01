package main

import (
	"fmt"
	"time"
)

func main() {
	Sort()
	SortByRoutine()
}

func worker(c chan []int) {
	for true {
		arr := <-c
		bubbleSort(arr)
	}
}

func SortByRoutine() {
	start := time.Now().UnixNano()
	const NUM int = 100
	arr := []int{1, 2, 10, 4, 5, 6, 7, 8, 9}
	ch := make(chan []int)
	worker(ch)
	for i := 0; i < NUM; i++ {
		go func() {
			ch <- arr
		}()
	}
	//打印消耗时间
	fmt.Println(time.Now().UnixNano() - start)
}

func Sort() {
	start := time.Now().UnixNano()
	const NUM int = 100
	for i := 0; i < NUM; i++ {
		arr := []int{1, 2, 10, 4, 5, 6, 7, 8, 9}
		bubbleSort(arr)
	}
	//打印消耗时间
	fmt.Println(time.Now().UnixNano() - start)
}

// 排序
func bubbleSort(arr []int) {
	for j := 0; j < len(arr)-1; j++ {
		for k := 0; k < len(arr)-1-j; k++ {
			if arr[k] < arr[k+1] {
				temp := arr[k]
				arr[k] = arr[k+1]
				arr[k+1] = temp
			}
		}
	}
}
