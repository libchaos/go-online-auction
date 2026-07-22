package command

import (
	"context"

	"auction/internal/modules/deposit/domain/enum"
	domainevent "auction/internal/modules/deposit/domain/event"
	"auction/internal/modules/deposit/infra/event/envelope"
	"auction/internal/modules/deposit/ports"
	ledgerports "auction/internal/modules/ledger/ports"
	"auction/internal/shared/modules/logger"
)

type CancelDepositCommandInput struct {
	DepositID uint64
}

type CancelDepositCommandOutput struct {
	DepositID uint64
	Status    string
}

type CancelDepositCommand struct {
	uowFactory ports.DepositUnitOfWorkFactory
	logger     logger.Logger
}

func NewCancelDepositCommand(
	uowFactory ports.DepositUnitOfWorkFactory,
	logger logger.Logger,
) *CancelDepositCommand {
	return &CancelDepositCommand{
		uowFactory: uowFactory,
		logger:     logger,
	}
}

func (command *CancelDepositCommand) Execute(
	ctx context.Context,
	input CancelDepositCommandInput,
) (CancelDepositCommandOutput, error) {
	unitOfWork, beginErr := command.uowFactory.Begin(ctx)
	if beginErr != nil {
		return CancelDepositCommandOutput{}, beginErr
	}
	defer func() { _ = unitOfWork.Rollback(ctx) }()

	deposit, findErr := unitOfWork.DepositRepository().FindByID(ctx, input.DepositID)
	if findErr != nil {
		return CancelDepositCommandOutput{}, findErr
	}

	currentStatus := deposit.Status()
	wasHeld := currentStatus.String() == enum.EnumDepositStatusHeld

	if cancelErr := deposit.Cancel(); cancelErr != nil {
		return CancelDepositCommandOutput{}, cancelErr
	}

	if wasHeld {
		ledger := unitOfWork.LedgerRepository()
		accountID, accountErr := buyerLedgerAccountID(ctx, ledger, deposit.UserID(), deposit.Currency())
		if accountErr != nil {
			return CancelDepositCommandOutput{}, accountErr
		}

		_, unfreezeErr := ledger.Unfreeze(ctx, ledgerports.UnfreezeInput{
			AccountID:      accountID,
			Amount:         deposit.Amount().AmountInCents(),
			IdempotencyKey: depositLedgerIdempotencyKey("cancel", input.DepositID),
			Reference:      deposit.Reference(),
			Description:    "deposit cancelled, released back to available balance",
		})
		if unfreezeErr != nil {
			command.logger.Error().Err(unfreezeErr).
				Uint64("deposit_id", deposit.ID()).
				Msg("failed to unfreeze deposit funds in ledger on cancel")

			return CancelDepositCommandOutput{}, unfreezeErr
		}
	}

	persisted, saveErr := unitOfWork.DepositRepository().Update(ctx, deposit)
	if saveErr != nil {
		return CancelDepositCommandOutput{}, saveErr
	}

	cancelledEvent := domainevent.NewDepositReleasedEvent(
		persisted.ID(),
		persisted.UserID(),
		persisted.AuctionID(),
		persisted.Amount(),
		persisted.Currency(),
	)

	outboxEvent, envelopeErr := envelope.FromDepositReleased(cancelledEvent)
	if envelopeErr != nil {
		return CancelDepositCommandOutput{}, envelopeErr
	}

	if saveErr = unitOfWork.OutboxRepository().Save(ctx, outboxEvent); saveErr != nil {
		return CancelDepositCommandOutput{}, saveErr
	}

	if completeErr := unitOfWork.Complete(ctx); completeErr != nil {
		return CancelDepositCommandOutput{}, completeErr
	}

	return CancelDepositCommandOutput{
		DepositID: persisted.ID(),
		Status:    depositReleasedStatus,
	}, nil
}
