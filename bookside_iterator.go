package openbookdexgolang

type BookSideIter struct {
	FixedIter        *OrderTreeIter // Pointer to OrderTreeIter
	OraclePeggedIter *OrderTreeIter // Pointer to OrderTreeIter
	NowTs            uint64         // Current timestamp
	OraclePriceLots  *int64         // Pointer to int64 to represent Option<i64>
}

func NewBookSideIter(bookSide *BookSide, nowTs uint64, oraclePriceLots *int64) *BookSideIter {
	return &BookSideIter{
		FixedIter:        bookSide.Nodes.Iter(bookSide.Root(FixedOrderTree)),
		OraclePeggedIter: bookSide.Nodes.Iter(bookSide.Root(OraclePeggedOrderTree)),
		NowTs:            nowTs,
		OraclePriceLots:  oraclePriceLots,
	}
}
