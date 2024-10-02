package openbookdexgolang

import (
	bin "github.com/gagliardetto/binary"
	"github.com/gagliardetto/solana-go"
)

const (
	MAX_ORDERTREE_NODES = 1024
	MAX_NUM_EVENTS      = 600
)

type NonZeroPubkeyOption struct {
	Key solana.PublicKey
}

type OracleConfig struct {
	ConfFilter        float64
	MaxStalenessSlots int64
	Reserved          [72]byte
}

type EventHeap struct {
	Header   EventHeapHeader
	Nodes    [MAX_NUM_EVENTS]EventNode
	Reserved [64]byte
}

type EventHeapHeader struct {
	FreeHead uint16
	UsedHead uint16
	Count    uint16
	Padd     uint16
	SeqNum   uint64
}

// LeafNodeWithHandle is a helper struct to store NodeHandle and LeafNode reference
type LeafNodeWithHandle struct {
	Handle   NodeHandle
	LeafNode *LeafNode
}

type EventNode struct {
	Next  uint16
	Prev  uint16
	Pad   [4]byte
	Event AnyEvent
}

type AnyEvent struct {
	EventType uint8
	Padding   [143]byte
}

type NodeHandle uint32
type InnerNode struct {
	Tag                 uint8
	Padding             [3]byte
	PrefixLen           uint32
	Key                 bin.Uint128
	Children            [2]NodeHandle
	ChildEarliestExpiry [2]uint64
	Reserved            [40]byte
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

type Decimal struct {
	Flags uint32
	Hi    uint32
	Lo    uint32
	Mid   uint32
}

type Side int
