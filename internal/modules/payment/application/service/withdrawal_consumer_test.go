package service

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"auction/internal/modules/payment/domain/enum"
	"auction/internal/modules/payment/domain/event"
	"auction/internal/modules/payment/domain/model"
	"auction/internal/modules/payment/infra/event/envelope"
	"auction/internal/modules/payment/ports"
	ledgermodel "auction/internal/modules/ledger/domain/model"
	ledgerports "auction/internal/modules/ledger/ports"
	"auction/internal/shared/modules/config"
	"auction/tests/mocks"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

const (
	consumerWithdrawalUserID    = uint64(100)
	consumerWithdrawalAccountID = uint64(7)
	consumerWithdrawalAmount    = uint64(5000)
	consumerWithdrawalOutBizNo  = "out-biz-001"
	consumerWithdrawalAlipay    = "user@example.com"
	consumerWithdrawalRealName  = "张三"
)

type WithdrawalConsumerTestSuite struct {
	suite.Suite
	sut            *WithdrawalConsumer
	alipayPortMock *mocks.MockAlipayPort
	uowFactoryMock *mocks.MockPaymentUnitOfWorkFactory
	uowMock        *mocks.MockPaymentUnitOfWork
	ledgerMock     *mocks.MockLedgerRepository
	repoMock       *mocks.MockWithdrawalRepository
	outboxMock     *mocks.MockPaymentOutboxRepository
	loggerMock     *mocks.MockLogger
	alipayCfg      config.Alipay
}

func (s *WithdrawalConsumerTestSuite) SetupTest() {
	s.alipayPortMock = mocks.NewMockAlipayPort(s.T())
	s.uowFactoryMock = mocks.NewMockPaymentUnitOfWorkFactory(s.T())
	s.uowMock = mocks.NewMockPaymentUnitOfWork(s.T())
	s.ledgerMock = mocks.NewMockLedgerRepository(s.T())
	s.repoMock = mocks.NewMockWithdrawalRepository(s.T())
	s.outboxMock = mocks.NewMockPaymentOutboxRepository(s.T())
	s.loggerMock = mocks.NewMockLogger(s.T())
	s.alipayCfg = config.Alipay{Provider: "mock", PlatformAccountOwner: "platform"}
	nopLogger := zerolog.Nop()
	s.loggerMock.On("Info").Return(nopLogger.Info()).Maybe()
	s.loggerMock.On("Error").Return(nopLogger.Error()).Maybe()

	s.sut = NewWithdrawalConsumer(nil, s.alipayPortMock, s.uowFactoryMock, s.alipayCfg, s.loggerMock)
}

func TestWithdrawalConsumerSuite(t *testing.T) {
	suite.Run(t, new(WithdrawalConsumerTestSuite))
}

func (s *WithdrawalConsumerTestSuite) frozenWithdrawal() model.WithdrawalModel {
	withdrawal, err := model.RestoreWithdrawalModel(
		55, consumerWithdrawalUserID, consumerWithdrawalAccountID, consumerWithdrawalAlipay, consumerWithdrawalRealName,
		consumerWithdrawalAmount, "CNY", enum.WithdrawalStatusFrozen, consumerWithdrawalOutBizNo,
		consumerWithdrawalOutBizNo, "", "", 1, time.Now(), time.Now(),
	)
	s.Require().NoError(err)

	return withdrawal
}

func (s *WithdrawalConsumerTestSuite) platformAccount() ledgermodel.AccountModel {
	account, err := ledgermodel.RestoreAccountModel(
		1, "platform", 1_000_000_000, 0, "CNY", 1, time.Now(), time.Now(),
	)
	s.Require().NoError(err)

	return account
}

func (s *WithdrawalConsumerTestSuite) withdrawalPayload() envelope.WithdrawalRequestedPayload {
	return envelope.WithdrawalRequestedPayload{
		EventID:         "evt-w-1",
		WithdrawalID:    55,
		UserID:          consumerWithdrawalUserID,
		LedgerAccountID: consumerWithdrawalAccountID,
		AlipayAccount:   consumerWithdrawalAlipay,
		AlipayRealName:  consumerWithdrawalRealName,
		AmountInCents:   consumerWithdrawalAmount,
		Currency:        "CNY",
		OutBizNo:        consumerWithdrawalOutBizNo,
		FrozenOpID:      consumerWithdrawalOutBizNo,
		OccurredAt:      time.Now().Format(time.RFC3339),
	}
}

func (s *WithdrawalConsumerTestSuite) TestHandle_PayoutSuccess_WithdrawsFrozenBalance() {
	// Arrange
	ctx := context.Background()
	data, marshalErr := json.Marshal(s.withdrawalPayload())
	s.Require().NoError(marshalErr)

	frozen := s.frozenWithdrawal()
	platform := s.platformAccount()

	// Phases 1 and 3 share one uow; Begin is called twice (read + write).
	s.uowFactoryMock.EXPECT().Begin(mock.Anything).Return(s.uowMock, nil).Times(2)
	s.uowMock.EXPECT().WithdrawalRepository().Return(s.repoMock).Times(3)
	s.uowMock.EXPECT().LedgerRepository().Return(s.ledgerMock).Times(1)
	s.uowMock.EXPECT().Rollback(mock.Anything).Return(nil).Times(2)
	s.uowMock.EXPECT().Complete(mock.Anything).Return(nil).Times(1)
	s.repoMock.EXPECT().FindByOutBizNo(mock.Anything, consumerWithdrawalOutBizNo).Return(frozen, nil).Times(2)
	s.repoMock.EXPECT().Update(mock.Anything, mock.Anything).Return(frozen, nil).Times(1)
	s.ledgerMock.EXPECT().GetOrCreateAccountByOwner(mock.Anything, "platform", "CNY").Return(platform, nil).Times(1)
	s.ledgerMock.EXPECT().WithdrawFromFrozen(mock.Anything, mock.MatchedBy(func(in ledgerports.WithdrawFromFrozenInput) bool {
		return in.AccountID == consumerWithdrawalAccountID && in.CounterpartyAccountID == 1 &&
			in.Amount == consumerWithdrawalAmount && in.IdempotencyKey == consumerWithdrawalOutBizNo+":withdraw"
	})).Return(ledgermodel.OperationModel{}, nil).Times(1)
	s.alipayPortMock.EXPECT().TransferToAlipayAccount(mock.Anything, mock.Anything).
		Return(ports.TransferOutput{AlipayOrderID: "alipay-order-1"}, nil).Times(1)
	s.uowMock.EXPECT().OutboxRepository().Return(s.outboxMock).Times(1)
	s.outboxMock.EXPECT().Save(mock.Anything, mock.MatchedBy(func(outboxEvent ports.OutboxEvent) bool {
		return outboxEvent.Subject == event.SubjectWithdrawalCompleted &&
			outboxEvent.EventType == event.WithdrawalCompletedEventType
	})).Return(nil).Times(1)

	// Act
	s.sut.handle(ctx, data)

	// Assert: verified by mock expectations (WithdrawFromFrozen + withdrawal-completed outbox event).
}

func (s *WithdrawalConsumerTestSuite) TestHandle_PayoutFailure_CompensatesByUnfreezing() {
	// Arrange
	ctx := context.Background()
	payload := s.withdrawalPayload()
	payload.EventID = "evt-w-2"
	payload.WithdrawalID = 56
	data, marshalErr := json.Marshal(payload)
	s.Require().NoError(marshalErr)

	frozen := s.frozenWithdrawal()
	platform := s.platformAccount()
	payoutErr := errors.New("alipay payout declined")

	s.uowFactoryMock.EXPECT().Begin(mock.Anything).Return(s.uowMock, nil).Times(2)
	s.uowMock.EXPECT().WithdrawalRepository().Return(s.repoMock).Times(3)
	s.uowMock.EXPECT().LedgerRepository().Return(s.ledgerMock).Times(1)
	s.uowMock.EXPECT().Rollback(mock.Anything).Return(nil).Times(2)
	s.uowMock.EXPECT().Complete(mock.Anything).Return(nil).Times(1)
	s.repoMock.EXPECT().FindByOutBizNo(mock.Anything, consumerWithdrawalOutBizNo).Return(frozen, nil).Times(2)
	s.repoMock.EXPECT().Update(mock.Anything, mock.Anything).Return(frozen, nil).Times(1)
	s.ledgerMock.EXPECT().GetOrCreateAccountByOwner(mock.Anything, "platform", "CNY").Return(platform, nil).Times(1)
	s.ledgerMock.EXPECT().Unfreeze(mock.Anything, mock.MatchedBy(func(in ledgerports.UnfreezeInput) bool {
		return in.AccountID == consumerWithdrawalAccountID &&
			in.Amount == consumerWithdrawalAmount && in.IdempotencyKey == consumerWithdrawalOutBizNo+":unfreeze"
	})).Return(ledgermodel.OperationModel{}, nil).Times(1)
	s.alipayPortMock.EXPECT().TransferToAlipayAccount(mock.Anything, mock.Anything).
		Return(ports.TransferOutput{}, payoutErr).Times(1)
	s.uowMock.EXPECT().OutboxRepository().Return(s.outboxMock).Times(1)
	s.outboxMock.EXPECT().Save(mock.Anything, mock.MatchedBy(func(outboxEvent ports.OutboxEvent) bool {
		return outboxEvent.Subject == event.SubjectWithdrawalFailed &&
			outboxEvent.EventType == event.WithdrawalFailedEventType
	})).Return(nil).Times(1)

	// Act
	s.sut.handle(ctx, data)

	// Assert: verified by mock expectations (Unfreeze compensation + withdrawal-failed outbox event).
}

func (s *WithdrawalConsumerTestSuite) TestHandle_AlreadyTerminal_SkipsSaga() {
	// Arrange
	ctx := context.Background()
	payload := s.withdrawalPayload()
	payload.EventID = "evt-w-3"
	payload.WithdrawalID = 57
	data, marshalErr := json.Marshal(payload)
	s.Require().NoError(marshalErr)

	successWithdrawal, restoreErr := model.RestoreWithdrawalModel(
		57, consumerWithdrawalUserID, consumerWithdrawalAccountID, consumerWithdrawalAlipay, consumerWithdrawalRealName,
		consumerWithdrawalAmount, "CNY", enum.WithdrawalStatusSuccess, consumerWithdrawalOutBizNo,
		consumerWithdrawalOutBizNo, "alipay-order-x", "", 2, time.Now(), time.Now(),
	)
	s.Require().NoError(restoreErr)

	s.uowFactoryMock.EXPECT().Begin(mock.Anything).Return(s.uowMock, nil).Times(1)
	s.uowMock.EXPECT().WithdrawalRepository().Return(s.repoMock).Times(1)
	s.uowMock.EXPECT().Rollback(mock.Anything).Return(nil).Times(1)
	s.repoMock.EXPECT().FindByOutBizNo(mock.Anything, consumerWithdrawalOutBizNo).Return(successWithdrawal, nil).Times(1)

	// Act: a terminal withdrawal must skip the Alipay call and the Saga entirely.
	s.sut.handle(ctx, data)
}
