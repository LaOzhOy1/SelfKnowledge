package main

import "fmt"

func main() {

	arr := [...]int{1, 2, 3, 4, 5, 6, 7, 8}

	fmt.Println(divideTwo(arr[:], 0, len(arr)-1, 5))
	fmt.Println(divideTwo(arr[:], 0, len(arr)-1, 4))
	fmt.Println(divideTwo(arr[:], 0, len(arr)-1, 3))
	fmt.Println(divideTwo(arr[:], 0, len(arr)-1, 2))
}

func divideTwo(arr []int, low int, high int, target int) int {

	l := low
	h := high
	for l <= h {
		mid := l + (h-l)/2
		if target > arr[mid] {
			l = mid + 1
		} else if arr[mid] == target {
			return mid
		} else if target < arr[mid] {
			h = mid - 1
		}
	}
	return -1
}

func LIS(arr []int) []int {
	// write code here

	end := make([]int, len(arr)+1)

	last := -1
	length := 1
	for i := 0; i < len(arr); i++ {
		if last == -1 {
			last = arr[i]
			end[length] = last
			continue
		}
		if arr[i] > last {
			length++

		} else {
			for j := length; j >= 1; j-- {
				if end[j] < arr[i] {
					//length =
				}
			}
		}

		last = arr[i]
	}
	return nil
}
