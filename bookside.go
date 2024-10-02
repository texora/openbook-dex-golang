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

func (b *BookSide) root(component BookSideOrderTree) *OrderTreeRoot {
	return &b.Roots[int(component)]
}

func (b *BookSide) IterAllIncludingInvalid(nowTs uint64, oraclePriceLots *int64) *BookSideIter {
	return newBookSideIter(b, nowTs, oraclePriceLots)
}
