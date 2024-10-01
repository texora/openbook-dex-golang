package openbookdexgolang

import (
	"math/big"
)

const (
	MAXIMUM_TAKEN_ORDERS       = 45
	DROP_EXPIRED_ORDER_LIMIT   = 5
	FILL_EVENT_REMAINING_LIMIT = 5
)

const (
	Bid Side = iota
	Ask
)

type Orderbook struct {
	Bids *BookSide
	Asks *BookSide
}

type Amounts struct {
	TotalBaseTakenNative  uint64
	TotalQuoteTakenNative uint64
	Fee                   uint64
	NotEnoughLiquidity    bool
}

func (o *Orderbook) BookSide(side Side) *BookSide {
	switch side {
	case Bid:
		return o.Bids
	case Ask:
		return o.Asks
	default:
		return nil
	}
}

func AmountsFromBook(
	book Orderbook,
	side Side,
	maxBaseLots int64,
	maxQuoteLotsIncludingFees int64,
	market *Market,
	oraclePrice *big.Float,
	nowTs uint64,
) (Amounts, error) {

	// Handle the optional oracle price logic
	var oraclePriceLots *int64
	if oraclePrice != nil {
		priceLot, err := market.NativePriceToLot(oraclePrice)
		if err != nil {
			return Amounts{}, err
		}
		oraclePriceLots = &priceLot
	}

	// Placeholder for accounts array, if needed
	accounts := make([]interface{}, 0) // Adjust based on what the accounts array should hold

	// Call iterateBook, simulating the book iteration logic
	totalBaseLotsTaken, totalQuoteLotsTaken, makersRebates, notEnoughLiquidity := IterateBook(
		book,
		side,
		maxBaseLots,
		maxQuoteLotsIncludingFees,
		market,
		oraclePriceLots,
		nowTs,
		&accounts,
	)

	// Calculate total_base_taken_native and total_quote_taken_native
	totalBaseTakenNative := uint64(totalBaseLotsTaken * market.BaseLotSize)
	totalQuoteTakenNative := uint64(totalQuoteLotsTaken * market.QuoteLotSize)

	// Return the calculated Amounts struct
	return Amounts{
		TotalBaseTakenNative:  totalBaseTakenNative,
		TotalQuoteTakenNative: totalQuoteTakenNative,
		Fee:                   makersRebates,
		NotEnoughLiquidity:    notEnoughLiquidity,
	}, nil
}

func (s Side) InvertSide() Side {
	if s == Bid {
		return Ask
	}
	return Bid
}
