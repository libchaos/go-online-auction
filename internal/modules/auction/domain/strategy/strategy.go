package strategy

import (
	"time"

	"auction/internal/modules/auction/domain/enum"
	"auction/internal/modules/auction/domain/errs"
)

// TradingStrategy encapsulates the core rules of one trading mode.
type TradingStrategy interface {
	Mode() enum.TradingModeEnum
	ValidateBid(auction AuctionView, amount MoneyView) error
	SuggestNextPrice(auction AuctionView) (MoneyView, error)
	DetermineWinner(auction AuctionView, bids []BidView) (Winner, error)
	ShouldCloseOnAccept() bool
}

// Resolver resolves the strategy for a given trading mode.
type Resolver interface {
	ForMode(mode enum.TradingModeEnum) (TradingStrategy, error)
}

// Strategies is a collection of trading strategies used for dependency injection.
type Strategies []TradingStrategy

type Winner struct {
	UserID    uint64
	PayAmount MoneyView
	BidID     *uint64
}

// AuctionView exposes the read-only auction state needed by strategies.
type AuctionView interface {
	State() enum.AuctionStateEnum
	TradingMode() enum.TradingModeEnum
	HighestBidAmount() *uint64
	CurrentPrice() *uint64
	StartingPrice() *uint64
	PriceStep() *uint64
	ReservePrice() *uint64
	AntiSnipeEnabled() bool
	ExtensionWindowSec() int64
	StartTime() *time.Time
	EndTime() time.Time
}

// BidView exposes the read-only bid state needed by strategies.
type BidView interface {
	ID() uint64
	UserID() uint64
	Amount() MoneyView
	MaxAmount() MoneyView
}

// MoneyView exposes the minimal monetary value needed by strategies.
type MoneyView interface {
	AmountInCents() uint64
}

// ProxyAction describes a public bid to persist during proxy resolution.
type ProxyAction struct {
	UserID    uint64
	Amount    MoneyView
	MaxAmount MoneyView
}

// ProxyResolvable is implemented by trading modes that support automatic
// proxy (eBay-style) bidding, where a bidder submits a maximum and the system
// derives the minimum public bid needed to keep them as the leader.
type ProxyResolvable interface {
	ResolveProxyBids(auction AuctionView, existing []BidView, newBidder uint64, newMax MoneyView) ([]ProxyAction, error)
}

type mapResolver struct {
	strategies map[string]TradingStrategy
}

func NewResolver(strategies []TradingStrategy) Resolver {
	resolver := &mapResolver{strategies: make(map[string]TradingStrategy, len(strategies))}
	for _, strategy := range strategies {
		mode := strategy.Mode()
		resolver.strategies[mode.String()] = strategy
	}

	return resolver
}

func (resolver *mapResolver) ForMode(mode enum.TradingModeEnum) (TradingStrategy, error) {
	if strategy, ok := resolver.strategies[mode.String()]; ok {
		return strategy, nil
	}

	return nil, errs.ErrUnsupportedTradingMode
}

// DefaultStrategies returns the built-in trading strategies for every mode.
// It is a pure constructor (no package-level mutable state) used as the source
// of the application-wide resolver and as a safe fallback for models restored
// outside the dependency-injection graph (e.g. in tests).
func DefaultStrategies() []TradingStrategy {
	return []TradingStrategy{
		NewEnglishAuctionStrategy(),
		NewDutchAuctionStrategy(),
		NewSealedBidAuctionStrategy(),
		NewVickreyAuctionStrategy(),
		NewFixedPriceAuctionStrategy(),
		NewEbayProxyAuctionStrategy(),
	}
}

// NewDefaultResolver builds the application-wide resolver from DefaultStrategies.
func NewDefaultResolver() Resolver {
	return NewResolver(DefaultStrategies())
}

func mode(value string) enum.TradingModeEnum {
	resolved, err := enum.NewTradingModeEnum(value)
	if err != nil {
		panic(err)
	}

	return resolved
}
