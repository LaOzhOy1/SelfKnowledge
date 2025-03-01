package find

import (
	"fmt"
	"testing"
)

func TestExist(t *testing.T) {
	arr := []int{1, 25, 90, 124, 422}
	fmt.Println(Exist(arr, 25))
}

/**
while(left <= right) 的终⽌条件是 left == right + 1，写成区间的形式就是 [right + 1, right]，或者带个具体的数字进去 [3, 2]，可⻅这时候区间为空，
因 为没有数字既⼤于等于 3 ⼜⼩于等于 2 的吧。所以这时候 while 循环终⽌是 正确的，直接返回 -1 即可。
while(left < right) 的终⽌条件是 left == right，写成区间的形式就是 [left, right]，或者带个具体的数字进去 [2, 2]，这时候区间⾮空，还有⼀个数 2，
但此时 while 循环终⽌了。也就是说这区间 [2, 2] 被漏掉了，索引 2 没有被 搜索，如果这时候直接返回 -1 就是错误的。

*/
// review done
func Exist(arr []int, target int) bool {
	var mid, l, r int
	r = len(arr) - 1
	for l <= r {
		mid = l + ((r - l) >> 1)
		if arr[mid] == target {
			return true
		} else if arr[mid] > target {
			r = mid - 1
		} else if arr[mid] < target {
			l = mid + 1
		}
	}
	r = len(arr) - 1
	mid = 0
	l = 0
	for l <= r {
		mid = l + (r-l)>>1
		if target == arr[mid] {
			return true
		} else if target > arr[mid] {
			l = mid + 1
		} else {
			r = mid - 1
		}
	}
	return false
}
