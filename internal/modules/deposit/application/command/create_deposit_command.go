package command

import (
	"context"

	"github.com/google/uuid"

	"auction/internal/modules/deposit/domain/errs"
	domainevent "auction/internal/modules/deposit/domain/event"
	"auction/internal/modules/deposit/domain/model"
	"auction/internal/modules/deposit/infra/event/envelope"
	"auction/internal/modules/deposit/ports"
	"auction/internal/shared/modules/logger"
)

type CreateDepositCommandInput struct {
	UserID         uint64
	AuctionID      uint64
	AmountInCents  uint64
	Currency       string
	IdempotencyKey string
}

type CreateDepositCommandOutput struct {
	DepositID uint64
	Status    string
}

const depositHeldStatus = "held"

type CreateDepositCommand struct {
	uowFactory  ports.DepositUnitOfWorkFactory
	paymentPort ports.PaymentPort
	logger      logger.Logger
}

func NewCreateDepositCommand(
	uowFactory ports.DepositUnitOfWorkFactory,
	paymentPort ports.PaymentPort,
	logger logger.Logger,
) *CreateDepositCommand {
	return &CreateDepositCommand{
		uowFactory:  uowFactory,
		paymentPort: paymentPort,
		logger:      logger,
	}
}

func (command *CreateDepositCommand) Execute(
	ctx context.Context,
	input CreateDepositCommandInput,
) (CreateDepositCommandOutput, error) {
	reference := input.IdempotencyKey
	if reference == "" {
		reference = uuid.NewString()
	}

	amount := model.NewMoneyModel(input.AmountInCents)

	deposit, buildErr := model.NewDeposit(input.UserID, input.AuctionID, amount, input.Currency, reference)
	if buildErr != nil {
		return CreateDepositCommandOutput{}, buildErr
	}

	externalReference, holdErr := command.paymentPort.Hold(ctx, input.UserID, amount, input.Currency, reference)
	if holdErr != nil {
		command.logger.Error().Err(holdErr).
			Uint64("user_id", input.UserID).
			Uint64("auction_id", input.AuctionID).
			Msg("failed to hold deposit funds with payment provider")

		return CreateDepositCommandOutput{}, errs.ErrDepositExternalFailure
	}

	if confirmErr := deposit.ConfirmHold(externalReference); confirmErr != nil {
		return CreateDepositCommandOutput{}, confirmErr
	}

	unitOfWork, beginErr := command.uowFactory.Begin(ctx)
	if beginErr != nil {
		return CreateDepositCommandOutput{}, beginErr
	}
	defer func() { _ = unitOfWork.Rollback(ctx) }()

	persisted, saveErr := unitOfWork.DepositRepository().Save(ctx, deposit)
	if saveErr != nil {
		return CreateDepositCommandOutput{}, saveErr
	}

	heldEvent := domainevent.NewDepositHeldEvent(
		persisted.ID(),
		persisted.UserID(),
		persisted.AuctionID(),
		persisted.Amount(),
		persisted.Currency(),
		persisted.ExternalReference(),
	)

	outboxEvent, envelopeErr := envelope.FromDepositHeld(heldEvent)
	if envelopeErr != nil {
		return CreateDepositCommandOutput{}, envelopeErr
	}

	if saveErr = unitOfWork.OutboxRepository().Save(ctx, outboxEvent); saveErr != nil {
		return CreateDepositCommandOutput{}, saveErr
	}

	if completeErr := unitOfWork.Complete(ctx); completeErr != nil {
		return CreateDepositCommandOutput{}, completeErr
	}

	command.logger.Info().
		Uint64("deposit_id", persisted.ID()).
		Uint64("user_id", input.UserID).
		Uint64("auction_id", input.AuctionID).
		Msg("deposit held successfully")

	return CreateDepositCommandOutput{
		DepositID: persisted.ID(),
		Status:    depositHeldStatus,
	}, nil
}
