package find

import (
	"fmt"
	"testing"
)

func TestPart(t *testing.T) {
	arr := []int{1, 25, 25, 25, 90, 124, 422}
	fmt.Println(partTest(arr, 25))
}

func partTest(arr []int, target int) int {
	var mid, l, r int
	r = len(arr) - 1
	if arr[l] < arr[l+1] {
		return l
	}
	if arr[r] < arr[r-1] {
		return r
	}
	l = 1
	r = r - 1
	for l <= r {
		mid = l + (r-l)>>1
		if arr[mid] > arr[mid+1] {
			l = mid + 1
		} else if arr[mid] > arr[mid-1] {
			r = mid - 1
		} else {
			return arr[mid]
		}
	}
	return -1
}
