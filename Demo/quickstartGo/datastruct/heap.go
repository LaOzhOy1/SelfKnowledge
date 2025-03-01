package main

import "fmt"

// 那么最坏情况下的时间复杂度为O(k + (n-k)logk) = O(nlogk)
// 用最大堆 获取最小的k个数
// https://blog.csdn.net/weixin_40051278/article/details/103640680
func main() {

	var a = 1
	var b = 2
	swapValue(a, b)
	fmt.Println(a)
	fmt.Println(b)

	arr := [...]int{1, 2, 6, 5, 42, 3, 4, 5, 5}

	//CreateMaxHeap(arr[:])
	topK(arr[:], 7)
	//swap(arr[:],0, len(arr)-1)
	//
	fmt.Printf("%v", arr)
}

// 维护最大堆
func topK(data []int, k int) {
	// 建立前K个数的最大堆
	CreateMaxHeap(data[0:k])
	for i := k; i < len(data); i++ {
		if data[i] < data[0] { //如果剩余的数中有小的数则替换
			swap(data, i, 0)
			adjustHeap(data[0:k], 0)
		}
	}
}

func CreateMaxHeap(array []int) {
	for i := len(array)/2 - 1; i >= 0; i-- {
		adjustHeap(array, i)
	}
}

func adjustHeap(subArray []int, root int) {
	for left := 2*root + 1; left < len(subArray); left = 2*left + 1 {
		if left == len(subArray)-1 {
			if subArray[left] > subArray[root] {
				swap(subArray, left, root)
				root = left
			}
		} else {
			if subArray[left] > subArray[left+1] &&
				subArray[left] > subArray[root] {
				swap(subArray, left, root)
				root = left
			} else if subArray[left] < subArray[left+1] &&
				subArray[left+1] > subArray[root] {
				swap(subArray, left+1, root)
				root = left
			}
		}
	}
}

func swap(arr []int, left int, right int) {

	arr[left] ^= arr[right]
	arr[right] ^= arr[left]
	arr[left] ^= arr[right]

}

func swapValue(a int, b int) {
	a ^= b
	b ^= a
	a ^= b

}
