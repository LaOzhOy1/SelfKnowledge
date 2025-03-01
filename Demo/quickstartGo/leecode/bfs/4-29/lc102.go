package main

import "fmt"

//给你一个二叉树，请你返回其按 层序遍历 得到的节点值。 （即逐层地，从左到右访问所有节点）。

type TreeNode struct {
	Val   int
	Left  *TreeNode
	Right *TreeNode
}

func levelOrder(root *TreeNode) [][]int {

	var result [][]int
	arr := []*TreeNode{root}
	//arr := make([]*TreeNode,0)
	//arr = append(arr, root)
	t := 0
	for len(arr) > 0 {
		var tmpArr []*TreeNode
		result = append(result, []int{})
		for i := 0; i < len(arr); i++ {
			tmp := arr[i]

			// 改变了原来的数组长度
			//arr = arr[1:]
			result[t] = append(result[t], tmp.Val)
			if tmp.Left != nil {
				tmpArr = append(tmpArr, tmp.Left)
				fmt.Println(tmp.Left)
			}

			if tmp.Right != nil {
				tmpArr = append(tmpArr, tmp.Right)
				fmt.Println(tmp.Right)
			}
		}
		arr = tmpArr
		t++
	}
	return result
}
