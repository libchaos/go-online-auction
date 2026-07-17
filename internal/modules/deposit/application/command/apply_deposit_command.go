package command

import (
	"context"

	"auction/internal/modules/deposit/domain/errs"
	domainevent "auction/internal/modules/deposit/domain/event"
	"auction/internal/modules/deposit/domain/model"
	"auction/internal/modules/deposit/infra/event/envelope"
	"auction/internal/modules/deposit/ports"
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
	uowFactory  ports.DepositUnitOfWorkFactory
	paymentPort ports.PaymentPort
	logger      logger.Logger
}

func NewApplyDepositCommand(
	uowFactory ports.DepositUnitOfWorkFactory,
	paymentPort ports.PaymentPort,
	logger logger.Logger,
) *ApplyDepositCommand {
	return &ApplyDepositCommand{
		uowFactory:  uowFactory,
		paymentPort: paymentPort,
		logger:      logger,
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

	if captureErr := command.paymentPort.Capture(ctx, deposit.ExternalReference(), captureAmount); captureErr != nil {
		command.logger.Error().Err(captureErr).
			Uint64("deposit_id", deposit.ID()).
			Msg("failed to capture held funds with payment provider")

		return ApplyDepositCommandOutput{}, captureErr
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
