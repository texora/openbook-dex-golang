package openbookdexgolang

import "unsafe"

type OrderTreeIter struct {
	OrderTree *OrderTreeNodes // Pointer to OrderTreeNodes
	Stack     []*InnerNode    // Slice of pointers to InnerNode
	NextLeaf  *struct {
		handle NodeHandle
		leaf   *LeafNode
	} // Struct to hold NodeHandle and *LeafNode
	Left  int
	Right int
}

func new(orderTree *OrderTreeNodes, root *OrderTreeRoot) *OrderTreeIter {
	var left, right int
	if orderTree.order_tree_type() == Bids {
		left, right = 1, 0
	} else {
		left, right = 0, 1
	}

	iter := &OrderTreeIter{
		OrderTree: orderTree,
		Stack:     []*InnerNode{},
		NextLeaf:  nil,
		Left:      left,
		Right:     right,
	}

	if r := root.node(); r != nil {
		iter.NextLeaf = iter.findLeftmostLeaf(*r)
	}

	return iter
}

func (iter *OrderTreeIter) findLeftmostLeaf(start NodeHandle) *struct {
	handle NodeHandle
	leaf   *LeafNode
} {
	current := start
	for {
		node := iter.OrderTree.node(current)
		if node == nil {
			return nil
		}

		switch n := node.Case().Inner; n {
		case (*InnerNode)(unsafe.Pointer(&node)):
			iter.Stack = append(iter.Stack, n)
			current = n.Children[iter.Left]
		case nil:
			return &struct {
				handle NodeHandle
				leaf   *LeafNode
			}{
				handle: current,
				leaf:   nil,
			}
		}
	}
}
