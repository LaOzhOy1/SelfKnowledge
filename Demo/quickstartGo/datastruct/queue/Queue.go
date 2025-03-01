package queue

/*
*
参考 实现 https://blog.csdn.net/qq_36027670/article/details/105817450
*/
type Queue interface {
	Offer(e interface{})
	Poll() interface{}
	Peek() interface{}
	Clear()
	Size() int
	IsEmpty() bool
}
