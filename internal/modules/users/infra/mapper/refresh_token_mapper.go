package mapper

import (
	"auction/internal/modules/users/domain/model"
	"auction/internal/modules/users/infra/sqlcgen"
)

type RefreshTokenMapper struct{}

func NewRefreshTokenMapper() *RefreshTokenMapper {
	return &RefreshTokenMapper{}
}

func (m *RefreshTokenMapper) ToDomain(t sqlcgen.RefreshToken) (model.RefreshTokenModel, error) {
	var replacedBy *uint64
	if t.ReplacedBy != nil {
		v := uint64(*t.ReplacedBy)
		replacedBy = &v
	}

	return model.RestoreRefreshTokenModel(
		uint64(t.ID),
		uint64(t.UserID),
		t.TokenHash,
		t.ExpiresAt,
		t.RevokedAt,
		replacedBy,
		t.CreatedAt,
	)
}

func (m *RefreshTokenMapper) ToCreateParams(token model.RefreshTokenModel) sqlcgen.CreateRefreshTokenParams {
	return sqlcgen.CreateRefreshTokenParams{
		UserID:     int64(token.UserID()),
		TokenHash:  token.TokenHash(),
		ExpiresAt:  token.ExpiresAt(),
		RevokedAt:  token.RevokedAt(),
		ReplacedBy: toNullableInt64(token.ReplacedBy()),
		CreatedAt:  token.CreatedAt(),
	}
}

func (m *RefreshTokenMapper) ToUpdateParams(token model.RefreshTokenModel) sqlcgen.UpdateRefreshTokenParams {
	return sqlcgen.UpdateRefreshTokenParams{
		RevokedAt:  token.RevokedAt(),
		ReplacedBy: toNullableInt64(token.ReplacedBy()),
		ID:         int64(token.ID()),
	}
}

func toNullableInt64(v *uint64) *int64 {
	if v == nil {
		return nil
	}
	i := int64(*v)
	return &i
}
