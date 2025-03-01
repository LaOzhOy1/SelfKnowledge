package recursion

import "math"

// 0...length -1
func findMax(arr []int, start int, end int) int {
	// spec
	if start == end {
		return arr[start]
	}
	mid := start + (end-start)>>1
	rMax := findMax(arr, mid+1, end)
	lMax := findMax(arr, start, mid)
	return int(math.Max(float64(arr[mid]), math.Max(float64(rMax), float64(lMax))))
}
