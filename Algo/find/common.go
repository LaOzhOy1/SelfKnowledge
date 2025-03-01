package find

import "math/rand"

func exchange(a []int, i, j int) {
	temp := a[i]
	a[i] = a[j]
	a[j] = temp
}

func NewRandomArray(maxSize int, maxValue int) []int {
	size := rand.Intn(maxSize) + 1
	arr := make([]int, size)
	for i := 0; i < size; i++ {
		arr[i] = rand.Intn(maxValue) - rand.Intn(maxValue)
	}
	return arr
}
