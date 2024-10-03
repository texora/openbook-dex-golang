package openbookdexgolang

import (
	"errors"
	"math"
	"unsafe"

	bin "github.com/gagliardetto/binary"
	"github.com/gagliardetto/solana-go"
)

type AnyNode struct {
	Tag        uint8
	Data       [79]byte
	ForceAlign uint64
}

type NodeRef struct {
	Inner *InnerNode
	Leaf  *LeafNode
}

type LeafNode struct {
	Tag           uint8
	OwnerSlot     uint8
	TimeInForce   uint16
	Padding       [4]byte
	Key           bin.Uint128
	Owner         solana.PublicKey
	Quantity      int64
	Timestamp     uint64
	PegLimit      int64
	ClientOrderID uint64
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

func (ln *LeafNode) PriceData() uint64 {
	return uint64(ln.Key.Hi)
}

func oraclePeggedPriceOffset(priceData uint64) int64 {
	// Wrapping subtract logic
	return int64(priceData - (math.MaxUint64/2 + 1))
}

func fixedPriceData(priceLots int64) (uint64, error) {
	// Security measure: Ensure priceLots is >= 1
	if priceLots < 1 {
		return 0, errors.New("price_lots must be greater than or equal to 1")
	}
	return uint64(priceLots), nil
}
