package link

type Single struct {
	Next *Single
}

type Double struct {
	Next *Double
	Pre  *Double
}

// review done
func reverseSingle(link *Single) {
	pre := new(Single)
	next := new(Single)

	for link.Next != nil {
		next = link.Next
		link.Next = pre
		pre = link

		link = next
	}

	for link.Next != nil {
		next = link.Next
		link.Next = pre
		pre = link
		link = next
	}
}

func reverseDouble(link *Double) {
	pre := new(Double)
	next := new(Double)

	for link.Next != nil {
		next = link.Next
		link.Next = pre
		link.Pre = next
		pre = link

		link = next
	}
}
