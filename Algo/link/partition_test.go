package link

import (
	"fmt"
	"math/rand"
	"testing"
)

type listPartitionNode struct {
	Next *listPartitionNode
	V    int
}

func (l *listPartitionNode) println() {
	t := l
	for t != nil {
		fmt.Printf("%d, ", t.V)
		t = t.Next
	}
	fmt.Println()
}

func exchange(arr []int, i, j int) {
	tmp := arr[i]
	arr[i] = arr[j]
	arr[j] = tmp
}

// 通过 partition 数组分区

func TestPartition1(t *testing.T) {
	var first *listPartitionNode = nil
	for i := 0; i < 10; i++ {
		s := new(listPartitionNode)
		s.V = rand.Intn(100)
		s.Next = first
		first = s
	}
	first.V = 81
	first.println()
	//ListPartition(first)
	sort := partitionO1(first)
	sort.println()
}

func TestPartition(t *testing.T) {
	var first *listPartitionNode = nil
	for i := 0; i < 10; i++ {
		s := new(listPartitionNode)
		s.V = rand.Intn(100)
		s.Next = first
		first = s
	}
	first.println()
	ListPartition(first)
	first.println()
}

func ListPartition(head *listPartitionNode) {
	arr := make([]int, 0)
	t := head
	for t != nil {
		arr = append(arr, t.V)
		t = t.Next
	}
	partition(arr, 0, len(arr)-1)
	for _, val := range arr {
		fmt.Printf("%d, ", val)
	}
	fmt.Println()
	l := head
	for _, v := range arr {
		l.V = v
		l = l.Next
	}
}

func partition(arr []int, start, end int) {
	less := start - 1
	more := end
	target := arr[end]
	cur := start
	for cur < more {
		if arr[cur] < target {
			less++
			exchange(arr, cur, less)
			cur++
		} else if arr[cur] == target {
			cur++
		} else {
			more--
			exchange(arr, cur, more)
		}
	}
	exchange(arr, more, end)
}

// O（1） 优化

func partitionO1(head *listPartitionNode) *listPartitionNode {
	var lessL, lessR, equalL, equalR, moreL, moreR *listPartitionNode
	cur := head
	if cur == nil {
		return head
	}
	target := cur.V
	equalL = cur
	equalR = cur
	next := cur.Next
	cur.Next = nil
	cur = next
	for cur != nil {
		if cur.V < target {
			if lessL == nil || lessR == nil {
				lessL = cur
				lessR = cur
			} else {
				lessR.Next = cur
				lessR = lessR.Next
			}
		} else if cur.V == target {
			equalR.Next = cur
			equalR = equalR.Next
		} else {
			if moreL == nil || moreR == nil {
				moreL = cur
				moreR = cur
			} else {
				moreR.Next = cur
				moreR = moreR.Next
			}
		}
		next := cur.Next
		cur.Next = nil
		cur = next
	}

	if lessR == nil && moreL == nil {
		return equalL
	} else if lessL == nil {
		equalR.Next = moreL
		return equalL
	} else if moreL == nil {
		lessR.Next = equalL
		return lessL
	} else {
		equalR.Next = moreL
		lessR.Next = equalL
		return lessL
	}
}
