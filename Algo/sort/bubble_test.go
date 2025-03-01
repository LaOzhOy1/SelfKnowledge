package sort

import (
	"fmt"
	"testing"
)

func TestBubble(t *testing.T) {

	a := NewRandomArray(10, 100)
	Bubble(a)
	fmt.Println(a)
}
