package sort

import (
	"fmt"
	"testing"
)

func TestHeapSort(t *testing.T) {
	arr := NewRandomArray(15, 10)
	fmt.Println(arr)
	HeapSort(arr)
}

// 排序后 移动的位置不超过K

func TestKHeapSort(t *testing.T) {

}
