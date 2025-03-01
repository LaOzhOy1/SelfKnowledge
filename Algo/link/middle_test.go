package link

import (
	"fmt"
	"testing"
)

type middleList struct {
	Next *middleList
	Val  rune
}

func (l *middleList) println() {
	t := l
	for t != nil {
		fmt.Printf("%s, ", string(t.Val))
		t = t.Next
	}
	fmt.Println()
}

// 	对于奇数 长度的链表 返回 中点， 对于偶数长度的链表返回上中点

func TestMiddle(t *testing.T) {
	str := "abcdefgh"
	var first *middleList
	for _, s := range str {
		tmp := new(middleList)
		tmp.Next = first
		tmp.Val = s
		first = tmp
	}
	first.println()
	fmt.Println(string(Middle(first).Val))
}

func Middle(head *middleList) *middleList {
	if head == nil || head.Next == nil || head.Next.Next == nil {
		return head
	}
	slow := head.Next
	fast := head.Next.Next
	for fast.Next != nil && fast.Next.Next != nil {
		slow = slow.Next
		fast = fast.Next.Next
	}
	return slow
}

// 对于奇数 长度的链表 返回 中点， 对于偶数长度的链表返回下 中点
func TestMiddle1(t *testing.T) {

}

// 对于奇数 长度的链表 返回 中点 上一个， 对于偶数长度的链表返回上中点 上一个
func TestMiddle2(t *testing.T) {

}

// 对于奇数 长度的链表 返回 中点 上一个， 对于偶数长度的链表返回下 中点 上一个
func TestMiddle3(t *testing.T) {

}
