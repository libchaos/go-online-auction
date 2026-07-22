package command

import (
	"context"

	"auction/internal/modules/deposit/domain/errs"
	domainevent "auction/internal/modules/deposit/domain/event"
	"auction/internal/modules/deposit/domain/model"
	"auction/internal/modules/deposit/infra/event/envelope"
	"auction/internal/modules/deposit/ports"
	ledgerports "auction/internal/modules/ledger/ports"
	"auction/internal/shared/modules/logger"
)

type ApplyDepositCommandInput struct {
	DepositID            uint64
	CaptureAmountInCents uint64
}

type ApplyDepositCommandOutput struct {
	DepositID uint64
	Status    string
}

const depositAppliedStatus = "applied"

type ApplyDepositCommand struct {
	uowFactory ports.DepositUnitOfWorkFactory
	logger     logger.Logger
}

func NewApplyDepositCommand(
	uowFactory ports.DepositUnitOfWorkFactory,
	logger logger.Logger,
) *ApplyDepositCommand {
	return &ApplyDepositCommand{
		uowFactory: uowFactory,
		logger:     logger,
	}
}

func (command *ApplyDepositCommand) Execute(
	ctx context.Context,
	input ApplyDepositCommandInput,
) (ApplyDepositCommandOutput, error) {
	unitOfWork, beginErr := command.uowFactory.Begin(ctx)
	if beginErr != nil {
		return ApplyDepositCommandOutput{}, beginErr
	}
	defer func() { _ = unitOfWork.Rollback(ctx) }()

	deposit, findErr := unitOfWork.DepositRepository().FindByID(ctx, input.DepositID)
	if findErr != nil {
		return ApplyDepositCommandOutput{}, findErr
	}

	captureAmount := deposit.Amount()
	if input.CaptureAmountInCents > 0 {
		captureAmount = model.NewMoneyModel(input.CaptureAmountInCents)
	}

	if captureAmount.IsGreaterThan(deposit.Amount()) {
		return ApplyDepositCommandOutput{}, errs.ErrDepositCaptureTooLarge
	}

	if applyErr := deposit.ApplyToWinning(); applyErr != nil {
		return ApplyDepositCommandOutput{}, applyErr
	}

	ledger := unitOfWork.LedgerRepository()
	if settleErr := command.settleToPlatform(ctx, ledger, deposit, captureAmount); settleErr != nil {
		command.logger.Error().Err(settleErr).
			Uint64("deposit_id", deposit.ID()).
			Msg("failed to withdraw frozen deposit funds in ledger")

		return ApplyDepositCommandOutput{}, settleErr
	}

	persisted, saveErr := unitOfWork.DepositRepository().Update(ctx, deposit)
	if saveErr != nil {
		return ApplyDepositCommandOutput{}, saveErr
	}

	appliedEvent := domainevent.NewDepositAppliedEvent(
		persisted.ID(),
		persisted.UserID(),
		persisted.AuctionID(),
		persisted.Amount(),
		persisted.Currency(),
	)

	outboxEvent, envelopeErr := envelope.FromDepositApplied(appliedEvent)
	if envelopeErr != nil {
		return ApplyDepositCommandOutput{}, envelopeErr
	}

	if saveErr = unitOfWork.OutboxRepository().Save(ctx, outboxEvent); saveErr != nil {
		return ApplyDepositCommandOutput{}, saveErr
	}

	if completeErr := unitOfWork.Complete(ctx); completeErr != nil {
		return ApplyDepositCommandOutput{}, completeErr
	}

	return ApplyDepositCommandOutput{
		DepositID: persisted.ID(),
		Status:    depositAppliedStatus,
	}, nil
}

func (command *ApplyDepositCommand) settleToPlatform(
	ctx context.Context,
	ledger ledgerports.LedgerRepository,
	deposit model.DepositModel,
	captureAmount model.MoneyModel,
) error {
	buyerAccountID, accountErr := buyerLedgerAccountID(ctx, ledger, deposit.UserID(), deposit.Currency())
	if accountErr != nil {
		return accountErr
	}

	platformAccountID, platformErr := platformLedgerAccountID(ctx, ledger, deposit.Currency())
	if platformErr != nil {
		return platformErr
	}

	_, withdrawErr := ledger.WithdrawFromFrozen(ctx, ledgerports.WithdrawFromFrozenInput{
		AccountID:             buyerAccountID,
		CounterpartyAccountID: platformAccountID,
		Amount:                captureAmount.AmountInCents(),
		IdempotencyKey:        depositLedgerIdempotencyKey("apply", deposit.ID()),
		Reference:             deposit.Reference(),
		Description:           "deposit applied to winning bid, settled to platform",
	})

	return withdrawErr
}
