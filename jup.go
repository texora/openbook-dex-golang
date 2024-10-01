package openbookdexgolang

import (
	bin "github.com/gagliardetto/binary"
	"github.com/gagliardetto/solana-go"
)

const (
	SideBid Side = iota // Starts from 0
	SideAsk             // Automatically 1
)

type OpenBookMarket struct {
	market          Market
	eventHeap       EventHeap
	bids            BookSide
	asks            BookSide
	timestamp       uint64
	key             solana.PublicKey
	label           string
	relatedAccounts []solana.PublicKey
	reserveMints    [2]solana.PublicKey
	oraclePrice     *bin.Int128
	isPermissioned  bool
}

type QuoteParams struct {
	InAmount   uint64
	InputMint  solana.PublicKey
	OutputMint solana.PublicKey
}

type Quote struct {
	NotEnoughLiquidity bool
	MinInAmount        *uint64 // Option<u64> is represented as a pointer
	MinOutAmount       *uint64 // Option<u64> is represented as a pointer
	InAmount           uint64
	OutAmount          uint64
	FeeAmount          uint64
	FeeMint            solana.PublicKey
	FeePct             Decimal
}

func (obm *OpenBookMarket) Quote(quoteParams *QuoteParams) (*Quote, error) {
	// Check if the market is permissioned
	if obm.isPermissioned {
		return &Quote{
			NotEnoughLiquidity: true,
			// Default other fields if needed (you may need to define default behavior for Quote)
		}, nil
	}

	// Determine the side based on input mint
	var side Side
	if quoteParams.InputMint == obm.market.QuoteMint {
		side = SideBid
	} else {
		side = SideAsk
	}

	// Convert input amount to int64
	inputAmount := int64(quoteParams.InAmount)

	// Calculate max base lots and max quote lots including fees
	var maxBaseLots, maxQuoteLotsIncludingFees int64
	switch side {
	case SideBid:
		maxBaseLots = obm.market.MaxBaseLots()
		maxQuoteLotsIncludingFees = (inputAmount + int64(obm.market.QuoteLotSize) - 1) / int64(obm.market.QuoteLotSize)
	case SideAsk:
		maxBaseLots = (inputAmount + int64(obm.market.BaseLotSize) - 1) / int64(obm.market.BaseLotSize)
		maxQuoteLotsIncludingFees = obm.market.MaxQuoteLots()
	}

	// Use Go references to create mutable order book references
	bidsRef := obm.bids
	asksRef := obm.asks
	book := Orderbook{
		Bids: &bidsRef,
		Asks: &asksRef,
	}

	// Calculate order amounts from the order book
	orderAmounts, err := amountsFromBook(
		book,
		side,
		maxBaseLots,
		maxQuoteLotsIncludingFees,
		&obm.market,
		obm.oraclePrice,
		0,
	)
	if err != nil {
		return nil, err
	}

	// Calculate in_amount and out_amount based on the side
	var inAmount, outAmount int64
	switch side {
	case SideBid:
		inAmount = orderAmounts.TotalQuoteTakenNative - orderAmounts.Fee
		outAmount = orderAmounts.TotalBaseTakenNative
	case SideAsk:
		inAmount = orderAmounts.TotalBaseTakenNative
		outAmount = orderAmounts.TotalQuoteTakenNative + orderAmounts.Fee
	}

	// Return the quote
	return &Quote{
		InAmount:           uint64(inAmount),
		OutAmount:          uint64(outAmount),
		FeeMint:            obm.market.QuoteMint,
		FeeAmount:          orderAmounts.Fee,
		NotEnoughLiquidity: orderAmounts.NotEnoughLiquidity,
		// You can initialize other fields of Quote here as needed
	}, nil
}
