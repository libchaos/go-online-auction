package command

import (
	"context"

	domainevent "auction/internal/modules/deposit/domain/event"
	"auction/internal/modules/deposit/infra/event/envelope"
	"auction/internal/modules/deposit/ports"
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
	uowFactory  ports.DepositUnitOfWorkFactory
	paymentPort ports.PaymentPort
	logger      logger.Logger
}

func NewReleaseDepositCommand(
	uowFactory ports.DepositUnitOfWorkFactory,
	paymentPort ports.PaymentPort,
	logger logger.Logger,
) *ReleaseDepositCommand {
	return &ReleaseDepositCommand{
		uowFactory:  uowFactory,
		paymentPort: paymentPort,
		logger:      logger,
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

	if externalErr := command.paymentPort.Release(ctx, deposit.ExternalReference()); externalErr != nil {
		command.logger.Error().Err(externalErr).
			Uint64("deposit_id", deposit.ID()).
			Msg("failed to release held funds with payment provider")

		return ReleaseDepositCommandOutput{}, externalErr
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
