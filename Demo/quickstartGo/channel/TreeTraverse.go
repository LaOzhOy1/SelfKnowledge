package main

import (
	"fmt"
	"time"
)

type TreeNode struct {
	val   int
	left  *TreeNode
	right *TreeNode
}

func initTree(node *TreeNode) {
	arr := []int{1, 2, 3, 4, 5, 6}

	if node == nil {
		fmt.Println("nil")
		node = &TreeNode{
			val:   0,
			left:  nil,
			right: nil,
		}
	}
	head := node
	tmp := node
	last := node
	for i := 0; i < len(arr); i++ {

		for tmp != nil {
			if arr[i] >= tmp.val {
				last = tmp
				tmp = tmp.right
			} else {
				last = tmp
				tmp = tmp.left
			}
		}

		tmp = &TreeNode{
			val:   arr[i],
			left:  nil,
			right: nil,
		}
		if tmp.val < last.val {
			last.left = tmp
		} else {
			last.right = tmp
		}
		tmp = head
		last = head
	}
}

func travelMid(node *TreeNode) {
	if node == nil {
		return
	}
	travelMid(node.left)
	fmt.Println(node.val)
	travelMid(node.right)
}

func (node *TreeNode) travelMidX() {
	if node == nil {
		return
	}
	travelMid(node.left)
	fmt.Println(node.val)
	travelMid(node.right)
}

func (node *TreeNode) travelMidXWithFunc(f func(node *TreeNode)) {
	if node == nil {
		return
	}
	node.left.travelMidXWithFunc(f)
	f(node)
	node.right.travelMidXWithFunc(f)
}

func travelByMidWithChannel(node *TreeNode) chan *TreeNode {
	out := make(chan *TreeNode)
	go func() {
		node.travelMidXWithFunc(func(node *TreeNode) {
			//fmt.Println(node.val)
			//if node!=nil {
			out <- node
			//}
		})
		// 子携程 没有关闭 将会导致 父携程阻塞无法完成  相反父携程跑完 子携程channel自动释放
		close(out)
	}()
	return out
}

func main() {
	head := new(TreeNode)
	//fmt.Println(head)
	initTree(head)
	//travelMid(head)
	//head.travelMidX()
	channel := travelByMidWithChannel(head)
	for c := range channel {
		fmt.Println(c.val)
	}
	time.Sleep(1000)
}
