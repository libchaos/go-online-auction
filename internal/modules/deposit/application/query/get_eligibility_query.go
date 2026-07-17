package query

import (
	"context"
	"errors"

	"auction/internal/modules/deposit/domain/errs"
	"auction/internal/modules/deposit/ports"
)

type GetEligibilityQueryInput struct {
	UserID    uint64
	AuctionID uint64
}

type GetEligibilityQueryOutput struct {
	Eligible bool
}

type GetEligibilityQuery struct {
	guard ports.DepositGuard
}

func NewGetEligibilityQuery(guard ports.DepositGuard) *GetEligibilityQuery {
	return &GetEligibilityQuery{guard: guard}
}

func (depositQuery *GetEligibilityQuery) Execute(
	ctx context.Context,
	input GetEligibilityQueryInput,
) (GetEligibilityQueryOutput, error) {
	err := depositQuery.guard.EnsureEligible(ctx, input.UserID, input.AuctionID)
	if err != nil {
		if errors.Is(err, errs.ErrDepositRequired) || errors.Is(err, errs.ErrDepositInsufficient) ||
			errors.Is(err, errs.ErrDepositNotHeld) {
			return GetEligibilityQueryOutput{Eligible: false}, nil
		}

		return GetEligibilityQueryOutput{}, err
	}

	return GetEligibilityQueryOutput{Eligible: true}, nil
}
