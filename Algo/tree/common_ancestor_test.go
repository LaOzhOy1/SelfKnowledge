package tree

import (
	"fmt"
	"testing"
)

// 返回一颗树的共有祖先

type AncestorTree struct {
	Val   int
	Left  *AncestorTree
	Right *AncestorTree
}

type AncestorTreeQueue struct {
	Data []*AncestorTree
	Size int
}

func (q *AncestorTreeQueue) Push(p *AncestorTree) {
	q.Data = append(q.Data, p)
	q.Size++
}

func (q *AncestorTreeQueue) Pop() *AncestorTree {
	if q.Size <= 0 {
		return nil
	}
	q.Size--
	target := q.Data[0]
	q.Data = q.Data[1:]
	return target
}
func NewAncestorTreeQueue() *AncestorTreeQueue {
	return &AncestorTreeQueue{
		Data: nil,
		Size: 0,
	}
}

func NewAncestorTree(n int) *AncestorTree {
	list := make([]*AncestorTree, 0, 32)
	for i := 1; i <= n; i++ {
		list = append(list, &AncestorTree{
			Val: i,
		})
	}
	idx := 0
	var head = list[0]
	queue := NewAncestorTreeQueue()
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

func TestAncestorTree(t *testing.T) {
	root := NewAncestorTree(13)
	fmt.Println(CommonAncestor(root, 7, 13).Tree.Val)
}

// 相同祖先的情况
// 1） 相同分支
// 2） 左右分支
// 3） 自己是祖先
// 4） 两个分支都没有
// 需要左右节点返回 1. 是否找到节点, 2, 找到的节点值

type AncestorInfo struct {
	Tree  *AncestorTree
	Find1 bool
	Find2 bool
}

func CommonAncestor(tree *AncestorTree, target1 int, target2 int) AncestorInfo {
	if tree == nil {
		return AncestorInfo{
			Tree:  nil,
			Find1: false,
			Find2: false,
		}
	}

	lInfo := CommonAncestor(tree.Left, target1, target2)
	rInfo := CommonAncestor(tree.Right, target1, target2)

	if lInfo.Tree != nil {
		return lInfo
	}
	if rInfo.Tree != nil {
		return rInfo
	}

	if lInfo.Find1 && rInfo.Find2 || lInfo.Find2 && rInfo.Find1 || tree.Val == target2 && tree.Val == target1 {
		return AncestorInfo{
			Tree:  tree,
			Find1: true,
			Find2: true,
		}
	}

	if lInfo.Find1 && !rInfo.Find2 || rInfo.Find1 && !lInfo.Find2 {
		if tree.Val == target2 {
			return AncestorInfo{
				Tree:  tree,
				Find1: true,
				Find2: true,
			}
		} else {
			return AncestorInfo{
				Tree:  nil,
				Find1: true,
				Find2: false,
			}
		}

	} else if lInfo.Find2 && !rInfo.Find1 || rInfo.Find2 && !lInfo.Find1 {
		if tree.Val == target1 {
			return AncestorInfo{
				Tree:  tree,
				Find1: true,
				Find2: true,
			}
		} else {
			return AncestorInfo{
				Tree:  nil,
				Find1: false,
				Find2: true,
			}
		}
	}

	return AncestorInfo{
		Tree:  nil,
		Find1: target1 == tree.Val,
		Find2: target2 == tree.Val,
	}
}
