package command_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"auction/internal/modules/deposit/application/command"
	"auction/internal/modules/deposit/domain/enum"
	"auction/internal/modules/deposit/domain/errs"
	"auction/internal/modules/deposit/domain/model"
	ledgermodel "auction/internal/modules/ledger/domain/model"
	ledgerenum "auction/internal/modules/ledger/domain/enum"
	"auction/tests/mocks"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

const (
	commandUserID    = uint64(100)
	commandAuctionID = uint64(200)
	commandAmount    = uint64(5000)
)

type CreateDepositCommandTestSuite struct {
	suite.Suite
	sut            *command.CreateDepositCommand
	uowFactoryMock *mocks.MockDepositUnitOfWorkFactory
	uowMock        *mocks.MockDepositUnitOfWork
	ledgerRepoMock *mocks.MockLedgerRepository
	depositRepoMock *mocks.MockDepositRepository
	outboxRepoMock *mocks.MockDepositOutboxRepository
	loggerMock     *mocks.MockLogger
}

func (s *CreateDepositCommandTestSuite) SetupTest() {
	s.uowFactoryMock = mocks.NewMockDepositUnitOfWorkFactory(s.T())
	s.uowMock = mocks.NewMockDepositUnitOfWork(s.T())
	s.ledgerRepoMock = mocks.NewMockLedgerRepository(s.T())
	s.depositRepoMock = mocks.NewMockDepositRepository(s.T())
	s.outboxRepoMock = mocks.NewMockDepositOutboxRepository(s.T())
	s.loggerMock = mocks.NewMockLogger(s.T())
	nopLogger := zerolog.Nop()
	s.loggerMock.On("Info").Return(nopLogger.Info()).Maybe()
	s.loggerMock.On("Error").Return(nopLogger.Error()).Maybe()

	s.sut = command.NewCreateDepositCommand(s.uowFactoryMock, s.loggerMock)
}

func TestCreateDepositCommandSuite(t *testing.T) {
	suite.Run(t, new(CreateDepositCommandTestSuite))
}

func (s *CreateDepositCommandTestSuite) ledgerAccount() ledgermodel.AccountModel {
	account, err := ledgermodel.RestoreAccountModel(
		7, "100", commandAmount, 0, "CNY", 1, time.Now(), time.Now(),
	)
	s.Require().NoError(err)

	return account
}

func (s *CreateDepositCommandTestSuite) ledgerFreezeOperation() ledgermodel.OperationModel {
	operationType, err := ledgerenum.NewOperationTypeEnum(ledgerenum.EnumOperationTypeFreeze)
	s.Require().NoError(err)

	operationStatus, err := ledgerenum.NewOperationStatusEnum(ledgerenum.EnumOperationStatusCommitted)
	s.Require().NoError(err)

	operation := ledgermodel.RestoreOperationModel(
		9, 7, nil, operationType, commandAmount,
		"ledger-op-key-9", operationStatus, "ref",
		"deposit hold for auction 200", time.Now(), time.Now(),
	)

	return operation
}

func (s *CreateDepositCommandTestSuite) TestExecute_Success_HoldsPersistsAndReturnsHeld() {
	// Arrange
	ctx := context.Background()
	input := command.CreateDepositCommandInput{
		UserID:        commandUserID,
		AuctionID:     commandAuctionID,
		AmountInCents: commandAmount,
		Currency:      "CNY",
	}

	account := s.ledgerAccount()
	operation := s.ledgerFreezeOperation()

	persisted, err := model.RestoreDepositModel(
		42, commandUserID, commandAuctionID, commandAmount, "CNY",
		enum.EnumDepositStatusHeld, "ext-ref", "ref", 2, time.Now(), time.Now(),
	)
	s.Require().NoError(err)

	s.uowFactoryMock.On("Begin", mock.Anything).Return(s.uowMock, nil)
	s.uowMock.On("LedgerRepository").Return(s.ledgerRepoMock)
	s.uowMock.On("DepositRepository").Return(s.depositRepoMock)
	s.uowMock.On("OutboxRepository").Return(s.outboxRepoMock)
	s.uowMock.On("Rollback", mock.Anything).Return(nil)
	s.ledgerRepoMock.On("GetOrCreateAccountByOwner", mock.Anything, mock.Anything, mock.Anything).
		Return(account, nil)
	s.ledgerRepoMock.On("Freeze", mock.Anything, mock.Anything).Return(operation, nil)
	s.depositRepoMock.On("Save", mock.Anything, mock.Anything).Return(persisted, nil)
	s.outboxRepoMock.On("Save", mock.Anything, mock.Anything).Return(nil)
	s.uowMock.On("Complete", mock.Anything).Return(nil)

	// Act
	output, err := s.sut.Execute(ctx, input)

	// Assert
	s.Require().NoError(err)
	s.Equal(uint64(42), output.DepositID)
	s.Equal("held", output.Status)
	s.Equal(uint64(7), output.AccountID)
}

func (s *CreateDepositCommandTestSuite) TestExecute_LedgerFreezeFailure_ReturnsError() {
	// Arrange
	ctx := context.Background()
	input := command.CreateDepositCommandInput{
		UserID:        commandUserID,
		AuctionID:     commandAuctionID,
		AmountInCents: commandAmount,
		Currency:      "CNY",
	}

	account := s.ledgerAccount()
	freezeErr := errors.New("ledger freeze failed")

	s.uowFactoryMock.On("Begin", mock.Anything).Return(s.uowMock, nil)
	s.uowMock.On("LedgerRepository").Return(s.ledgerRepoMock)
	s.uowMock.On("Rollback", mock.Anything).Return(nil)
	s.ledgerRepoMock.On("GetOrCreateAccountByOwner", mock.Anything, mock.Anything, mock.Anything).
		Return(account, nil)
	s.ledgerRepoMock.On("Freeze", mock.Anything, mock.Anything).
		Return(ledgermodel.OperationModel{}, freezeErr)

	// Act
	output, err := s.sut.Execute(ctx, input)

	// Assert
	s.Require().ErrorIs(err, freezeErr)
	s.Equal(uint64(0), output.DepositID)
}

func (s *CreateDepositCommandTestSuite) TestExecute_SaveFailure_ReturnsError() {
	// Arrange
	ctx := context.Background()
	input := command.CreateDepositCommandInput{
		UserID:        commandUserID,
		AuctionID:     commandAuctionID,
		AmountInCents: commandAmount,
		Currency:      "CNY",
	}

	account := s.ledgerAccount()
	operation := s.ledgerFreezeOperation()
	saveErr := errors.New("db write failed")

	s.uowFactoryMock.On("Begin", mock.Anything).Return(s.uowMock, nil)
	s.uowMock.On("LedgerRepository").Return(s.ledgerRepoMock)
	s.uowMock.On("DepositRepository").Return(s.depositRepoMock)
	s.uowMock.On("Rollback", mock.Anything).Return(nil)
	s.ledgerRepoMock.On("GetOrCreateAccountByOwner", mock.Anything, mock.Anything, mock.Anything).
		Return(account, nil)
	s.ledgerRepoMock.On("Freeze", mock.Anything, mock.Anything).Return(operation, nil)
	s.depositRepoMock.On("Save", mock.Anything, mock.Anything).Return(model.DepositModel{}, saveErr)

	// Act
	output, err := s.sut.Execute(ctx, input)

	// Assert
	s.Require().ErrorIs(err, saveErr)
	s.Equal(uint64(0), output.DepositID)
}

func (s *CreateDepositCommandTestSuite) TestExecute_ZeroAmount_ReturnsValidationError() {
	// Arrange
	ctx := context.Background()
	input := command.CreateDepositCommandInput{
		UserID:        commandUserID,
		AuctionID:     commandAuctionID,
		AmountInCents: 0,
		Currency:      "CNY",
	}

	// Act
	output, err := s.sut.Execute(ctx, input)

	// Assert
	s.Require().ErrorIs(err, errs.ErrDepositAmountRequired)
	s.Equal(uint64(0), output.DepositID)
}
