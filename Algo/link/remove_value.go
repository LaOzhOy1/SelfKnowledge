package link

type Value struct {
	Next   *Value
	Number int
}

func removeValue(head *Value, target int) {
	pre := new(Value)
	cur := new(Value)

	cur = head
	for cur != nil && cur.Number == target {
		cur = cur.Next
	}
	if cur == nil {
		return
	}
	pre = cur
	cur = cur.Next
	for cur != nil {
		if cur.Number == target {
			pre.Next = cur.Next
			cur = cur.Next
		} else {
			pre = cur
			cur = cur.Next
		}
	}
}

//  不提供 头结点 无法删除最后的节点

func RemoveByCertainNode(node *Value) {
	if node.Next == nil {
		return
	}
	node.Number = node.Next.Number
	node.Next = node.Next.Next
}
