package main

type TreeNode struct {
	value       int
	left, right *TreeNode
}

type NewTreeNode struct {
	node *TreeNode
}

func (myNode *NewTreeNode) postPrint(f func(node *NewTreeNode)) {
	if myNode == nil || myNode.node == nil {
		return
	}

	//var left = NewTreeNode{myNode.node.left}
	myNode.postPrint(f)
	//var right = NewTreeNode{myNode.node.right}
	myNode.postPrint(f)
	f(myNode)
}

func main() {

}
