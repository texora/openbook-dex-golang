package openbookdexgolang

import (
	"errors"
	"math"
	"math/big"

	bin "github.com/gagliardetto/binary"
	"github.com/gagliardetto/solana-go"
)

const FEES_SCALE_FACTOR = 1000000

type Market struct {
	// PDA bump
	Bump uint8

	// Number of decimals used for the base token.
	//
	// Used to convert the oracle's price into a native/native price.
	BaseDecimals  uint8
	QuoteDecimals uint8

	Padding1 [5]byte

	// Pda for signing vault txs
	MarketAuthority solana.PublicKey

	// No expiry = 0. Market will expire and no trading allowed after time_expiry
	TimeExpiry int64

	// Admin who can collect fees from the market
	CollectFeeAdmin solana.PublicKey
	// Admin who must sign off on all order creations
	OpenOrdersAdmin NonZeroPubkeyOption
	// Admin who must sign off on all event consumptions
	ConsumeEventsAdmin NonZeroPubkeyOption
	// Admin who can set market expired, prune orders and close the market
	CloseMarketAdmin NonZeroPubkeyOption

	// Name. Trailing zero bytes are ignored.
	Name [16]byte

	// Address of the BookSide account for bids
	Bids solana.PublicKey
	// Address of the BookSide account for asks
	Asks solana.PublicKey
	// Address of the EventHeap account
	EventHeap solana.PublicKey

	// Oracles account address
	OracleA NonZeroPubkeyOption
	OracleB NonZeroPubkeyOption
	// Oracle configuration
	OracleConfig OracleConfig

	// Number of quote native in a quote lot. Must be a power of 10.
	//
	// Primarily useful for increasing the tick size on the market: A lot price
	// of 1 becomes a native price of quote_lot_size/base_lot_size becomes a
	// ui price of quote_lot_size*base_decimals/base_lot_size/quote_decimals.
	QuoteLotSize int64

	// Number of base native in a base lot. Must be a power of 10.
	//
	// Example: If base decimals for the underlying asset is 6, base lot size
	// is 100 and and base position lots is 10_000 then base position native is
	// 1_000_000 and base position ui is 1.
	BaseLotSize int64

	// Total number of orders seen
	SeqNum uint64

	// Timestamp in seconds that the market was registered at.
	RegistrationTime int64

	// Fees
	//
	// Fee (in 10^-6) when matching maker orders.
	// maker_fee < 0 it means some of the taker_fees goes to the maker
	// maker_fee > 0, it means no taker_fee to the maker, and maker fee goes to the referral
	MakerFee int64
	// Fee (in 10^-6) for taker orders, always >= 0.
	TakerFee int64

	// Total fees accrued in native quote
	FeesAccrued bin.Uint128
	// Total fees settled in native quote
	FeesToReferrers bin.Uint128

	// Referrer rebates to be distributed
	ReferrerRebatesAccrued uint64

	// Fees generated and available to withdraw via sweep_fees
	FeesAvailable uint64

	// Cumulative maker volume (same as taker volume) in quote native units
	MakerVolume bin.Uint128

	// Cumulative taker volume in quote native units due to place take orders
	TakerVolumeWoOo bin.Uint128

	BaseMint  solana.PublicKey
	QuoteMint solana.PublicKey

	MarketBaseVault  solana.PublicKey
	BaseDepositTotal uint64

	MarketQuoteVault  solana.PublicKey
	QuoteDepositTotal uint64

	Reserved [128]byte
}

func (m *Market) MaxBaseLots() int64 {
	return math.MaxInt64 / m.BaseLotSize
}

func (m *Market) MaxQuoteLots() int64 {
	return math.MaxInt64 / m.QuoteLotSize
}

func (m *Market) NativePriceToLot(price *big.Float) (int64, error) {
	baseLotSize := big.NewFloat(float64(m.BaseLotSize))
	quoteLotSize := big.NewFloat(float64(m.QuoteLotSize))

	result := price.Mul(price, baseLotSize)
	result.Quo(result, quoteLotSize)

	if !result.IsInt() {
		return 0, errors.New("invalid oracle price")
	}

	intResult, _ := result.Int64()
	return intResult, nil
}

func (m *Market) subtractTakerFees(quote int64) int64 {
	quoteBig := big.NewInt(quote)
	feeScaleFactor := big.NewInt(FEES_SCALE_FACTOR)
	takerFee := big.NewInt(int64(m.TakerFee))

	numerator := quoteBig.Mul(quoteBig, feeScaleFactor)
	denominator := feeScaleFactor.Add(feeScaleFactor, takerFee)

	result := quoteBig.Div(numerator, denominator)

	return result.Int64()
}

func (m *Market) MakerRebateFloor(amount uint64) uint64 {
	if m.MakerFee > 0 {
		return 0
	} else {
		return m.unsignedMakerFeesFloor(amount)
	}
}

func (m *Market) unsignedMakerFeesFloor(amount uint64) uint64 {
	amountBig := big.NewInt(int64(amount))
	makerFeeBig := big.NewInt(int64(math.Abs(float64(m.MakerFee))))
	feesScaleFactorBig := big.NewInt(FEES_SCALE_FACTOR)

	result := amountBig.Mul(amountBig, makerFeeBig)
	result.Div(result, feesScaleFactorBig)

	return result.Uint64()
}

func (m *Market) SubtractTakerFees(quote int64) int64 {
	return int64((float64(quote) * float64(FEES_SCALE_FACTOR)) / (float64(FEES_SCALE_FACTOR) + float64(m.TakerFee)))
}
