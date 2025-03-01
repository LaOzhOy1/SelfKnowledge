package heap

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
	heap.adjust()
	return heap
}

func (heap *Heap) adjust() {
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
	if heap.Size == 0 {
		return -1
	}
	target := heap.Data[0]
	heap.Size--
	exchange(heap.Data, 0, heap.Size)
	heap.adjust()
	return target
}

func (heap *Heap) formatChild(parent int) {
	left := parent*2 + 1
	for left < len(heap.Data) {
		right := left + 1
		min := heap.Data[left]
		minIdx := left
		if right < len(heap.Data) {
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

func (heap *Heap) formatParent(index int) {
	parent := (index - 1) / 2
	for heap.Data[index] < heap.Data[parent] {
		exchange(heap.Data, index, parent)
		index = parent
	}
}
