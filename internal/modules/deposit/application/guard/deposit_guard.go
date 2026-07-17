package guard

import (
	"context"
	"errors"

	"auction/internal/modules/deposit/domain/errs"
	"auction/internal/modules/deposit/ports"
)

type DepositGuard struct {
	repository    ports.DepositRepository
	auctionConfig ports.AuctionConfigPort
}

func NewDepositGuard(
	repository ports.DepositRepository,
	auctionConfig ports.AuctionConfigPort,
) *DepositGuard {
	return &DepositGuard{
		repository:    repository,
		auctionConfig: auctionConfig,
	}
}

func (depositGuard *DepositGuard) EnsureEligible(
	ctx context.Context,
	userID uint64,
	auctionID uint64,
) error {
	config, err := depositGuard.auctionConfig.GetRequiredDeposit(ctx, auctionID)
	if err != nil {
		return err
	}

	if !config.Required {
		return nil
	}

	deposit, findErr := depositGuard.repository.FindByUserAndAuction(ctx, userID, auctionID)
	if findErr != nil {
		if errors.Is(findErr, errs.ErrDepositNotFound) {
			return errs.ErrDepositNotHeld
		}

		return findErr
	}

	status := deposit.Status()
	if status.String() != "held" {
		return errs.ErrDepositNotHeld
	}

	if !deposit.Amount().IsGreaterThanOrEqual(config.Amount) {
		return errs.ErrDepositInsufficient
	}

	return nil
}
