package mapper

import (
	"encoding/json"

	"auction/internal/modules/listing/domain/enum"
	"auction/internal/modules/listing/domain/model"
	"auction/internal/modules/listing/infra/sqlcgen"
)

type SpuMapper struct{}

func NewSpuMapper() *SpuMapper {
	return &SpuMapper{}
}

func (m *SpuMapper) ToDomain(s sqlcgen.Spu) (model.SpuModel, error) {
	status, err := enum.NewListingStatusEnum(string(s.Status))
	if err != nil {
		return model.SpuModel{}, err
	}

	var images []string
	if len(s.Images) > 0 {
		if err = json.Unmarshal(s.Images, &images); err != nil {
			return model.SpuModel{}, err
		}
	}

	return model.RestoreSpuModel(
		uint64(s.ID),
		s.Title,
		s.Description,
		uint64(s.CategoryID),
		s.Brand,
		images,
		status,
		uint64(s.Version),
		s.CreatedAt,
		s.UpdatedAt,
	)
}

func (m *SpuMapper) ToCreateParams(spu model.SpuModel) (sqlcgen.CreateSpuParams, error) {
	images, err := marshalImages(spu.Images())
	if err != nil {
		return sqlcgen.CreateSpuParams{}, err
	}

	status := spu.Status()
	return sqlcgen.CreateSpuParams{
		Title:       spu.Title(),
		Description: spu.Description(),
		CategoryID:  int64(spu.CategoryID()),
		Brand:       spu.Brand(),
		Images:      images,
		Status:      sqlcgen.ListingStatus(status.String()),
		Version:     int64(spu.Version()),
		CreatedAt:   spu.CreatedAt(),
		UpdatedAt:   spu.UpdatedAt(),
	}, nil
}

func (m *SpuMapper) ToUpdateParams(spu model.SpuModel) (sqlcgen.UpdateSpuParams, error) {
	images, err := marshalImages(spu.Images())
	if err != nil {
		return sqlcgen.UpdateSpuParams{}, err
	}

	status := spu.Status()
	return sqlcgen.UpdateSpuParams{
		Title:       spu.Title(),
		Description: spu.Description(),
		CategoryID:  int64(spu.CategoryID()),
		Brand:       spu.Brand(),
		Images:      images,
		Status:      sqlcgen.ListingStatus(status.String()),
		Version:     int64(spu.Version()),
		UpdatedAt:   spu.UpdatedAt(),
		ID:          int64(spu.ID()),
	}, nil
}

// marshalImages serializes the image URL list to JSONB, defaulting nil to []
// so the column never stores SQL NULL.
func marshalImages(images []string) ([]byte, error) {
	if images == nil {
		images = []string{}
	}
	return json.Marshal(images)
}
