package strategy_test

import (
	"testing"
	"time"

	"auction/internal/modules/auction/domain/enum"
	"auction/internal/modules/auction/domain/errs"
	"auction/internal/modules/auction/domain/strategy"
	"github.com/stretchr/testify/require"
)

// testMoney adapts a raw cents value to the MoneyView interface for tests.
type testMoney uint64

func (money testMoney) AmountInCents() uint64 {
	return uint64(money)
}

// testBid adapts fixed values to the BidView interface for tests.
type testBid struct {
	id        uint64
	userID    uint64
	amount    uint64
	maxAmount *uint64
}

func (bid testBid) ID() uint64 {
	return bid.id
}

func (bid testBid) UserID() uint64 {
	return bid.userID
}

func (bid testBid) Amount() strategy.MoneyView {
	return strategy.NewMoney(bid.amount)
}

func (bid testBid) MaxAmount() strategy.MoneyView {
	if bid.maxAmount == nil {
		return nil
	}

	return strategy.NewMoney(*bid.maxAmount)
}

// testAuction adapts fixed values to the AuctionView interface for tests.
type testAuction struct {
	state              enum.AuctionStateEnum
	tradingMode        enum.TradingModeEnum
	highestBidAmount   *uint64
	currentPrice       *uint64
	startingPrice      *uint64
	priceStep          *uint64
	reservePrice       *uint64
	antiSnipeEnabled   bool
	extensionWindowSec int64
	startTime          *time.Time
	endTime            time.Time
}

func (auction testAuction) State() enum.AuctionStateEnum {
	return auction.state
}

func (auction testAuction) TradingMode() enum.TradingModeEnum {
	return auction.tradingMode
}

func (auction testAuction) HighestBidAmount() *uint64 {
	return auction.highestBidAmount
}

func (auction testAuction) CurrentPrice() *uint64 {
	return auction.currentPrice
}

func (auction testAuction) StartingPrice() *uint64 {
	return auction.startingPrice
}

func (auction testAuction) PriceStep() *uint64 {
	return auction.priceStep
}

func (auction testAuction) ReservePrice() *uint64 {
	return auction.reservePrice
}

func (auction testAuction) AntiSnipeEnabled() bool {
	return auction.antiSnipeEnabled
}

func (auction testAuction) ExtensionWindowSec() int64 {
	return auction.extensionWindowSec
}

func (auction testAuction) StartTime() *time.Time {
	return auction.startTime
}

func (auction testAuction) EndTime() time.Time {
	return auction.endTime
}

func mustMode(value string) enum.TradingModeEnum {
	resolved, err := enum.NewTradingModeEnum(value)
	if err != nil {
		panic(err)
	}

	return resolved
}

func ptr(value uint64) *uint64 {
	return &value
}

func TestEnglishAuctionStrategy(t *testing.T) {
	sut := strategy.NewEnglishAuctionStrategy()

	t.Run("Mode returns english", func(t *testing.T) {
		// Arrange / Act
		mode := sut.Mode()

		// Assert
		require.Equal(t, enum.EnumTradingModeEnglish, mode.String())
	})

	t.Run("ValidateBid rejects zero first bid", func(t *testing.T) {
		// Arrange
		auction := testAuction{}

		// Act
		err := sut.ValidateBid(auction, strategy.NewMoney(0))

		// Assert
		require.ErrorIs(t, err, errs.ErrFirstBidMustBePositive)
	})

	t.Run("ValidateBid accepts positive first bid", func(t *testing.T) {
		// Arrange
		auction := testAuction{}

		// Act
		err := sut.ValidateBid(auction, strategy.NewMoney(1000))

		// Assert
		require.NoError(t, err)
	})

	t.Run("ValidateBid rejects bid that does not exceed highest", func(t *testing.T) {
		// Arrange
		auction := testAuction{highestBidAmount: ptr(2000)}

		// Act
		err := sut.ValidateBid(auction, strategy.NewMoney(2000))

		// Assert
		require.ErrorIs(t, err, errs.ErrBidMustExceedHighest)
	})

	t.Run("ValidateBid accepts bid above highest", func(t *testing.T) {
		// Arrange
		auction := testAuction{highestBidAmount: ptr(2000)}

		// Act
		err := sut.ValidateBid(auction, strategy.NewMoney(3000))

		// Assert
		require.NoError(t, err)
	})

	t.Run("SuggestNextPrice returns starting price when no bids", func(t *testing.T) {
		// Arrange
		auction := testAuction{startingPrice: ptr(500)}

		// Act
		next, err := sut.SuggestNextPrice(auction)

		// Assert
		require.NoError(t, err)
		require.Equal(t, uint64(500), next.AmountInCents())
	})

	t.Run("SuggestNextPrice increments highest by step", func(t *testing.T) {
		// Arrange
		auction := testAuction{highestBidAmount: ptr(2000), priceStep: ptr(100)}

		// Act
		next, err := sut.SuggestNextPrice(auction)

		// Assert
		require.NoError(t, err)
		require.Equal(t, uint64(2100), next.AmountInCents())
	})

	t.Run("DetermineWinner returns no winner without bids", func(t *testing.T) {
		// Arrange
		auction := testAuction{}

		// Act
		winner, err := sut.DetermineWinner(auction, nil)

		// Assert
		require.NoError(t, err)
		require.Equal(t, uint64(0), winner.UserID)
		require.Nil(t, winner.BidID)
	})

	t.Run("DetermineWinner selects highest bid", func(t *testing.T) {
		// Arrange
		auction := testAuction{}
		bids := []strategy.BidView{
			testBid{id: 1, userID: 10, amount: 1000},
			testBid{id: 2, userID: 20, amount: 2500},
		}

		// Act
		winner, err := sut.DetermineWinner(auction, bids)

		// Assert
		require.NoError(t, err)
		require.Equal(t, uint64(20), winner.UserID)
		require.Equal(t, uint64(2500), winner.PayAmount.AmountInCents())
		require.NotNil(t, winner.BidID)
		require.Equal(t, uint64(2), *winner.BidID)
	})

	t.Run("DetermineWinner rejects when reserve not met", func(t *testing.T) {
		// Arrange
		auction := testAuction{reservePrice: ptr(5000)}
		bids := []strategy.BidView{testBid{id: 1, userID: 10, amount: 1000}}

		// Act
		_, err := sut.DetermineWinner(auction, bids)

		// Assert
		require.ErrorIs(t, err, errs.ErrReserveNotMet)
	})

	t.Run("ShouldCloseOnAccept is false", func(t *testing.T) {
		// Arrange / Act / Assert
		require.False(t, sut.ShouldCloseOnAccept())
	})
}

func TestDutchAuctionStrategy(t *testing.T) {
	sut := strategy.NewDutchAuctionStrategy()

	t.Run("Mode returns dutch", func(t *testing.T) {
		// Arrange / Act / Assert
		mode := sut.Mode()
		require.Equal(t, enum.EnumTradingModeDutch, mode.String())
	})

	t.Run("ValidateBid errors when price unavailable", func(t *testing.T) {
		// Arrange
		auction := testAuction{}

		// Act
		err := sut.ValidateBid(auction, strategy.NewMoney(100))

		// Assert
		require.ErrorIs(t, err, errs.ErrDutchPriceNotAvailable)
	})

	t.Run("ValidateBid requires exact price match", func(t *testing.T) {
		// Arrange
		auction := testAuction{currentPrice: ptr(1000)}

		// Act
		err := sut.ValidateBid(auction, strategy.NewMoney(900))

		// Assert
		require.ErrorIs(t, err, errs.ErrDutchBidMustMatchPrice)
	})

	t.Run("ValidateBid accepts matching price", func(t *testing.T) {
		// Arrange
		auction := testAuction{currentPrice: ptr(1000)}

		// Act
		err := sut.ValidateBid(auction, strategy.NewMoney(1000))

		// Assert
		require.NoError(t, err)
	})

	t.Run("SuggestNextPrice errors when unavailable", func(t *testing.T) {
		// Arrange
		auction := testAuction{}

		// Act
		_, err := sut.SuggestNextPrice(auction)

		// Assert
		require.ErrorIs(t, err, errs.ErrDutchPriceNotAvailable)
	})

	t.Run("DetermineWinner selects first bid", func(t *testing.T) {
		// Arrange
		auction := testAuction{}
		bids := []strategy.BidView{testBid{id: 7, userID: 42, amount: 800}}

		// Act
		winner, err := sut.DetermineWinner(auction, bids)

		// Assert
		require.NoError(t, err)
		require.Equal(t, uint64(42), winner.UserID)
		require.Equal(t, uint64(800), winner.PayAmount.AmountInCents())
	})

	t.Run("ShouldCloseOnAccept is true", func(t *testing.T) {
		// Arrange / Act / Assert
		require.True(t, sut.ShouldCloseOnAccept())
	})
}

func TestSealedBidAuctionStrategy(t *testing.T) {
	sut := strategy.NewSealedBidAuctionStrategy()

	t.Run("Mode returns sealed_bid", func(t *testing.T) {
		// Arrange / Act / Assert
		mode := sut.Mode()
		require.Equal(t, enum.EnumTradingModeSealedBid, mode.String())
	})

	t.Run("ValidateBid rejects zero", func(t *testing.T) {
		// Arrange
		auction := testAuction{}

		// Act
		err := sut.ValidateBid(auction, strategy.NewMoney(0))

		// Assert
		require.ErrorIs(t, err, errs.ErrFirstBidMustBePositive)
	})

	t.Run("ValidateBid accepts positive", func(t *testing.T) {
		// Arrange
		auction := testAuction{}

		// Act
		err := sut.ValidateBid(auction, strategy.NewMoney(500))

		// Assert
		require.NoError(t, err)
	})

	t.Run("DetermineWinner selects highest bid", func(t *testing.T) {
		// Arrange
		auction := testAuction{}
		bids := []strategy.BidView{
			testBid{id: 1, userID: 10, amount: 1000},
			testBid{id: 2, userID: 20, amount: 4000},
		}

		// Act
		winner, err := sut.DetermineWinner(auction, bids)

		// Assert
		require.NoError(t, err)
		require.Equal(t, uint64(20), winner.UserID)
		require.Equal(t, uint64(4000), winner.PayAmount.AmountInCents())
	})

	t.Run("ShouldCloseOnAccept is false", func(t *testing.T) {
		// Arrange / Act / Assert
		require.False(t, sut.ShouldCloseOnAccept())
	})
}

func TestVickreyAuctionStrategy(t *testing.T) {
	sut := strategy.NewVickreyAuctionStrategy()

	t.Run("Mode returns vickrey", func(t *testing.T) {
		// Arrange / Act / Assert
		mode := sut.Mode()
		require.Equal(t, enum.EnumTradingModeVickrey, mode.String())
	})

	t.Run("ValidateBid rejects zero", func(t *testing.T) {
		// Arrange
		auction := testAuction{}

		// Act
		err := sut.ValidateBid(auction, strategy.NewMoney(0))

		// Assert
		require.ErrorIs(t, err, errs.ErrFirstBidMustBePositive)
	})

	t.Run("DetermineWinner pays second highest price", func(t *testing.T) {
		// Arrange
		auction := testAuction{}
		bids := []strategy.BidView{
			testBid{id: 1, userID: 10, amount: 1000},
			testBid{id: 2, userID: 20, amount: 3000},
		}

		// Act
		winner, err := sut.DetermineWinner(auction, bids)

		// Assert
		require.NoError(t, err)
		require.Equal(t, uint64(20), winner.UserID)
		require.Equal(t, uint64(1000), winner.PayAmount.AmountInCents())
	})

	t.Run("DetermineWinner pays own bid with single bidder", func(t *testing.T) {
		// Arrange
		auction := testAuction{}
		bids := []strategy.BidView{testBid{id: 1, userID: 10, amount: 2500}}

		// Act
		winner, err := sut.DetermineWinner(auction, bids)

		// Assert
		require.NoError(t, err)
		require.Equal(t, uint64(10), winner.UserID)
		require.Equal(t, uint64(2500), winner.PayAmount.AmountInCents())
	})

	t.Run("DetermineWinner rejects when reserve not met", func(t *testing.T) {
		// Arrange
		auction := testAuction{reservePrice: ptr(5000)}
		bids := []strategy.BidView{testBid{id: 1, userID: 10, amount: 3000}}

		// Act
		_, err := sut.DetermineWinner(auction, bids)

		// Assert
		require.ErrorIs(t, err, errs.ErrReserveNotMet)
	})

	t.Run("ShouldCloseOnAccept is false", func(t *testing.T) {
		// Arrange / Act / Assert
		require.False(t, sut.ShouldCloseOnAccept())
	})
}

func TestFixedPriceAuctionStrategy(t *testing.T) {
	sut := strategy.NewFixedPriceAuctionStrategy()

	t.Run("Mode returns fixed_price", func(t *testing.T) {
		// Arrange / Act / Assert
		mode := sut.Mode()
		require.Equal(t, enum.EnumTradingModeFixedPrice, mode.String())
	})

	t.Run("ValidateBid errors when not configured", func(t *testing.T) {
		// Arrange
		auction := testAuction{}

		// Act
		err := sut.ValidateBid(auction, strategy.NewMoney(100))

		// Assert
		require.ErrorIs(t, err, errs.ErrFixedPriceNotConfigured)
	})

	t.Run("ValidateBid requires exact price", func(t *testing.T) {
		// Arrange
		auction := testAuction{startingPrice: ptr(1000)}

		// Act
		err := sut.ValidateBid(auction, strategy.NewMoney(900))

		// Assert
		require.ErrorIs(t, err, errs.ErrFixedPriceMismatch)
	})

	t.Run("ValidateBid accepts exact price", func(t *testing.T) {
		// Arrange
		auction := testAuction{startingPrice: ptr(1000)}

		// Act
		err := sut.ValidateBid(auction, strategy.NewMoney(1000))

		// Assert
		require.NoError(t, err)
	})

	t.Run("DetermineWinner selects first bid", func(t *testing.T) {
		// Arrange
		auction := testAuction{}
		bids := []strategy.BidView{testBid{id: 3, userID: 99, amount: 1000}}

		// Act
		winner, err := sut.DetermineWinner(auction, bids)

		// Assert
		require.NoError(t, err)
		require.Equal(t, uint64(99), winner.UserID)
		require.Equal(t, uint64(1000), winner.PayAmount.AmountInCents())
	})

	t.Run("ShouldCloseOnAccept is true", func(t *testing.T) {
		// Arrange / Act / Assert
		require.True(t, sut.ShouldCloseOnAccept())
	})
}

func TestEbayProxyAuctionStrategy(t *testing.T) {
	sut := strategy.NewEbayProxyAuctionStrategy()

	t.Run("Mode returns ebay_proxy", func(t *testing.T) {
		// Arrange / Act / Assert
		mode := sut.Mode()
		require.Equal(t, enum.EnumTradingModeEbayProxy, mode.String())
	})

	t.Run("ValidateBid rejects bid at or below public price", func(t *testing.T) {
		// Arrange
		auction := testAuction{highestBidAmount: ptr(1000)}

		// Act
		err := sut.ValidateBid(auction, strategy.NewMoney(1000))

		// Assert
		require.ErrorIs(t, err, errs.ErrProxyMaxTooLow)
	})

	t.Run("ValidateBid accepts bid above public price", func(t *testing.T) {
		// Arrange
		auction := testAuction{highestBidAmount: ptr(1000)}

		// Act
		err := sut.ValidateBid(auction, strategy.NewMoney(1500))

		// Assert
		require.NoError(t, err)
	})

	t.Run("DetermineWinner pays second highest plus step", func(t *testing.T) {
		// Arrange
		auction := testAuction{priceStep: ptr(10)}
		bids := []strategy.BidView{
			testBid{id: 1, userID: 10, amount: 100, maxAmount: ptr(100)},
			testBid{id: 2, userID: 20, amount: 80, maxAmount: ptr(80)},
		}

		// Act
		winner, err := sut.DetermineWinner(auction, bids)

		// Assert
		require.NoError(t, err)
		require.Equal(t, uint64(10), winner.UserID)
		require.Equal(t, uint64(90), winner.PayAmount.AmountInCents())
	})

	t.Run("DetermineWinner pays own max with single bidder", func(t *testing.T) {
		// Arrange
		auction := testAuction{priceStep: ptr(10)}
		bids := []strategy.BidView{testBid{id: 1, userID: 10, amount: 100, maxAmount: ptr(100)}}

		// Act
		winner, err := sut.DetermineWinner(auction, bids)

		// Assert
		require.NoError(t, err)
		require.Equal(t, uint64(10), winner.UserID)
		require.Equal(t, uint64(100), winner.PayAmount.AmountInCents())
	})

	t.Run("ResolveProxyBids derives public price above current leader", func(t *testing.T) {
		// Arrange
		auction := testAuction{
			startingPrice:    ptr(100),
			priceStep:        ptr(10),
			highestBidAmount: ptr(0),
		}
		existing := []strategy.BidView{
			testBid{id: 1, userID: 1, amount: 120, maxAmount: ptr(120)},
		}
		newBidder := uint64(2)
		newMax := strategy.NewMoney(200)

		// Act
		actions, err := sut.ResolveProxyBids(auction, existing, newBidder, newMax)

		// Assert
		require.NoError(t, err)
		require.Len(t, actions, 1)
		require.Equal(t, uint64(2), actions[0].UserID)
		require.Equal(t, uint64(130), actions[0].Amount.AmountInCents())
		require.Equal(t, uint64(200), actions[0].MaxAmount.AmountInCents())
	})

	t.Run("ResolveProxyBids returns no action when max too low", func(t *testing.T) {
		// Arrange
		auction := testAuction{
			startingPrice:    ptr(100),
			priceStep:        ptr(10),
			highestBidAmount: ptr(150),
		}
		existing := []strategy.BidView{
			testBid{id: 1, userID: 1, amount: 150, maxAmount: ptr(150)},
		}

		// Act
		actions, err := sut.ResolveProxyBids(auction, existing, 2, strategy.NewMoney(50))

		// Assert
		require.NoError(t, err)
		require.Empty(t, actions)
	})

	t.Run("ShouldCloseOnAccept is false", func(t *testing.T) {
		// Arrange / Act / Assert
		require.False(t, sut.ShouldCloseOnAccept())
	})
}

func TestStrategyResolver(t *testing.T) {
	resolver := strategy.NewResolver([]strategy.TradingStrategy{
		strategy.NewEnglishAuctionStrategy(),
		strategy.NewDutchAuctionStrategy(),
		strategy.NewSealedBidAuctionStrategy(),
		strategy.NewVickreyAuctionStrategy(),
		strategy.NewFixedPriceAuctionStrategy(),
		strategy.NewEbayProxyAuctionStrategy(),
	})

	t.Run("ForMode returns matching strategy", func(t *testing.T) {
		// Arrange / Act
		selected, err := resolver.ForMode(mustMode(enum.EnumTradingModeVickrey))

		// Assert
		require.NoError(t, err)
		mode := selected.Mode()
		require.Equal(t, enum.EnumTradingModeVickrey, mode.String())
	})

	t.Run("ForMode errors for unregistered mode", func(t *testing.T) {
		// Arrange
		partial := strategy.NewResolver([]strategy.TradingStrategy{
			strategy.NewEnglishAuctionStrategy(),
		})

		// Act
		_, err := partial.ForMode(mustMode(enum.EnumTradingModeDutch))

		// Assert
		require.ErrorIs(t, err, errs.ErrUnsupportedTradingMode)
	})
}

func TestStrategyMoney(t *testing.T) {
	t.Run("Add sums cents", func(t *testing.T) {
		// Arrange / Act
		result := strategy.NewMoney(100).Add(strategy.NewMoney(50))

		// Assert
		require.Equal(t, uint64(150), result.AmountInCents())
	})

	t.Run("IncrementBy adds step", func(t *testing.T) {
		// Arrange / Act
		result := strategy.NewMoney(100).IncrementBy(25)

		// Assert
		require.Equal(t, uint64(125), result.AmountInCents())
	})

	t.Run("IsGreaterThan compares correctly", func(t *testing.T) {
		// Arrange / Act / Assert
		require.True(t, strategy.NewMoney(200).IsGreaterThan(strategy.NewMoney(100)))
		require.False(t, strategy.NewMoney(100).IsGreaterThan(strategy.NewMoney(200)))
	})

	t.Run("IsGreaterThanOrEqual compares correctly", func(t *testing.T) {
		// Arrange / Act / Assert
		require.True(t, strategy.NewMoney(100).IsGreaterThanOrEqual(strategy.NewMoney(100)))
		require.False(t, strategy.NewMoney(99).IsGreaterThanOrEqual(strategy.NewMoney(100)))
	})
}
