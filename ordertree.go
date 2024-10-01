package openbookdexgolang

type OrderTreeNodes struct {
	OrderTreeType uint8
	Padding       [3]byte
	BumpIndex     uint32
	FreeListLen   uint32
	FreeListHead  NodeHandle
	Reserved      [512]byte
	Nodes         [MAX_ORDERTREE_NODES]AnyNode
}
