package link

import (
	"fmt"
	"testing"
)

type isX struct {
	Next *isX
}

// 判断两条链表是否相交， 如果相交返回相交的第一个点
func TestIsX(t *testing.T) {
	var first *isX
	var tail *isX
	for i := 0; i < 10; i++ {
		tmp := new(isX)
		tmp.Next = first
		first = tmp
		if i == 0 {
			tail = first
		}
		if i == 7 {
			tail.Next = first
		}
	}

	var second *isX
	var tail2 *isX
	for i := 0; i < 2; i++ {
		tmp := new(isX)
		tmp.Next = second
		second = tmp
		if i == 0 {
			tail2 = second
		}
	}

	tail2.Next = tail
	//tail2.Next = first.Next.Next
	fmt.Println(IsX(first, second))
}

func IsX(x *isX, y *isX) bool {
	x1 := isO(x)
	x2 := isO(y)
	if x1 == nil && x2 == nil {
		t1 := tail(x1)
		t2 := tail(x2)
		if t1 == t2 {
			return true
		} else {
			return false
		}
	} else {
		if x1 == nil || x2 == nil {
			return false
		}
		if x1 == x2 {
			return true
		}
		cur := x1.Next
		for cur != x1 {
			if cur == x2 {
				return true
			}
			cur = cur.Next
		}
		return false
	}
}
func tail(head *isX) *isX {
	cur := head
	if cur == nil {
		return cur
	}
	for cur.Next != nil {
		cur = cur.Next
	}
	return cur
}

func TestIsO(t *testing.T) {
	var first *isX
	//var tail *isX
	var insert *isX
	for i := 0; i < 10; i++ {
		tmp := new(isX)
		tmp.Next = first
		first = tmp
		//if i == 0 {
		//	tail = first
		//}
		//if i == 7 {
		//	tail.Next = first
		//	insert = first
		//}
	}
	x := isO(first)
	if x != nil {
		fmt.Println("is o insert == x ? :", x == insert)
	} else {
		fmt.Println("no o")
	}
}

func isO(list *isX) *isX {
	if list == nil || list.Next == nil || list.Next.Next == nil {
		return nil
	}
	slow := list.Next
	fast := list.Next.Next

	for fast.Next != nil && fast.Next.Next != nil && slow != fast {
		slow = slow.Next
		fast = fast.Next.Next
	}
	if fast.Next == nil || fast.Next.Next == nil {
		return nil
	}

	cur := list
	for cur != fast {
		fast = fast.Next
		cur = cur.Next
	}
	return cur
}
