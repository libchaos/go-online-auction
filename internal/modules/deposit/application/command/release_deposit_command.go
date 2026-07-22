package command

import (
	"context"

	domainevent "auction/internal/modules/deposit/domain/event"
	"auction/internal/modules/deposit/infra/event/envelope"
	"auction/internal/modules/deposit/ports"
	ledgerports "auction/internal/modules/ledger/ports"
	"auction/internal/shared/modules/logger"
)

type ReleaseDepositCommandInput struct {
	DepositID uint64
}

type ReleaseDepositCommandOutput struct {
	DepositID uint64
	Status    string
}

const depositReleasedStatus = "released"

type ReleaseDepositCommand struct {
	uowFactory ports.DepositUnitOfWorkFactory
	logger     logger.Logger
}

func NewReleaseDepositCommand(
	uowFactory ports.DepositUnitOfWorkFactory,
	logger logger.Logger,
) *ReleaseDepositCommand {
	return &ReleaseDepositCommand{
		uowFactory: uowFactory,
		logger:     logger,
	}
}

func (command *ReleaseDepositCommand) Execute(
	ctx context.Context,
	input ReleaseDepositCommandInput,
) (ReleaseDepositCommandOutput, error) {
	unitOfWork, beginErr := command.uowFactory.Begin(ctx)
	if beginErr != nil {
		return ReleaseDepositCommandOutput{}, beginErr
	}
	defer func() { _ = unitOfWork.Rollback(ctx) }()

	deposit, findErr := unitOfWork.DepositRepository().FindByID(ctx, input.DepositID)
	if findErr != nil {
		return ReleaseDepositCommandOutput{}, findErr
	}

	if releaseErr := deposit.Release(); releaseErr != nil {
		return ReleaseDepositCommandOutput{}, releaseErr
	}

	ledger := unitOfWork.LedgerRepository()
	accountID, accountErr := buyerLedgerAccountID(ctx, ledger, deposit.UserID(), deposit.Currency())
	if accountErr != nil {
		return ReleaseDepositCommandOutput{}, accountErr
	}

	_, unfreezeErr := ledger.Unfreeze(ctx, ledgerports.UnfreezeInput{
		AccountID:      accountID,
		Amount:         deposit.Amount().AmountInCents(),
		IdempotencyKey: depositLedgerIdempotencyKey("release", input.DepositID),
		Reference:      deposit.Reference(),
		Description:    "deposit released back to available balance",
	})
	if unfreezeErr != nil {
		command.logger.Error().Err(unfreezeErr).
			Uint64("deposit_id", deposit.ID()).
			Msg("failed to unfreeze deposit funds in ledger")

		return ReleaseDepositCommandOutput{}, unfreezeErr
	}

	persisted, saveErr := unitOfWork.DepositRepository().Update(ctx, deposit)
	if saveErr != nil {
		return ReleaseDepositCommandOutput{}, saveErr
	}

	releasedEvent := domainevent.NewDepositReleasedEvent(
		persisted.ID(),
		persisted.UserID(),
		persisted.AuctionID(),
		persisted.Amount(),
		persisted.Currency(),
	)

	outboxEvent, envelopeErr := envelope.FromDepositReleased(releasedEvent)
	if envelopeErr != nil {
		return ReleaseDepositCommandOutput{}, envelopeErr
	}

	if saveErr = unitOfWork.OutboxRepository().Save(ctx, outboxEvent); saveErr != nil {
		return ReleaseDepositCommandOutput{}, saveErr
	}

	if completeErr := unitOfWork.Complete(ctx); completeErr != nil {
		return ReleaseDepositCommandOutput{}, completeErr
	}

	return ReleaseDepositCommandOutput{
		DepositID: persisted.ID(),
		Status:    depositReleasedStatus,
	}, nil
}
