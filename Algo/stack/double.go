package stack

type Double struct {
	Next *Double
	Pre  *Double
	val  int
}

type Stack struct {
	Top  *Double
	Size int
}

func NewStack() Stack {
	return Stack{}
}

func (s *Stack) Push(double *Double) {
	if s.Size == 0 {
		s.Top = double
		s.Size++
		return
	}

	s.Top.Next = double
	double.Pre = s.Top
	s.Top = double
	s.Size++
}

func (s *Stack) Pop() *Double {
	if s.Size == 0 {
		return nil
	}
	target := s.Top
	s.Top = s.Top.Pre
	target.Pre = nil
	s.Top.Next = nil
	return target
}
