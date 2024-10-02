package openbookdexgolang

import "unsafe"

type OrderTreeType int

const (
	Bids OrderTreeType = iota
	Asks
)

type NodeTag int

const (
	uninitialized NodeTag = iota
	innerNode
	leafNode
	freeNode
	lastFreeNode
)

type OrderTreeNodes struct {
	OrderTreeType uint8
	Padding       [3]byte
	BumpIndex     uint32
	FreeListLen   uint32
	FreeListHead  NodeHandle
	Reserved      [512]byte
	Nodes         [MAX_ORDERTREE_NODES]AnyNode
}

func (o *OrderTreeNodes) order_tree_type() OrderTreeType {
	return OrderTreeType(o.OrderTreeType)
}

func (o *OrderTreeNodes) iter(root *OrderTreeRoot) *OrderTreeIter {
	return new(o, root)
}

func (o *OrderTreeNodes) node(handle NodeHandle) *AnyNode {
	node := &o.Nodes[int(handle)]
	tag := NodeTag(node.Tag)
	if tag == innerNode || tag == leafNode {
		return node
	}
	return nil
}

type OrderTreeRoot struct {
	MaybeNode NodeHandle
	LeafCount uint32
}

// Ensure the size of OrderTreeRoot is 8 bytes (similar to const_assert_eq in Rust)
func assertOrderTreeRootSize() {
	if unsafe.Sizeof(OrderTreeRoot{}) != 8 {
		panic("OrderTreeRoot size is not 8 bytes")
	}
}

func (o *OrderTreeRoot) node() *NodeHandle {
	if o.LeafCount == 0 {
		return nil
	}
	return &o.MaybeNode
}
