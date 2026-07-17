package model

import (
	"time"

	"auction/internal/modules/listing/domain/enum"
	"auction/internal/modules/listing/domain/errs"
)

// SkuModel is the Stock Keeping Unit: a concrete spec combination of an SPU
// (e.g. {"颜色":"红","尺寸":"L"}) with its own price, quantity and lifecycle.
type SkuModel struct {
	id           uint64
	spuID        uint64
	specValues   map[string]string
	priceInCents uint64
	quantity     uint64
	status       enum.ListingStatusEnum
	version      uint64
	createdAt    time.Time
	updatedAt    time.Time
}

func NewSkuModel(
	spuID uint64,
	specValues map[string]string,
	priceInCents, quantity uint64,
) (SkuModel, error) {
	if err := validateSku(spuID, specValues, priceInCents); err != nil {
		return SkuModel{}, err
	}

	status, err := enum.NewListingStatusEnum(enum.EnumListingStatusDraft)
	if err != nil {
		return SkuModel{}, err
	}

	now := time.Now().UTC()
	return SkuModel{
		spuID:        spuID,
		specValues:   specValues,
		priceInCents: priceInCents,
		quantity:     quantity,
		status:       status,
		version:      1,
		createdAt:    now,
		updatedAt:    now,
	}, nil
}

func RestoreSkuModel(
	id, spuID uint64,
	specValues map[string]string,
	priceInCents, quantity uint64,
	status enum.ListingStatusEnum,
	version uint64,
	createdAt, updatedAt time.Time,
) (SkuModel, error) {
	if id == 0 {
		return SkuModel{}, errs.ErrSkuIDRequired
	}

	if err := validateSku(spuID, specValues, priceInCents); err != nil {
		return SkuModel{}, err
	}

	return SkuModel{
		id:           id,
		spuID:        spuID,
		specValues:   specValues,
		priceInCents: priceInCents,
		quantity:     quantity,
		status:       status,
		version:      version,
		createdAt:    createdAt,
		updatedAt:    updatedAt,
	}, nil
}

func (s *SkuModel) ID() uint64 {
	return s.id
}

func (s *SkuModel) SpuID() uint64 {
	return s.spuID
}

func (s *SkuModel) SpecValues() map[string]string {
	return s.specValues
}

func (s *SkuModel) PriceInCents() uint64 {
	return s.priceInCents
}

func (s *SkuModel) Quantity() uint64 {
	return s.quantity
}

func (s *SkuModel) Status() enum.ListingStatusEnum {
	return s.status
}

func (s *SkuModel) Version() uint64 {
	return s.version
}

func (s *SkuModel) CreatedAt() time.Time {
	return s.createdAt
}

func (s *SkuModel) UpdatedAt() time.Time {
	return s.updatedAt
}

// IsAuctionable reports whether the SKU can back a new auction
func (s *SkuModel) IsAuctionable() bool {
	return s.status.IsPublished() && s.quantity > 0
}

// Publish makes the SKU sellable; allowed from draft and off_shelf.
// The cross-aggregate rule "parent SPU must be published" is enforced by the
// publish command inside the unit of work, not here.
func (s *SkuModel) Publish() error {
	if s.status.IsPublished() {
		return errs.ErrSkuAlreadyPublished
	}

	status, err := enum.NewListingStatusEnum(enum.EnumListingStatusPublished)
	if err != nil {
		return err
	}

	s.status = status
	s.touch()
	return nil
}

// OffShelf takes the SKU off shelf; allowed only from published
func (s *SkuModel) OffShelf() error {
	if !s.status.IsPublished() {
		return errs.ErrSkuNotPublished
	}

	status, err := enum.NewListingStatusEnum(enum.EnumListingStatusOffShelf)
	if err != nil {
		return err
	}

	s.status = status
	s.touch()
	return nil
}

// Update changes the SKU attributes; allowed only in draft and off_shelf
func (s *SkuModel) Update(specValues map[string]string, priceInCents, quantity uint64) error {
	if s.status.IsPublished() {
		return errs.ErrSkuNotEditable
	}

	if err := validateSku(s.spuID, specValues, priceInCents); err != nil {
		return err
	}

	s.specValues = specValues
	s.priceInCents = priceInCents
	s.quantity = quantity
	s.touch()
	return nil
}

func (s *SkuModel) touch() {
	s.version++
	s.updatedAt = time.Now().UTC()
}

func validateSku(spuID uint64, specValues map[string]string, priceInCents uint64) error {
	if spuID == 0 {
		return errs.ErrSpuIDRequired
	}

	if len(specValues) == 0 {
		return errs.ErrSkuSpecValuesRequired
	}

	if priceInCents == 0 {
		return errs.ErrSkuPriceInvalid
	}

	return nil
}
