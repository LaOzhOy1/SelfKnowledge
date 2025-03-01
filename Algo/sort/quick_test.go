package sort

import "testing"

func TestQuick1(t *testing.T) {
	arr := NewRandomArray(100, 100)
	Quick1(arr, 0, len(arr)-1)
}

func TestQuick2(t *testing.T) {
	arr := NewRandomArray(100, 100)
	Quick2(arr, 0, len(arr)-1)

}

func TestQuick3(t *testing.T) {
	arr := NewRandomArray(100, 100)
	Quick3(arr, 0, len(arr)-1)
}
