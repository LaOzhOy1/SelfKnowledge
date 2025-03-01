package queue

import "fmt"

/**
每次取出操作都会新建一个切片

*/

type ArrayList struct {
	items []interface{}
}

func (list *ArrayList) Peek() interface{} {

	if list.IsEmpty() {
		return nil
	}
	return list.items[0]
}

func (list *ArrayList) Poll() interface{} {
	if list.IsEmpty() {
		return nil
	}
	tmp := list.items[0]
	list.items = list.items[1:]
	return tmp
}

func (list *ArrayList) Offer(e interface{}) {
	list.items = append(list.items, e)
}

// 由于之前的slice被替换为空指针指向，会被gc移除
func (list *ArrayList) Clear() {
	list.items = make([]interface{}, 0)
}

func (list *ArrayList) Size() int {
	//panic("implement me")
	return len(list.items)
}

func (list *ArrayList) IsEmpty() bool {
	return list.Size() == 0
}

// 返回*和值的区别 进行方法调用的时候区别不大，，传值时可能要显示调用
func NewArrayList() *ArrayList {
	return &ArrayList{}
}

func (list *ArrayList) Print() {
	if list.IsEmpty() {
		return
	}
	for i := 0; i < len(list.items); i++ {
		fmt.Print(list.items[i].(string), " ")
	}
	fmt.Println()
}
