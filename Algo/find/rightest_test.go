package find

import (
	"fmt"
	"testing"
)

func TestRightest(t *testing.T) {
	arr := []int{1, 25, 25, 25, 90, 124, 422}
	fmt.Println(rightest(arr, 25))
}

func rightest(arr []int, target int) int {
	var mid, l, index, r int
	r = len(arr) - 1
	for l <= r {
		mid = l + ((r - l) >> 1)
		if arr[mid] > target {
			r = mid - 1
		} else if arr[mid] <= target {
			index = mid
			l = mid + 1
		}
	}
	return index
}
