package set

import (
	"fmt"
	"sync"
)

type ReadWriteHashSet struct {
	m map[interface{}]struct{}
	// 该set可以调用RWMutex中的方法
	sync.RWMutex
	//mutex sync.RWMutex
}

func (h *ReadWriteHashSet) Contains(i interface{}) bool {
	h.RLock()
	defer h.RUnlock()
	_, ok := h.m[i]
	return ok
}

func (h *ReadWriteHashSet) Put(i interface{}) {
	h.Lock()
	defer h.Unlock()
	h.m[i] = Exists
}

func (h *ReadWriteHashSet) Remove(i interface{}) error {
	h.Lock()
	defer h.Unlock()
	delete(h.m, i)
	return nil
}

// 由于之前的map被替换为空指针指向，会被gc移除
func (h *ReadWriteHashSet) Clear() {
	h.m = make(map[interface{}]struct{})
}

func (h *ReadWriteHashSet) Equals(set Set) bool {
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

func (h *ReadWriteHashSet) IsSubSet(father Set) bool {
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

func (h *ReadWriteHashSet) IsEmpty() bool {
	return h.Size() == 0
}

func (h *ReadWriteHashSet) Size() int {
	return len(h.m)
}

func (h *ReadWriteHashSet) Print() {
	for i, _ := range h.m {
		fmt.Println(i.(string))
	}
}

func NewReadWriteHashSet() *ReadWriteHashSet {
	return &ReadWriteHashSet{m: map[interface{}]struct{}{}}
}
