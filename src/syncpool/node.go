package syncpool

import "unsafe"

type Node struct {
	val  interface{}
	next unsafe.Pointer
	prev unsafe.Pointer
}

func NewNode(val interface{}) *Node {
	// TODO: Avoid allocating new node everytime
	n := &Node{
		val: val,
	}
	return n
}
