package main

type TreeNode struct {
	Val   int
	Left  *TreeNode
	Right *TreeNode
}

// 给定一个二叉树，检查它是否是镜像对称的。  队列
// 取root 两次装入 队列中 模拟左右节点
// 判断左右节点 中 镜像节点是否相等
func main() {

}

func isSymmetric(root *TreeNode) bool {

	if root == nil {
		return false
	}
	return isMirror(root.Left, root.Right)

}

func isMirror(left *TreeNode, right *TreeNode) bool {
	if left == nil && right == nil {
		return true
	}
	if left == nil || right == nil {
		return false
	}
	if left.Val != right.Val {
		return false
	}
	return isMirror(left.Left, right.Right) && isMirror(left.Right, right.Left)
}
