package command

import (
	"context"

	domainevent "auction/internal/modules/deposit/domain/event"
	"auction/internal/modules/deposit/infra/event/envelope"
	"auction/internal/modules/deposit/ports"
	"auction/internal/shared/modules/logger"
)

type ForfeitDepositCommandInput struct {
	DepositID uint64
}

type ForfeitDepositCommandOutput struct {
	DepositID uint64
	Status    string
}

const depositForfeitedStatus = "forfeited"

type ForfeitDepositCommand struct {
	uowFactory  ports.DepositUnitOfWorkFactory
	paymentPort ports.PaymentPort
	logger      logger.Logger
}

func NewForfeitDepositCommand(
	uowFactory ports.DepositUnitOfWorkFactory,
	paymentPort ports.PaymentPort,
	logger logger.Logger,
) *ForfeitDepositCommand {
	return &ForfeitDepositCommand{
		uowFactory:  uowFactory,
		paymentPort: paymentPort,
		logger:      logger,
	}
}

func (command *ForfeitDepositCommand) Execute(
	ctx context.Context,
	input ForfeitDepositCommandInput,
) (ForfeitDepositCommandOutput, error) {
	unitOfWork, beginErr := command.uowFactory.Begin(ctx)
	if beginErr != nil {
		return ForfeitDepositCommandOutput{}, beginErr
	}
	defer func() { _ = unitOfWork.Rollback(ctx) }()

	deposit, findErr := unitOfWork.DepositRepository().FindByID(ctx, input.DepositID)
	if findErr != nil {
		return ForfeitDepositCommandOutput{}, findErr
	}

	if forfeitErr := deposit.Forfeit(); forfeitErr != nil {
		return ForfeitDepositCommandOutput{}, forfeitErr
	}

	if externalErr := command.paymentPort.Forfeit(ctx, deposit.ExternalReference()); externalErr != nil {
		command.logger.Error().Err(externalErr).
			Uint64("deposit_id", deposit.ID()).
			Msg("failed to forfeit held funds with payment provider")

		return ForfeitDepositCommandOutput{}, externalErr
	}

	persisted, saveErr := unitOfWork.DepositRepository().Update(ctx, deposit)
	if saveErr != nil {
		return ForfeitDepositCommandOutput{}, saveErr
	}

	forfeitedEvent := domainevent.NewDepositForfeitedEvent(
		persisted.ID(),
		persisted.UserID(),
		persisted.AuctionID(),
		persisted.Amount(),
		persisted.Currency(),
	)

	outboxEvent, envelopeErr := envelope.FromDepositForfeited(forfeitedEvent)
	if envelopeErr != nil {
		return ForfeitDepositCommandOutput{}, envelopeErr
	}

	if saveErr = unitOfWork.OutboxRepository().Save(ctx, outboxEvent); saveErr != nil {
		return ForfeitDepositCommandOutput{}, saveErr
	}

	if completeErr := unitOfWork.Complete(ctx); completeErr != nil {
		return ForfeitDepositCommandOutput{}, completeErr
	}

	return ForfeitDepositCommandOutput{
		DepositID: persisted.ID(),
		Status:    depositForfeitedStatus,
	}, nil
}
