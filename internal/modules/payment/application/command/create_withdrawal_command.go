package command

import (
	"context"
	"strconv"

	"github.com/google/uuid"

	ledgerports "auction/internal/modules/ledger/ports"
	"auction/internal/modules/payment/domain/errs"
	"auction/internal/modules/payment/domain/event"
	"auction/internal/modules/payment/domain/model"
	"auction/internal/modules/payment/infra/event/envelope"
	"auction/internal/modules/payment/ports"
	"auction/internal/shared/modules/config"
	"auction/internal/shared/modules/logger"
)

type CreateWithdrawalCommandInput struct {
	UserID         uint64
	AlipayAccount  string
	AlipayRealName string
	AmountInCents  uint64
	Currency       string
	IdempotencyKey string
}

type CreateWithdrawalCommandOutput struct {
	WithdrawalID uint64
	OutBizNo     string
	Status       string
}

type CreateWithdrawalCommand struct {
	uowFactory ports.PaymentUnitOfWorkFactory
	alipayCfg  config.Alipay
	logger     logger.Logger
}

func NewCreateWithdrawalCommand(
	uowFactory ports.PaymentUnitOfWorkFactory,
	alipayCfg config.Alipay,
	logger logger.Logger,
) *CreateWithdrawalCommand {
	return &CreateWithdrawalCommand{
		uowFactory: uowFactory,
		alipayCfg:  alipayCfg,
		logger:     logger,
	}
}

// Execute freezes the user's ledger balance and records a withdrawal order in
// the FROZEN state, then writes a withdrawal-requested event to the outbox. The
// withdrawal consumer later calls Alipay and runs the payout Saga.
func (command *CreateWithdrawalCommand) Execute(
	ctx context.Context,
	input CreateWithdrawalCommandInput,
) (CreateWithdrawalCommandOutput, error) {
	currency := input.Currency
	if currency == "" {
		currency = defaultCurrency
	}

	if validationErr := validateWithdrawalInput(input); validationErr != nil {
		return CreateWithdrawalCommandOutput{}, validationErr
	}

	outBizNo := input.IdempotencyKey
	if outBizNo == "" {
		outBizNo = uuid.NewString()
	}

	unitOfWork, beginErr := command.uowFactory.Begin(ctx)
	if beginErr != nil {
		return CreateWithdrawalCommandOutput{}, beginErr
	}
	defer func() { _ = unitOfWork.Rollback(ctx) }()

	ledger := unitOfWork.LedgerRepository()

	owner := strconv.FormatUint(input.UserID, 10)
	account, accountErr := ledger.GetOrCreateAccountByOwner(ctx, owner, currency)
	if accountErr != nil {
		return CreateWithdrawalCommandOutput{}, accountErr
	}

	withdrawal, buildErr := model.NewWithdrawal(
		input.UserID,
		account.ID(),
		input.AlipayAccount,
		input.AlipayRealName,
		input.AmountInCents,
		currency,
		outBizNo,
	)
	if buildErr != nil {
		return CreateWithdrawalCommandOutput{}, buildErr
	}

	// Freeze the user's funds. The same out_biz_no is used as the freeze
	// idempotency key (stored as frozen_op_id) and as the Alipay transfer
	// idempotency key downstream. Subsequent Saga steps use derived keys so
	// they do not collide on the ledger_operations idempotency constraint.
	if _, freezeErr := ledger.Freeze(ctx, ledgerports.FreezeInput{
		AccountID:      account.ID(),
		Amount:         input.AmountInCents,
		IdempotencyKey: outBizNo,
		Reference:      outBizNo,
		Description:    "withdrawal freeze for user " + owner,
	}); freezeErr != nil {
		command.logger.Error().Err(freezeErr).Uint64("user_id", input.UserID).
			Msg("failed to freeze withdrawal funds in ledger")

		return CreateWithdrawalCommandOutput{}, freezeErr
	}

	if markErr := withdrawal.MarkFrozen(outBizNo); markErr != nil {
		return CreateWithdrawalCommandOutput{}, markErr
	}

	persisted, persistErr := command.persistFrozenWithdrawal(ctx, unitOfWork, withdrawal, outBizNo)
	if persistErr != nil {
		return CreateWithdrawalCommandOutput{}, persistErr
	}

	command.logger.Info().Uint64("withdrawal_id", persisted.ID()).Uint64("user_id", input.UserID).
		Msg("withdrawal order created and funds frozen")

	return CreateWithdrawalCommandOutput{
		WithdrawalID: persisted.ID(),
		OutBizNo:     persisted.OutBizNo(),
		Status:       string(persisted.Status()),
	}, nil
}

// persistFrozenWithdrawal saves the frozen withdrawal order, writes the
// withdrawal-requested outbox event, and commits the unit of work.
func (command *CreateWithdrawalCommand) persistFrozenWithdrawal(
	ctx context.Context,
	unitOfWork ports.PaymentUnitOfWork,
	withdrawal model.WithdrawalModel,
	outBizNo string,
) (model.WithdrawalModel, error) {
	persisted, saveErr := unitOfWork.WithdrawalRepository().Save(ctx, withdrawal)
	if saveErr != nil {
		return model.WithdrawalModel{}, saveErr
	}

	requestedEvent := event.NewWithdrawalRequestedEvent(
		persisted.ID(),
		persisted.UserID(),
		persisted.LedgerAccountID(),
		persisted.AlipayAccount(),
		persisted.AlipayRealName(),
		persisted.AmountInCents(),
		persisted.Currency(),
		persisted.OutBizNo(),
		outBizNo,
	)

	outboxEvent, envelopeErr := envelope.ToWithdrawalRequestedOutboxEvent(requestedEvent)
	if envelopeErr != nil {
		return model.WithdrawalModel{}, envelopeErr
	}

	if saveErr = unitOfWork.OutboxRepository().Save(ctx, outboxEvent); saveErr != nil {
		return model.WithdrawalModel{}, saveErr
	}

	if completeErr := unitOfWork.Complete(ctx); completeErr != nil {
		return model.WithdrawalModel{}, completeErr
	}

	return persisted, nil
}

func validateWithdrawalInput(input CreateWithdrawalCommandInput) error {
	if input.UserID == 0 {
		return errs.ErrWithdrawalUserRequired
	}
	if input.AlipayAccount == "" {
		return errs.ErrWithdrawalAlipayAccountRequired
	}
	if input.AlipayRealName == "" {
		return errs.ErrWithdrawalAlipayRealNameRequired
	}
	if input.AmountInCents == 0 {
		return errs.ErrWithdrawalAmountRequired
	}

	return nil
}
