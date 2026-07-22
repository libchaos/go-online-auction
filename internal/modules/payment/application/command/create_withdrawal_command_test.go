package command_test

import (
	"context"
	"errors"
	"strconv"
	"testing"
	"time"

	"auction/internal/modules/payment/application/command"
	"auction/internal/modules/payment/domain/enum"
	"auction/internal/modules/payment/domain/errs"
	"auction/internal/modules/payment/domain/model"
	ledgermodel "auction/internal/modules/ledger/domain/model"
	"auction/internal/shared/modules/config"
	"auction/tests/mocks"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

const (
	withdrawalUserID    = uint64(100)
	withdrawalAccountID = uint64(7)
	withdrawalAmount    = uint64(5000)
	withdrawalOutBizNo  = "out-biz-001"
	withdrawalAlipay    = "user@example.com"
	withdrawalRealName  = "张三"
)

type CreateWithdrawalCommandTestSuite struct {
	suite.Suite
	sut               *command.CreateWithdrawalCommand
	uowFactoryMock    *mocks.MockPaymentUnitOfWorkFactory
	uowMock           *mocks.MockPaymentUnitOfWork
	ledgerMock        *mocks.MockLedgerRepository
	withdrawalRepoMock *mocks.MockWithdrawalRepository
	outboxRepoMock    *mocks.MockPaymentOutboxRepository
	loggerMock        *mocks.MockLogger
	alipayCfg         config.Alipay
}

func (s *CreateWithdrawalCommandTestSuite) SetupTest() {
	s.uowFactoryMock = mocks.NewMockPaymentUnitOfWorkFactory(s.T())
	s.uowMock = mocks.NewMockPaymentUnitOfWork(s.T())
	s.ledgerMock = mocks.NewMockLedgerRepository(s.T())
	s.withdrawalRepoMock = mocks.NewMockWithdrawalRepository(s.T())
	s.outboxRepoMock = mocks.NewMockPaymentOutboxRepository(s.T())
	s.loggerMock = mocks.NewMockLogger(s.T())
	s.alipayCfg = config.Alipay{Provider: "mock", PlatformAccountOwner: "platform"}
	nopLogger := zerolog.Nop()
	s.loggerMock.On("Info").Return(nopLogger.Info()).Maybe()
	s.loggerMock.On("Error").Return(nopLogger.Error()).Maybe()

	s.sut = command.NewCreateWithdrawalCommand(s.uowFactoryMock, s.alipayCfg, s.loggerMock)
}

func TestCreateWithdrawalCommandSuite(t *testing.T) {
	suite.Run(t, new(CreateWithdrawalCommandTestSuite))
}

func (s *CreateWithdrawalCommandTestSuite) TestExecute_Success_FreezesAndPersistsFrozenOrder() {
	// Arrange
	ctx := context.Background()
	input := command.CreateWithdrawalCommandInput{
		UserID:         withdrawalUserID,
		AlipayAccount:  withdrawalAlipay,
		AlipayRealName: withdrawalRealName,
		AmountInCents:  withdrawalAmount,
		Currency:       "CNY",
		IdempotencyKey: withdrawalOutBizNo,
	}

	owner := strconv.FormatUint(withdrawalUserID, 10)
	account, accountErr := ledgermodel.RestoreAccountModel(
		withdrawalAccountID, owner, withdrawalAmount, 0, "CNY", 1, time.Now(), time.Now(),
	)
	s.Require().NoError(accountErr)

	persisted, persistedErr := model.RestoreWithdrawalModel(
		55, withdrawalUserID, withdrawalAccountID, withdrawalAlipay, withdrawalRealName,
		withdrawalAmount, "CNY", enum.WithdrawalStatusFrozen, withdrawalOutBizNo,
		withdrawalOutBizNo, "", "", 2, time.Now(), time.Now(),
	)
	s.Require().NoError(persistedErr)

	s.uowFactoryMock.EXPECT().Begin(mock.Anything).Return(s.uowMock, nil)
	s.uowMock.EXPECT().LedgerRepository().Return(s.ledgerMock)
	s.uowMock.EXPECT().WithdrawalRepository().Return(s.withdrawalRepoMock)
	s.uowMock.EXPECT().OutboxRepository().Return(s.outboxRepoMock)
	s.uowMock.EXPECT().Rollback(mock.Anything).Return(nil)
	s.ledgerMock.EXPECT().GetOrCreateAccountByOwner(mock.Anything, mock.Anything, mock.Anything).
		Return(account, nil)
	s.ledgerMock.EXPECT().Freeze(mock.Anything, mock.Anything).Return(ledgermodel.OperationModel{}, nil)
	s.withdrawalRepoMock.EXPECT().Save(mock.Anything, mock.Anything).Return(persisted, nil)
	s.outboxRepoMock.EXPECT().Save(mock.Anything, mock.Anything).Return(nil)
	s.uowMock.EXPECT().Complete(mock.Anything).Return(nil)

	// Act
	output, err := s.sut.Execute(ctx, input)

	// Assert
	s.Require().NoError(err)
	s.Equal(uint64(55), output.WithdrawalID)
	s.Equal(withdrawalOutBizNo, output.OutBizNo)
	s.Equal(string(enum.WithdrawalStatusFrozen), output.Status)
}

func (s *CreateWithdrawalCommandTestSuite) TestExecute_LedgerFreezeFailure_ReturnsError() {
	// Arrange
	ctx := context.Background()
	input := command.CreateWithdrawalCommandInput{
		UserID:         withdrawalUserID,
		AlipayAccount:  withdrawalAlipay,
		AlipayRealName: withdrawalRealName,
		AmountInCents:  withdrawalAmount,
		Currency:       "CNY",
		IdempotencyKey: withdrawalOutBizNo,
	}

	owner := strconv.FormatUint(withdrawalUserID, 10)
	account, accountErr := ledgermodel.RestoreAccountModel(
		withdrawalAccountID, owner, withdrawalAmount, 0, "CNY", 1, time.Now(), time.Now(),
	)
	s.Require().NoError(accountErr)
	freezeErr := errors.New("ledger freeze failed")

	s.uowFactoryMock.EXPECT().Begin(mock.Anything).Return(s.uowMock, nil)
	s.uowMock.EXPECT().LedgerRepository().Return(s.ledgerMock)
	s.uowMock.EXPECT().Rollback(mock.Anything).Return(nil)
	s.ledgerMock.EXPECT().GetOrCreateAccountByOwner(mock.Anything, mock.Anything, mock.Anything).
		Return(account, nil)
	s.ledgerMock.EXPECT().Freeze(mock.Anything, mock.Anything).
		Return(ledgermodel.OperationModel{}, freezeErr)

	// Act
	output, err := s.sut.Execute(ctx, input)

	// Assert
	s.Require().ErrorIs(err, freezeErr)
	s.Equal(uint64(0), output.WithdrawalID)
}

func (s *CreateWithdrawalCommandTestSuite) TestExecute_ZeroAmount_ReturnsValidationError() {
	// Arrange
	ctx := context.Background()
	input := command.CreateWithdrawalCommandInput{
		UserID:         withdrawalUserID,
		AlipayAccount:  withdrawalAlipay,
		AlipayRealName: withdrawalRealName,
		AmountInCents:  0,
		Currency:       "CNY",
		IdempotencyKey: withdrawalOutBizNo,
	}

	// Act
	output, err := s.sut.Execute(ctx, input)

	// Assert
	s.Require().ErrorIs(err, errs.ErrWithdrawalAmountRequired)
	s.Equal(uint64(0), output.WithdrawalID)
}
