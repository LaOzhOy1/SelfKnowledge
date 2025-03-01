package sort

import (
	"fmt"
	"testing"
)

func TestExchange(t *testing.T) {
	a := NewRandomArray(10, 100)
	Exchange(a)
	fmt.Println(a)
}
