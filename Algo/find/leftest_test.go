package find

import (
	"fmt"
	"testing"
)

func TestLeftest(t *testing.T) {
	arr := []int{1, 25, 25, 25, 90, 124, 422}
	fmt.Println(leftest(arr, 25))
}

func leftest(arr []int, target int) int {
	var mid, l, index, r int
	r = len(arr) - 1
	for l <= r {
		mid = l + ((r - l) >> 1)
		if arr[mid] >= target {
			index = mid
			r = mid - 1
		} else if arr[mid] < target {
			l = mid + 1
		}
	}
	return index
}
