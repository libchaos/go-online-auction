package mapper

import (
	"auction/internal/modules/auction/domain/enum"
	"auction/internal/modules/auction/domain/model"
	"auction/internal/modules/auction/domain/strategy"
	"auction/internal/modules/auction/infra/sqlcgen"
)

type AuctionMapper struct {
	resolver strategy.Resolver
}

func NewAuctionMapper(resolver strategy.Resolver) *AuctionMapper {
	return &AuctionMapper{resolver: resolver}
}

func (m *AuctionMapper) ToDomain(a sqlcgen.Auction) (model.AuctionModel, error) {
	state, err := enum.NewAuctionStateEnum(string(a.State))
	if err != nil {
		return model.AuctionModel{}, err
	}

	tradingMode, err := enum.NewTradingModeEnum(a.TradingMode)
	if err != nil {
		return model.AuctionModel{}, err
	}

	return model.RestoreAuctionModelWithMode(
		uint64(a.ID),
		uint64(a.ListingID),
		a.StartTime,
		a.EndTime,
		state,
		tradingMode,
		toNullableUint64(a.HighestBidAmountInCents),
		toNullableUint64(a.StartingPrice),
		toNullableUint64(a.PriceStep),
		toNullableUint64(a.ReservePrice),
		toNullableUint64(a.CurrentPrice),
		toNullableUint64(a.WinnerUserID),
		toNullableUint64(a.WinningBidID),
		toNullableUint64(a.WinningBidAmount),
		a.AntiSnipeEnabled,
		a.ExtensionWindowSec,
		uint64(a.Version),
		a.CreatedAt,
		a.UpdatedAt,
		m.resolver,
	)
}

func (m *AuctionMapper) ToCreateParams(auction model.AuctionModel) sqlcgen.CreateAuctionParams {
	state := auction.State()
	tradingMode := auction.TradingMode()

	return sqlcgen.CreateAuctionParams{
		ListingID:               int64(auction.ListingID()),
		EndTime:                 auction.EndTime(),
		State:                   sqlcgen.AuctionState(state.String()),
		TradingMode:             tradingMode.String(),
		StartingPrice:           toNullableInt64(auction.StartingPrice()),
		PriceStep:               toNullableInt64(auction.PriceStep()),
		ReservePrice:            toNullableInt64(auction.ReservePrice()),
		CurrentPrice:            toNullableInt64(auction.CurrentPrice()),
		HighestBidAmountInCents: toNullableInt64(auction.HighestBidAmount()),
		WinnerUserID:            toNullableInt64(auction.WinnerUserID()),
		WinningBidID:            toNullableInt64(auction.WinningBidID()),
		WinningBidAmount:        toNullableInt64(auction.WinningBidAmount()),
		AntiSnipeEnabled:        auction.AntiSnipeEnabled(),
		ExtensionWindowSec:      auction.ExtensionWindowSec(),
		Version:                 int64(auction.Version()),
		CreatedAt:               auction.CreatedAt(),
		UpdatedAt:               auction.UpdatedAt(),
	}
}

func (m *AuctionMapper) ToUpdateParams(auction model.AuctionModel) sqlcgen.UpdateAuctionParams {
	state := auction.State()
	tradingMode := auction.TradingMode()

	return sqlcgen.UpdateAuctionParams{
		ListingID:               int64(auction.ListingID()),
		StartTime:               auction.StartTime(),
		EndTime:                 auction.EndTime(),
		State:                   sqlcgen.AuctionState(state.String()),
		TradingMode:             tradingMode.String(),
		StartingPrice:           toNullableInt64(auction.StartingPrice()),
		PriceStep:               toNullableInt64(auction.PriceStep()),
		ReservePrice:            toNullableInt64(auction.ReservePrice()),
		CurrentPrice:            toNullableInt64(auction.CurrentPrice()),
		HighestBidAmountInCents: toNullableInt64(auction.HighestBidAmount()),
		WinnerUserID:            toNullableInt64(auction.WinnerUserID()),
		WinningBidID:            toNullableInt64(auction.WinningBidID()),
		WinningBidAmount:        toNullableInt64(auction.WinningBidAmount()),
		AntiSnipeEnabled:        auction.AntiSnipeEnabled(),
		ExtensionWindowSec:      auction.ExtensionWindowSec(),
		Version:                 int64(auction.Version()),
		UpdatedAt:               auction.UpdatedAt(),
		ID:                      int64(auction.ID()),
	}
}

func toNullableUint64(v *int64) *uint64 {
	if v == nil {
		return nil
	}
	u := uint64(*v)
	return &u
}

func toNullableInt64(v *uint64) *int64 {
	if v == nil {
		return nil
	}
	i := int64(*v)
	return &i
}
