package command

import (
	"context"

	domainevent "auction/internal/modules/deposit/domain/event"
	"auction/internal/modules/deposit/infra/event/envelope"
	"auction/internal/modules/deposit/ports"
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

	if cancelErr := deposit.Cancel(); cancelErr != nil {
		return CancelDepositCommandOutput{}, cancelErr
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
