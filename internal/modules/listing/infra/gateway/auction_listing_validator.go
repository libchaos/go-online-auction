// Package gateway adapts the listing module to ports defined by other modules
// (anti-corruption layer for cross-module integration).
package gateway

import (
	"context"
	"errors"

	auctionports "auction/internal/modules/auction/ports"
	"auction/internal/modules/listing/domain/errs"
	"auction/internal/modules/listing/ports"
)

var _ auctionports.ListingValidator = (*AuctionListingValidator)(nil)

// AuctionListingValidator implements the auction module's ListingValidator
// port: an auction's listing_id refers to a SKU, which is auctionable when it
// is published with quantity available and its parent SPU is published.
type AuctionListingValidator struct {
	skuRepository ports.SkuRepository
	spuRepository ports.SpuRepository
}

func NewAuctionListingValidator(
	skuRepository ports.SkuRepository,
	spuRepository ports.SpuRepository,
) *AuctionListingValidator {
	return &AuctionListingValidator{
		skuRepository: skuRepository,
		spuRepository: spuRepository,
	}
}

func (v *AuctionListingValidator) IsAuctionable(ctx context.Context, listingID uint64) (bool, error) {
	sku, err := v.skuRepository.FindByID(ctx, listingID)
	if err != nil {
		if errors.Is(err, errs.ErrSkuNotFound) {
			return false, nil
		}
		return false, err
	}

	if !sku.IsAuctionable() {
		return false, nil
	}

	spu, err := v.spuRepository.FindByID(ctx, sku.SpuID())
	if err != nil {
		return false, err
	}

	status := spu.Status()
	return status.IsPublished(), nil
}
