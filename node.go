package openbookdexgolang

import "unsafe"

type AnyNode struct {
	Tag        uint8
	Data       [79]byte
	ForceAlign uint64
}

type NodeRef struct {
	Inner *InnerNode
	Leaf  *LeafNode
}

func (node *AnyNode) Case() *NodeRef {
	tag := NodeTag(node.Tag)

	switch tag {
	case innerNode:
		return &NodeRef{
			Inner: (*InnerNode)(unsafe.Pointer(&node)),
			Leaf:  nil,
		}
	case leafNode:
		return &NodeRef{
			Inner: nil,
			Leaf:  (*LeafNode)(unsafe.Pointer(&node)),
		}
	default:
		return nil
	}
}
