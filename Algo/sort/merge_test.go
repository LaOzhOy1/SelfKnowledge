package sort

import "testing"

func TestMerge(t *testing.T) {
	arr := NewRandomArray(50, 1024)
	IterateMerge(arr)

	arr = NewRandomArray(50, 1024)
	RecursionMerge(arr, 0, len(arr)-1)
}

// 求 左侧最小和
func TestLeftMinSum(t *testing.T) {
	arr := NewRandomArray(50, 1024)
	RecursionMergeWithMin(arr, 0, len(arr)-1)
}

func RecursionMergeWithMin(a []int, start int, end int) int {
	if start == end {
		return 0
	}
	mid := start + (end-start)>>1
	return RecursionMergeWithMin(a, start, mid) + RecursionMergeWithMin(a, mid+1, end) + mergeWithMin(a, start, mid, end)
}

func mergeWithMin(a []int, start int, mid int, end int) int {
	arr := make([]int, (end-start)+1)
	p1 := start
	p2 := mid + 1
	sum := 0
	idx := 0
	for p1 <= mid && p2 <= end {
		if a[p1] < a[p2] {

			sum += (end - p2 + 1) * a[p1]
			arr[idx] = a[p1]
			p1++
		} else {
			arr[idx] = a[p2]
			p2++
		}
		idx++
	}
	for p1 <= mid {
		arr[idx] = a[p1]
		p1++
		idx++
	}
	for p2 <= end {
		arr[idx] = a[p2]
		p2++
		idx++
	}
	for i := 0; i < len(arr); i++ {
		a[start+i] = arr[i]
	}
	return sum
}

// 逆序对
func TestCouple(t *testing.T) {
	arr := NewRandomArray(50, 1024)
	RecursionMergeWithCouple(arr, 0, len(arr)-1)
}
func RecursionMergeWithCouple(a []int, start int, end int) int {
	if start == end {
		return 0
	}
	mid := start + (end-start)>>1
	return RecursionMergeWithMin(a, start, mid) + RecursionMergeWithMin(a, mid+1, end) + mergeWithCouple(a, start, mid, end)
}
func mergeWithCouple(a []int, start int, mid int, end int) int {
	arr := make([]int, (end-start)+1)
	p1 := start
	p2 := mid + 1
	sum := 0
	idx := 0
	for p1 <= mid && p2 <= end {
		if a[p1] > a[p2] {
			sum += end - p2 + 1
			arr[idx] = a[p1]
			p1++
		} else {
			arr[idx] = a[p2]
			p2++
		}
		idx++
	}
	for p1 <= mid {
		arr[idx] = a[p1]
		p1++
		idx++
	}
	for p2 <= end {
		arr[idx] = a[p2]
		p2++
		idx++
	}
	for i := 0; i < len(arr); i++ {
		a[start+i] = arr[i]
	}
	return sum
}
