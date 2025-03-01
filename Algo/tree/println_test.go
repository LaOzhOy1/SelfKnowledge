package tree

import (
	"fmt"
	"testing"
)

type PrintlnTree struct {
	Val   int
	Left  *PrintlnTree
	Right *PrintlnTree
}

type PrintlnTreeQueue struct {
	Data []*PrintlnTree
	Size int
}

func (q *PrintlnTreeQueue) Push(p *PrintlnTree) {
	q.Data = append(q.Data, p)
	q.Size++
}

func (q *PrintlnTreeQueue) Pop() *PrintlnTree {
	if q.Size <= 0 {
		return nil
	}
	q.Size--
	target := q.Data[0]
	q.Data = q.Data[1:]
	return target
}
func NewQueue() *PrintlnTreeQueue {
	return &PrintlnTreeQueue{
		Data: nil,
		Size: 0,
	}
}
func NewTree(n int) *PrintlnTree {
	list := make([]*PrintlnTree, 0, 32)
	for i := 1; i <= n; i++ {
		list = append(list, &PrintlnTree{
			Val: i,
		})
	}
	idx := 0
	var head = list[0]
	queue := NewQueue()
	for idx < len(list) {
		if queue.Size == 0 {
			queue.Push(list[idx])
		} else {
			tree := queue.Pop()
			idx++
			if idx >= len(list) {
				break
			}
			tree.Left = list[idx]
			queue.Push(list[idx])
			idx++
			if idx >= len(list) {
				break
			}
			tree.Right = list[idx]
			queue.Push(list[idx])
		}
	}
	return head
}

func TestPrintlnTree(t *testing.T) {
	root := NewTree(3)
	RecursionPrintlnMid(root)

}

func TestPrintlnTree2(t *testing.T) {
	root := NewTree(3)
	//IterateWithPre(root)
	IterateWithMid(root)
}

// 递归实现
func RecursionPrintlnMid(root *PrintlnTree) {
	if root == nil {
		return
	}
	RecursionPrintlnMid(root.Left)
	fmt.Printf("%d ", root.Val)
	RecursionPrintlnMid(root.Right)

}

type PrintlnStack struct {
	Data []*PrintlnTree
	Top  int
}

func (p *PrintlnStack) Push(val *PrintlnTree) {
	if p.Top < len(p.Data) {
		p.Data[p.Top] = val
	} else {
		p.Data = append(p.Data, val)
	}
	p.Top++
}
func (p *PrintlnStack) Pop() *PrintlnTree {
	if p.Top <= 0 {
		return nil
	}
	p.Top--
	return p.Data[p.Top]
}

func (p *PrintlnStack) Peek() *PrintlnTree {
	if p.Top <= 0 {
		return nil
	}
	return p.Data[p.Top-1]
}

func NewPrintlnStack() *PrintlnStack {
	return &PrintlnStack{}
}

// 先序遍历

func IterateWithPre(root *PrintlnTree) {
	stack := NewPrintlnStack()
	stack.Push(root)
	for stack.Top > 0 {
		tree := stack.Pop()
		if tree.Right != nil {
			stack.Push(tree.Right)
		}
		if tree.Left != nil {
			stack.Push(tree.Left)
		}
		fmt.Printf("%d ", tree.Val)
	}
}

// 栈实现 后序遍历 把先序再加一个栈倒过来

// 栈实现 中序编历

func IterateWithMid(root *PrintlnTree) {
	stack := NewPrintlnStack()
	stack.Push(root)
	pre := root

	for stack.Top > 0 {
		c := stack.Peek()
		if c.Left != nil && c.Left != pre && c.Right != pre {
			stack.Push(stack.Peek().Left)
			continue
		}
		tree := stack.Pop()
		fmt.Printf("%d ", tree.Val)
		if tree.Right != nil {
			stack.Push(tree.Right)
		} else {
			pre = tree
		}

	}
}

// 打印二叉树

// 打印对折纸条凹凸明细
