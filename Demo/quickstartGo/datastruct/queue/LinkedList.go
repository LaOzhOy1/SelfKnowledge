package queue

/*
*
 */
type Node struct {
	data interface{}
	next *Node
}

type LinkedList struct {
	head *Node
	tail *Node
	size int
}

func (l *LinkedList) Offer(e interface{}) {
	if l.IsEmpty() {
		//赋值时需要将地址传递给指针
		l.tail = &Node{
			data: e,
			next: nil,
		}
		l.head = l.tail
		l.size++
		return
	}
	l.tail.next = &Node{
		data: e,
		next: nil,
	}
	l.tail = l.tail.next
	l.size++
}

// 返回值时不用显示指定指针返回值
func (l *LinkedList) Poll() interface{} {
	if l.IsEmpty() {
		return nil
	}
	node := l.head
	l.head = l.head.next
	l.size--
	return node
}

func (l *LinkedList) Peek() interface{} {
	if l.IsEmpty() {
		return nil
	}
	return l.head
}

// 根据gc 理论 要将指针指向的元素置为nil
func (l *LinkedList) Clear() {
	if l.IsEmpty() {
		return
	}
	p := l.head
	for p != nil {
		tmp := p.next
		p = nil
		p = tmp
	}
	l.size = 0
}

func (l *LinkedList) Size() int {
	return l.size
}

func (l *LinkedList) IsEmpty() bool {
	return l.size == 0
}
