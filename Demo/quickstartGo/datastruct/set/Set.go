package set

var Exists = struct{}{}

type Set interface {
	Contains(interface{}) bool
	Put(interface{})
	Remove(interface{}) error
	Clear()
	Equals(set Set) bool
	IsSubSet(father Set) bool
	IsEmpty() bool
	Size() int
}
