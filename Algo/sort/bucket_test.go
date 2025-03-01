package sort

import "testing"

func TestBucket(t *testing.T) {
	arr := NewRandomArray(15, 14)
	BucketSort(arr, 0, len(arr)-1, 2)
}
