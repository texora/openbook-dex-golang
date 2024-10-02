package openbookdexgolang

import (
	"math"
	"math/big"

	"github.com/gagliardetto/solana-go"
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
	accounts := make([]solana.PublicKey, 0) // Adjust based on what the accounts array should hold

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
		Fee:                   uint64(makersRebates),
		NotEnoughLiquidity:    notEnoughLiquidity,
	}, nil
}

func (s Side) InvertSide() Side {
	if s == Bid {
		return Ask
	}
	return Bid
}

func IterateBook(
	book Orderbook,
	side Side,
	maxBaseLots int64,
	maxQuoteLots int64,
	market *Market,
	oraclePriceLots *int64,
	nowTs uint64,
	accounts *[]solana.PublicKey,
) (int64, int64, int64, bool) {
	var limit = MAXIMUM_TAKEN_ORDERS
	var numberOfProcessedFillEvents = 0
	var numberOfDroppedExpiredOrders = 0

	var orderMaxBaseLots = maxBaseLots
	var orderMaxQuoteLots int64
	if side == Bid {
		orderMaxQuoteLots = market.SubtractTakerFees(maxQuoteLots)
	} else {
		orderMaxQuoteLots = maxQuoteLots
	}

	var makerRebatesAcc int64
	var remainingBaseLots = orderMaxBaseLots
	var remainingQuoteLots = orderMaxQuoteLots
	opposingBookSide := book.BookSide(side.InvertSide())
	iter := opposingBookSide.IterAllIncludingInvalid(nowTs, oraclePriceLots)
	for iter.Next() {
		bestOpposing := iter.Item()
		if !bestOpposing.IsValid() {
			if numberOfDroppedExpiredOrders < DROP_EXPIRED_ORDER_LIMIT {
				*accounts = append(*accounts, bestOpposing.Node.Owner)
				numberOfDroppedExpiredOrders++
			}
			continue
		}

		if remainingBaseLots == 0 || remainingQuoteLots == 0 || limit == 0 {
			break
		}

		bestOpposingPrice := bestOpposing.PriceLots
		maxMatchByQuote := remainingQuoteLots / bestOpposingPrice
		if maxMatchByQuote == 0 {
			break
		}

		matchBaseLots := int64(math.Min(float64(remainingBaseLots), float64(bestOpposing.Node.Quantity)))
		matchBaseLots = int64(math.Min(float64(matchBaseLots), float64(maxMatchByQuote)))
		matchQuoteLots := matchBaseLots * bestOpposingPrice

		makerRebatesAcc += int64(market.MakerRebateFloor(uint64(matchQuoteLots * market.QuoteLotSize)))

		remainingBaseLots -= matchBaseLots
		remainingQuoteLots -= matchQuoteLots

		limit--

		if numberOfProcessedFillEvents < FILL_EVENT_REMAINING_LIMIT {
			*accounts = append(*accounts, bestOpposing.Node.Owner)
			numberOfProcessedFillEvents++
		}
	}

	totalBaseLotsTaken := orderMaxBaseLots - remainingBaseLots
	totalQuoteLotsTaken := orderMaxQuoteLots - remainingQuoteLots

	notEnoughLiquidity := false
	if side == Ask {
		notEnoughLiquidity = remainingBaseLots != 0
	} else {
		notEnoughLiquidity = remainingQuoteLots != 0
	}

	return totalBaseLotsTaken, totalQuoteLotsTaken, makerRebatesAcc, notEnoughLiquidity
}
