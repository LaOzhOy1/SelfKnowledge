package link

import (
	"fmt"
	"testing"
)

type LinkString struct {
	Next *LinkString
	Val  rune
}

func (l *LinkString) println() {
	t := l
	for t != nil {
		fmt.Printf("%d, ", t.Val)
		t = t.Next
	}
	fmt.Println()
}

// 判断链表回文串 通过Stack

// 通过快慢指针优化

// O(1) 优化

func TestPalindromicString2(t *testing.T) {
	str := "abcdcba"
	var first *LinkString
	for _, s := range str {
		tmp := new(LinkString)
		tmp.Next = first
		tmp.Val = s
		first = tmp
	}
	first.println()
	fmt.Println(isPalindromicString2(first))
}

func isPalindromicString2(link *LinkString) bool {

	if link == nil || link.Next == nil || link.Next.Next == nil {
		return true
	}
	slow := link.Next
	fast := link.Next.Next

	for fast.Next != nil && fast.Next.Next != nil {
		slow = slow.Next
		fast = fast.Next.Next
	}

	var pre *LinkString
	reverse := slow

	for reverse != nil {
		next := reverse.Next
		reverse.Next = pre
		pre = reverse
		reverse = next
	}

	cmp := fast
	cmp2 := link
	for cmp != nil && cmp2 != nil {
		if cmp.Val != cmp2.Val {
			return false
		}
		cmp = cmp.Next
		cmp2 = cmp2.Next
	}
	return true
}
