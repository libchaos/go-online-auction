package command

import (
	"context"
	"strconv"

	"github.com/google/uuid"

	domainevent "auction/internal/modules/deposit/domain/event"
	"auction/internal/modules/deposit/domain/model"
	"auction/internal/modules/deposit/infra/event/envelope"
	"auction/internal/modules/deposit/ports"
	ledgerports "auction/internal/modules/ledger/ports"
	"auction/internal/shared/modules/logger"
)

const defaultCurrency = "CNY"

const depositHeldStatus = "held"

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
	AccountID uint64
}

type CreateDepositCommand struct {
	uowFactory ports.DepositUnitOfWorkFactory
	logger     logger.Logger
}

func NewCreateDepositCommand(
	uowFactory ports.DepositUnitOfWorkFactory,
	logger logger.Logger,
) *CreateDepositCommand {
	return &CreateDepositCommand{
		uowFactory: uowFactory,
		logger:     logger,
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

	currency := input.Currency
	if currency == "" {
		currency = defaultCurrency
	}

	amount := model.NewMoneyModel(input.AmountInCents)

	deposit, buildErr := model.NewDeposit(input.UserID, input.AuctionID, amount, currency, reference)
	if buildErr != nil {
		return CreateDepositCommandOutput{}, buildErr
	}

	unitOfWork, beginErr := command.uowFactory.Begin(ctx)
	if beginErr != nil {
		return CreateDepositCommandOutput{}, beginErr
	}
	defer func() { _ = unitOfWork.Rollback(ctx) }()

	ledger := unitOfWork.LedgerRepository()
	owner := strconv.FormatUint(input.UserID, 10)

	account, accountErr := ledger.GetOrCreateAccountByOwner(ctx, owner, currency)
	if accountErr != nil {
		command.logger.Error().Err(accountErr).
			Str("owner", owner).
			Msg("failed to resolve ledger account for deposit")

		return CreateDepositCommandOutput{}, accountErr
	}

	operation, freezeErr := ledger.Freeze(ctx, ledgerports.FreezeInput{
		AccountID:      account.ID(),
		Amount:         input.AmountInCents,
		IdempotencyKey: reference,
		Reference:      reference,
		Description:    "deposit hold for auction " + strconv.FormatUint(input.AuctionID, 10),
	})
	if freezeErr != nil {
		command.logger.Error().Err(freezeErr).
			Uint64("user_id", input.UserID).
			Uint64("auction_id", input.AuctionID).
			Msg("failed to freeze deposit funds in ledger")

		return CreateDepositCommandOutput{}, freezeErr
	}

	if confirmErr := deposit.ConfirmHold(operation.IdempotencyKey()); confirmErr != nil {
		return CreateDepositCommandOutput{}, confirmErr
	}

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
		Msg("deposit held successfully in ledger")

	return CreateDepositCommandOutput{
		DepositID: persisted.ID(),
		Status:    depositHeldStatus,
		AccountID: account.ID(),
	}, nil
}
