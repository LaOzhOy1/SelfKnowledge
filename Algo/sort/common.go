package sort

import (
	"fmt"
	"math"
	"math/rand"
)

func exchange(a []int, i, j int) {
	temp := a[i]
	a[i] = a[j]
	a[j] = temp
}

type Heap struct {
	Data []int
	Size int
}

func NewHeap(data []int) Heap {
	heap := Heap{Data: data, Size: len(data)}
	heap.Adjust()
	return heap
}

func (heap *Heap) Adjust() {
	idx := (heap.Size - 2) / 2
	for idx >= 0 {
		heap.formatChild(idx)
		idx--
	}
}

// O(N) * O(log N)
func (heap *Heap) Push(value int) {
	if heap.Size > len(heap.Data) {
		heap.Data = append(heap.Data, value)
	} else {
		heap.Data[heap.Size] = value
	}
	heap.formatParent(heap.Size)
	heap.Size++
}

func (heap *Heap) Pop() int {
	target := heap.Data[0]
	heap.Size--
	exchange(heap.Data, 0, heap.Size)
	heap.Adjust()
	return target
}
func (heap *Heap) formatParent(index int) {
	parent := (index - 1) / 2
	for heap.Data[index] < heap.Data[parent] {
		exchange(heap.Data, index, parent)
		index = parent
	}
}
func (heap *Heap) formatChild(parent int) {
	left := parent*2 + 1
	for left < heap.Size {
		right := left + 1
		min := heap.Data[left]
		minIdx := left
		if right < heap.Size {
			if heap.Data[right] < min {
				min = heap.Data[right]
				minIdx = right
			}
		}
		if min >= heap.Data[parent] {
			break
		}
		exchange(heap.Data, minIdx, parent)
		parent = minIdx
		left = parent*2 + 1
	}
}

func (heap Heap) Println() {
	for _, data := range heap.Data {
		fmt.Println(data)
	}
}

func HeapSort(arr []int) {
	heap := NewHeap(arr)
	fmt.Println(heap.Data)
	for heap.Size > 0 {
		heap.Size--
		exchange(heap.Data, 0, heap.Size)
		heap.Adjust()
		fmt.Println(heap.Data)
	}
}

func BucketSort(arr []int, start int, end int, radix int) {
	help := make([]int, end-start+1)
	radixSum := make([]int, 10)
	for i := 1; i <= radix; i++ {

		for _, v := range arr {
			c := getDigit(v, radix)
			radixSum[c]++
		}

		for idx := 1; idx < len(radixSum); idx++ {
			radixSum[idx] += radixSum[idx-1]
		}

		for _, v := range arr {
			n := getDigit(v, i)
			c := radixSum[n]
			help[c] = v
			radixSum[n]--
		}

		for i := 0; i < len(help); i++ {
			arr[start+i] = help[i]
		}
		radix--
	}
}

func getDigit(value int, radix int) int {
	r := int(math.Pow(10.0, float64(radix-1)))
	return value / r % 10
}

func getMaxRadix(a []int) int {
	var sum int
	for _, v := range a {
		temp := 0
		for v > 0 {
			temp++
			v /= 10
		}
		if temp > sum {
			sum = temp
		}
	}
	return sum
}

func Bubble(a []int) {
	for i := 0; i < len(a); i++ {
		for j := 0; j < len(a)-1; j++ {
			if a[j] > a[j+1] {
				exchange(a, j, j+1)
			}
		}
	}
}

// review done
func Exchange(a []int) {
	for i := 0; i < len(a)-1; i++ {
		var min = i
		for j := i; j < len(a); j++ {
			if a[j] < a[min] {
				min = j
			}
		}
		exchange(a, min, i)
	}

	for i := 0; i < len(a)-1; i++ {
		min := i
		for j := i + 1; j < len(a); j++ {
			if a[min] > a[j] {
				min = j
			}
		}
		exchange(a, min, i)
	}
}

// review done
func Insert(a []int) {
	for i := 1; i < len(a); i++ {
		for j := i - 1; j > 0 && a[j] > a[j+1]; j-- {
			exchange(a, j, j+1)
		}
	}

	for i := 1; i < len(a); i++ {
		for k := i - 1; k > 0 && a[k] > a[k+1]; k-- {
			exchange(a, k, k+1)
		}
	}
}

func Quick1(a []int, start int, end int) {
	if start >= end {
		return
	}
	mid := partition(a, start, end)
	Quick1(a, start, mid)
	Quick1(a, mid+1, end)
}

func Quick2(a []int, start int, end int) {
	if start >= end {
		return
	}
	mid := netherLandsFlag(a, start, end)
	Quick2(a, start, mid[0]-1)
	Quick2(a, mid[1]+1, end)
}

func Quick3(a []int, start int, end int) {
	if start >= end {
		return
	}
	exchange(a, rand.Intn(end-start+1), end)
	mid := netherLandsFlag(a, start, end)
	Quick3(a, start, mid[0]-1)
	Quick3(a, mid[1]+1, end)
}

// review done
func partition(a []int, start int, end int) int {
	if start > end {
		return -1
	}
	if start == end {
		return start
	}
	less := start - 1
	idx := start
	for idx < end {
		if a[idx] <= a[end] {
			less++
			exchange(a, idx, less)
		}
		idx++
	}
	less++
	exchange(a, end, less)

	if start > end {
		return -1
	}

	idx = 0
	less = start - 1
	target := a[end]
	for i := start; i < end; i++ {
		if a[i] < target {
			less++
			exchange(a, less, i)
		}
		idx++
	}
	less++
	exchange(a, less, end)
	return less
}

func netherLandsFlag(a []int, start int, end int) []int {
	if start > end {
		return nil
	}
	if start == end {
		return []int{start, start}
	}
	less := start - 1
	more := end
	idx := start
	for idx < more {
		if a[idx] < a[end] {
			less++
			exchange(a, idx, less)
			idx++
		} else if a[idx] == a[end] {
			idx++
		} else {
			more--
			exchange(a, idx, more)
		}
	}
	exchange(a, end, more)
	return []int{less + 1, more}
}

// review done
func IterateMerge(a []int) {
	partSize := 1
	for partSize < len(a) {
		l := 0

		for l < len(a) {
			m := l + partSize - 1
			if m >= len(a) {
				break
			}
			r := int(math.Min(float64(m+partSize), float64(len(a)-1)))
			merge(a, l, m, r)
			l = r + 1
		}
		if partSize > len(a)>>1 {
			break
		}
		partSize <<= 1
	}

	partSize = 1
	for partSize < len(a) {
		l := 0

		for l < len(a) {
			m := l + partSize - 1
			if m > len(a) {
				break
			}
			r := int(math.Min(float64(len(a)-1), float64(m+partSize)))
			merge(a, l, m, r)
			l = r + 1
		}
		if partSize > len(a)>>1 {
			break
		}
		partSize <<= 1
	}
}

func RecursionMerge(a []int, start int, end int) {
	if start == end {
		return
	}
	mid := start + (end-start)>>1
	RecursionMerge(a, start, mid)
	RecursionMerge(a, mid+1, end)
	merge(a, start, mid, end)
}

func merge(a []int, start int, mid int, end int) {
	mr := make([]int, end-start+1)
	p1 := start
	p2 := mid + 1
	idx := 0
	for p1 <= mid && p2 <= end {
		if a[p1] > a[p2] {
			mr[idx] = a[p2]
		} else {
			mr[idx] = a[p1]
		}
		idx++
	}
	for p2 <= end {
		mr[idx] = a[p2]
		idx++
	}
	for p1 <= mid {
		mr[idx] = a[p1]
		idx++
	}
	for i, m := range mr {
		a[start+i] = m
	}
}

func NewRandomArray(maxSize int, maxValue int) []int {
	size := rand.Intn(maxSize) + 1
	arr := make([]int, size)
	for i := 0; i < size; i++ {
		arr[i] = rand.Intn(maxValue) - rand.Intn(maxValue)
	}
	return arr
}
