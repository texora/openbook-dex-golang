package openbookdexgolang

type BookSide struct {
	Roots         [2]OrderTreeRoot
	ReservedRoots [4]OrderTreeRoot
	Reserved      [256]byte
	Nodes         OrderTreeNodes
}

type BookSideOrderTree int

const (
	FixedOrderTree        BookSideOrderTree = iota // Fixed = 0
	OraclePeggedOrderTree                          // OraclePegged = 1
)
