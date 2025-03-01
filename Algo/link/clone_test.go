package link

import (
	"fmt"
	"testing"
)

type CloneLink struct {
	Next *CloneLink
	Val  int
}

// 哈希表实现

// 非哈希表实现

func TestCloneWithoutHashMap(t *testing.T) {
	var first *CloneLink
	for i := 0; i < 10; i++ {
		tmp := new(CloneLink)
		tmp.Val = 9 - i
		tmp.Next = first
		first = tmp
	}
	c := Clone(first)
	cur := first
	for c != nil && cur != nil {
		if c.Val != cur.Val {
			fmt.Println("is not Equal ,copy fail!")
			break
		}
		c = c.Next
		cur = cur.Next
	}
	fmt.Println("copy success")
}

func Clone(list *CloneLink) *CloneLink {
	cloneList := cloneList(list)
	return separate(cloneList)
}

func cloneList(list *CloneLink) *CloneLink {

	cur := list
	for cur != nil {
		tmp := new(CloneLink)
		tmp.Val = cur.Val

		next := cur.Next
		cur.Next = tmp
		tmp.Next = next

		cur = next
	}
	return cur
}
func separate(list *CloneLink) *CloneLink {
	copyList := new(CloneLink)
	copy := copyList
	isCopy := false

	cur := list
	pre := list
	for cur != nil {
		if !isCopy {
			isCopy = true
			pre = cur
			cur = cur.Next
			continue
		}

		copy.Next = cur
		copy = copy.Next

		next := cur.Next
		cur.Next = nil
		pre.Next = next

		isCopy = false
		cur = next
	}
	return copyList
}
