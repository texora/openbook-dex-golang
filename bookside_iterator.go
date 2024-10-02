package openbookdexgolang

type BookSideIter struct {
	FixedIter        *OrderTreeIter // Pointer to OrderTreeIter
	OraclePeggedIter *OrderTreeIter // Pointer to OrderTreeIter
	NowTs            uint64         // Current timestamp
	OraclePriceLots  *int64         // Pointer to int64 to represent Option<i64>
}

func newBookSideIter(bookSide *BookSide, nowTs uint64, oraclePriceLots *int64) *BookSideIter {
	return &BookSideIter{
		FixedIter:        bookSide.Nodes.iter(bookSide.root(FixedOrderTree)),
		OraclePeggedIter: bookSide.Nodes.iter(bookSide.root(OraclePeggedOrderTree)),
		NowTs:            nowTs,
		OraclePriceLots:  oraclePriceLots,
	}
}
