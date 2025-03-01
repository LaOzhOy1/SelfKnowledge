package main

import "fmt"

/*
*
设计一个算法，找到数组中所有和为指定值的整数对。
*/
func main() {
	arr := []int{1, 2, 3, 4}
	results := findPairs(arr, 5)
	for i := 0; i < len(results); i++ {
		for j := 1; j < len(results[i]); j++ {
			fmt.Printf("%d key pair is %d\n", results[i][0], results[i][j])
		}
	}
}
func findPairs(arr []int, target int) [][]int {
	results := make([][]int, 0)
	for i := 0; i < len(arr); i++ {
		results = append(results, []int{arr[i]})
	}

	// 遍历n-1次
	for i := 0; i < len(arr)-1; i++ {
		// 查找与i和为指定值的键值对
		for j := i + 1; j < len(arr); j++ {
			if target-arr[i] == arr[j] {
				results[i] = append(results[i], arr[j])
			}
		}
	}
	return results
}
