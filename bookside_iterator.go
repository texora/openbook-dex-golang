package openbookdexgolang

import (
	"log"
	"math/big"
)

type BookSideIter struct {
	FixedIter        *OrderTreeIter // Pointer to OrderTreeIter
	OraclePeggedIter *OrderTreeIter // Pointer to OrderTreeIter
	NowTs            uint64         // Current timestamp
	OraclePriceLots  *int64         // Pointer to int64 to represent Option<i64>
}

type BookSideIterItem struct {
	Handle    BookSideOrderHandle // handle holds the order handle
	Node      *LeafNode           // node is a pointer to a LeafNode
	PriceLots int64               // priceLots represents the price in lots
	State     OrderState          // state indicates the order state
}

type BookSideOrderHandle struct {
	Node      NodeHandle
	OrderTree BookSideOrderTree
}

type OrderState int

const (
	Valid OrderState = iota
	Invalid
	Skipped
)

func newBookSideIter(bookSide *BookSide, nowTs uint64, oraclePriceLots *int64) *BookSideIter {
	return &BookSideIter{
		FixedIter:        bookSide.Nodes.iter(bookSide.root(FixedOrderTree)),
		OraclePeggedIter: bookSide.Nodes.iter(bookSide.root(OraclePeggedOrderTree)),
		NowTs:            nowTs,
		OraclePriceLots:  oraclePriceLots,
	}
}

func (item *BookSideIterItem) IsValid() bool {
	return item.State == Valid
}

func (iter *BookSideIter) Next() *BookSideIterItem {
	side := iter.FixedIter.Side()

	var oPeek *struct {
		handle NodeHandle
		leaf   *LeafNode
	}

	if iter.OraclePeggedIter != nil {
		oPeek = iter.OraclePeggedIter.Peek()
		for oPeek != nil {
			oNode := oPeek.leaf
			orderState, _ := oraclePeggedPrice(*(iter.OraclePriceLots), oNode, side)
			if orderState != Skipped {
				break
			}
			oPeek = iter.OraclePeggedIter.Next()
		}
	}

	fPeek := iter.FixedIter.Peek()

	better := rankOrders(
		side,
		fPeek,
		oPeek,
		false,
		iter.NowTs,
		iter.OraclePriceLots,
	)

	if better == nil {
		return nil
	}

	switch better.Handle.OrderTree {
	case FixedOrderTree:
		iter.FixedIter.Next()
	case OraclePeggedOrderTree:
		iter.OraclePeggedIter.Next()
	}

	return better
}

func oraclePeggedPrice(oraclePriceLots int64, node *LeafNode, side Side) (OrderState, int64) {
	priceData := node.PriceData()
	priceOffset := oraclePeggedPriceOffset(priceData)
	price := saturatingAdd(oraclePriceLots, priceOffset)

	if price >= 1 && price < (1<<63)-1 { // Equivalent to (1..i64::MAX) in Rust
		if node.PegLimit != -1 && side.IsPriceBetter(price, node.PegLimit) {
			return Invalid, price
		} else {
			return Valid, price
		}
	}
	if price < 1 {
		price = 1
	}
	return Skipped, price
}

func rankOrders(
	side Side,
	fixed *struct {
		handle NodeHandle
		leaf   *LeafNode
	},
	oraclePegged *struct {
		handle NodeHandle
		leaf   *LeafNode
	},
	returnWorse bool,
	nowTs uint64,
	oraclePriceLots *int64, // Simulate Option<i64>
) *BookSideIterItem {
	// Enrich oraclePegged if oracle_price_lots is present
	var oraclePegged1 *struct {
		handle NodeHandle
		leaf   *LeafNode
		price  int64
		state  OrderState
	}

	if oraclePriceLots != nil {
		state, price := oraclePeggedPrice(*oraclePriceLots, oraclePegged.leaf, side)
		oraclePegged1 = &struct {
			handle NodeHandle
			leaf   *LeafNode
			price  int64
			state  OrderState
		}{
			handle: oraclePegged.handle,
			leaf:   oraclePegged.leaf,
			price:  price,
			state:  state,
		}
	}

	// Determine ranking logic for fixed and oracle pegged
	if fixed != nil && oraclePegged1 != nil {
		isBetter := func(a, b uint64) bool {
			if side == Bid {
				return a > b
			}
			return a < b
		}

		oracleKey := keyForFixedPrice(oraclePegged1.leaf.Key.BigInt(), oraclePegged1.price)
		if isBetter(fixed.leaf.Key.BigInt().Uint64(), oracleKey.Uint64()) != returnWorse {
			return fixedToResult(fixed, nowTs)
		} else {
			return oraclePeggedToResult(oraclePegged1, nowTs)
		}
	} else if fixed == nil && oraclePegged1 != nil {
		return oraclePeggedToResult(oraclePegged1, nowTs)
	} else if fixed != nil && oraclePegged1 == nil {
		return fixedToResult(fixed, nowTs)
	} else {
		return nil
	}
}

func keyForFixedPrice(key *big.Int, priceLots int64) *big.Int {
	// We know this can never fail, because oracle pegged price will always be >= 1
	if priceLots < 1 {
		log.Fatal("priceLots must be >= 1")
	}
	priceData, _ := fixedPriceData(priceLots)
	if priceData == nil {
		log.Fatal("failed to get price data")
	}
	upper := uint64(*priceData) << 64
	lower := uint64(key.Uint64())
	return upper | lower
}

func fixedToResult(fixed *struct {
	handle NodeHandle
	leaf   *LeafNode
}, nowTs uint64) *BookSideIterItem {
	handle, node := fixed

	// Check if the node is expired
	expired := node.IsExpired(nowTs)

	// Create and return the result
	return &BookSideIterItem{
		Handle: BookSideOrderHandle{
			OrderTree: Fixed,
			Node:      handle,
		},
		Node:      node,
		PriceLots: int64(node.Key), // Assuming PriceData is stored in node.Key (modify if necessary)
		State:     getOrderState(expired),
	}
}

func oraclePeggedToResult(pegged *struct {
	handle NodeHandle
	leaf   *LeafNode
	price  int64
	state  OrderState
}, nowTs uint64) *BookSideIterItem {
	handle, node, priceLots, state := pegged

	// Check if the node is expired
	expired := node.IsExpired(nowTs)

	// Create and return the result
	return &BookSideIterItem{
		Handle: BookSideOrderHandle{
			OrderTree: OraclePegged,
			Node:      handle,
		},
		Node:      node,
		PriceLots: priceLots,
		State:     getOrderStateForPegged(expired, state),
	}
}
