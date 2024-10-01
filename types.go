package openbookdexgolang

import (
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

type BookSide struct {
	Roots         [2]OrderTreeRoot
	ReservedRoots [4]OrderTreeRoot
	Reserved      [256]byte
	Nodes         OrderTreeNodes
}

type Orderbook struct {
	Bids *BookSide
	Asks *BookSide
}

type EventHeapHeader struct {
	FreeHead uint16
	UsedHead uint16
	Count    uint16
	Padd     uint16
	SeqNum   uint64
}

type OrderTreeRoot struct {
	MaybeNode NodeHandle
	LeafCount uint32
}

type OrderTreeNodes struct {
	OrderTreeType uint8
	Padding       [3]byte
	BumpIndex     uint32
	FreeListLen   uint32
	FreeListHead  NodeHandle
	Reserved      [512]byte
	Nodes         [MAX_ORDERTREE_NODES]AnyNode
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

type AnyNode struct {
	Tag        uint8
	Data       [79]byte
	ForceAlign uint64
}

type Decimal struct {
	Flags uint32
	Hi    uint32
	Lo    uint32
	Mid   uint32
}

type Side int
