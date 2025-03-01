package main

type Stack struct {
	arr []int
}

func (s *Stack) insert(value int) {
	if value > s.arr[s.size()-1] {
		s.arr = append(s.arr, value)
	} else {
		for i := 0; i < s.size(); i++ {
		}
	}

}

func (s *Stack) size() int {
	return len(s.arr)
}

func main() {

}
