package queue

type Double struct {
	Next *Double
	Pre  *Double
	val  int
}

type Queue struct {
	Head *Double
	Tail *Double
	Size int
}

func NewQueue() Queue {
	return Queue{}
}

func (q *Queue) Enqueue(double *Double) {
	if q.Size == 0 {
		q.Head = double
		q.Tail = double
		q.Size++
	}
	q.Tail.Next = double
	double.Pre = q.Tail
	q.Tail = q.Tail.Next
	q.Size++
}

func (q *Queue) Pop() {
	if q.Size == 0 {
		return
	}
	target := q.Head
	q.Head = q.Head.Next
	q.Head.Pre = nil
	target.Next = nil
	q.Size--
}
