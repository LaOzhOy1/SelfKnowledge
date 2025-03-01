package sort

import (
	"fmt"
	"testing"
)

func TestInsert(t *testing.T) {
	a := NewRandomArray(10, 100)
	Insert(a)
	fmt.Println(a)
}
