package tree

import (
	"fmt"
	"testing"
)

type TireTree struct {
	Pass int
	End  int
	Next []*TireTree
}

func NewTireTree() *TireTree {
	t := new(TireTree)
	t.Next = make([]*TireTree, 26)
	return t
}

func Insert(root *TireTree, str string) {
	for _, c := range []rune(str) {
		i := int(c) - int('a')
		if root.Next[i] == nil {
			root.Next[i] = NewTireTree()
		}
		root = root.Next[i]
		root.Pass++
	}
	root.End++
}

func CountString(root *TireTree, str string) {
	for _, c := range []rune(str) {
		i := int(c) - int('a')
		root = root.Next[i]
	}
	fmt.Println(root.Pass)
	fmt.Println(root.End)
}

// 构造前缀树， 查询以 * 前缀的信息
func TestTrie(t *testing.T) {
	abc := "abc"
	abcd := "abcd"
	root := NewTireTree()
	Insert(root, abc)
	Insert(root, abcd)
	CountString(root, "abcd")
}
