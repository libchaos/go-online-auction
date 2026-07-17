package mapper

import (
	"encoding/json"

	"auction/internal/modules/listing/domain/enum"
	"auction/internal/modules/listing/domain/model"
	"auction/internal/modules/listing/infra/sqlcgen"
)

type SkuMapper struct{}

func NewSkuMapper() *SkuMapper {
	return &SkuMapper{}
}

func (m *SkuMapper) ToDomain(s sqlcgen.Sku) (model.SkuModel, error) {
	status, err := enum.NewListingStatusEnum(string(s.Status))
	if err != nil {
		return model.SkuModel{}, err
	}

	var specValues map[string]string
	if len(s.SpecValues) > 0 {
		if err = json.Unmarshal(s.SpecValues, &specValues); err != nil {
			return model.SkuModel{}, err
		}
	}

	return model.RestoreSkuModel(
		uint64(s.ID),
		uint64(s.SpuID),
		specValues,
		uint64(s.PriceInCents),
		uint64(s.Quantity),
		status,
		uint64(s.Version),
		s.CreatedAt,
		s.UpdatedAt,
	)
}

func (m *SkuMapper) ToCreateParams(sku model.SkuModel) (sqlcgen.CreateSkuParams, error) {
	specValues, err := json.Marshal(sku.SpecValues())
	if err != nil {
		return sqlcgen.CreateSkuParams{}, err
	}

	status := sku.Status()
	return sqlcgen.CreateSkuParams{
		SpuID:        int64(sku.SpuID()),
		SpecValues:   specValues,
		PriceInCents: int64(sku.PriceInCents()),
		Quantity:     int64(sku.Quantity()),
		Status:       sqlcgen.ListingStatus(status.String()),
		Version:      int64(sku.Version()),
		CreatedAt:    sku.CreatedAt(),
		UpdatedAt:    sku.UpdatedAt(),
	}, nil
}

func (m *SkuMapper) ToUpdateParams(sku model.SkuModel) (sqlcgen.UpdateSkuParams, error) {
	specValues, err := json.Marshal(sku.SpecValues())
	if err != nil {
		return sqlcgen.UpdateSkuParams{}, err
	}

	status := sku.Status()
	return sqlcgen.UpdateSkuParams{
		SpecValues:   specValues,
		PriceInCents: int64(sku.PriceInCents()),
		Quantity:     int64(sku.Quantity()),
		Status:       sqlcgen.ListingStatus(status.String()),
		Version:      int64(sku.Version()),
		UpdatedAt:    sku.UpdatedAt(),
		ID:           int64(sku.ID()),
	}, nil
}
