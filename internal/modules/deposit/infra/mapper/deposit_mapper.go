package mapper

import (
	"auction/internal/modules/deposit/domain/model"
	"auction/internal/modules/deposit/infra/sqlcgen"
)

type DepositMapper struct{}

func NewDepositMapper() *DepositMapper {
	return &DepositMapper{}
}

func (mapper *DepositMapper) ToDomain(deposit sqlcgen.Deposit) (model.DepositModel, error) {
	return model.RestoreDepositModel(
		uint64(deposit.ID),
		uint64(deposit.UserID),
		uint64(deposit.AuctionID),
		uint64(deposit.AmountInCents),
		deposit.Currency,
		deposit.Status,
		toString(deposit.ExternalReference),
		toString(deposit.Reference),
		uint64(deposit.Version),
		deposit.CreatedAt,
		deposit.UpdatedAt,
	)
}

func (mapper *DepositMapper) ToCreateParams(deposit model.DepositModel) sqlcgen.CreateDepositParams {
	status := deposit.Status()

	return sqlcgen.CreateDepositParams{
		UserID:            int64(deposit.UserID()),
		AuctionID:         int64(deposit.AuctionID()),
		AmountInCents:     int64(deposit.Amount().AmountInCents()),
		Currency:          deposit.Currency(),
		Status:            status.String(),
		ExternalReference: toNullableString(deposit.ExternalReference()),
		Reference:         toNullableString(deposit.Reference()),
		Version:           int64(deposit.Version()),
		CreatedAt:         deposit.CreatedAt(),
		UpdatedAt:         deposit.UpdatedAt(),
	}
}

func (mapper *DepositMapper) ToUpdateParams(deposit model.DepositModel) sqlcgen.UpdateDepositParams {
	status := deposit.Status()

	return sqlcgen.UpdateDepositParams{
		ID:                int64(deposit.ID()),
		UserID:            int64(deposit.UserID()),
		AuctionID:         int64(deposit.AuctionID()),
		AmountInCents:     int64(deposit.Amount().AmountInCents()),
		Currency:          deposit.Currency(),
		Status:            status.String(),
		ExternalReference: toNullableString(deposit.ExternalReference()),
		Reference:         toNullableString(deposit.Reference()),
		Version:           int64(deposit.Version()),
		UpdatedAt:         deposit.UpdatedAt(),
	}
}

func toString(value *string) string {
	if value == nil {
		return ""
	}

	return *value
}

func toNullableString(value string) *string {
	if value == "" {
		return nil
	}

	return &value
}
