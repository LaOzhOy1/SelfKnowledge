package set

import "fmt"

// 这里struct{} 与 interface{} 的区别
// interface{} 默认开辟16个字节的空间 在set的map实现中，map的值是可以忽略的
// struct {}是一个无元素的结构体类型，通常在没有信息存储时使用。优点是大小为0，不需要内存来存储struct {}类型的值。
// struct {} {}是一个复合字面量，它构造了一个struct {}类型的值，该值也是空
//  Go 语言规范中的典型回答是：Go 语言字典的键类型不可以是函数类型、字典类型和切片类型。
//
//  Go 语言规范规定，在键类型的值之间必须可以施加操作符==和!=。

type HashSet struct {
	m map[interface{}]struct{}
}

func (h *HashSet) Contains(i interface{}) bool {
	_, ok := h.m[i]
	return ok
}

func (h *HashSet) Put(i interface{}) {
	h.m[i] = Exists
}

func (h *HashSet) Remove(i interface{}) error {
	delete(h.m, i)
	return nil
}

// 由于之前的map被替换为空指针指向，会被gc移除
func (h *HashSet) Clear() {
	h.m = make(map[interface{}]struct{})
}

func (h *HashSet) Equals(set Set) bool {
	if h.IsEmpty() && set.IsEmpty() {
		return true
	}
	if set.IsEmpty() || h.IsEmpty() {
		return false
	}
	if set.Size() != h.Size() {
		return false
	}
	for key, _ := range h.m {
		if !set.Contains(key) {
			return false
		}
	}
	return true
}

func (h *HashSet) IsSubSet(father Set) bool {
	if father == nil {
		return false
	}
	if h.Size() > father.Size() {
		return false
	}
	for key, _ := range h.m {
		if !father.Contains(key) {
			return false
		}
	}
	return true
}

func (h *HashSet) IsEmpty() bool {
	return h.Size() == 0
}

func (h *HashSet) Size() int {
	return len(h.m)
}

func (h *HashSet) Print() {
	for i, _ := range h.m {
		fmt.Println(i.(string))
	}
}

func NewHashSet() *HashSet {
	return &HashSet{m: map[interface{}]struct{}{}}
}
