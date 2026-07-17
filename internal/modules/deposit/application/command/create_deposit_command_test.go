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
	depositmocks "auction/internal/modules/deposit/testmocks"
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
	uowFactoryMock *depositmocks.MockDepositUnitOfWorkFactory
	uowMock        *depositmocks.MockDepositUnitOfWork
	depositRepoMock *depositmocks.MockDepositRepository
	outboxRepoMock *depositmocks.MockOutboxRepository
	paymentMock    *depositmocks.MockPaymentPort
	loggerMock     *mocks.MockLogger
}

func (s *CreateDepositCommandTestSuite) SetupTest() {
	s.uowFactoryMock = depositmocks.NewMockDepositUnitOfWorkFactory(s.T())
	s.uowMock = depositmocks.NewMockDepositUnitOfWork(s.T())
	s.depositRepoMock = depositmocks.NewMockDepositRepository(s.T())
	s.outboxRepoMock = depositmocks.NewMockOutboxRepository(s.T())
	s.paymentMock = depositmocks.NewMockPaymentPort(s.T())
	s.loggerMock = mocks.NewMockLogger(s.T())
	nopLogger := zerolog.Nop()
	s.loggerMock.On("Info").Return(nopLogger.Info()).Maybe()
	s.loggerMock.On("Error").Return(nopLogger.Error()).Maybe()

	s.sut = command.NewCreateDepositCommand(s.uowFactoryMock, s.paymentMock, s.loggerMock)
}

func TestCreateDepositCommandSuite(t *testing.T) {
	suite.Run(t, new(CreateDepositCommandTestSuite))
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

	persisted, err := model.RestoreDepositModel(
		42, commandUserID, commandAuctionID, commandAmount, "CNY",
		enum.EnumDepositStatusHeld, "ext-ref", "ref", 2, time.Now(), time.Now(),
	)
	s.Require().NoError(err)

	s.paymentMock.On("Hold", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return("ext-ref", nil)
	s.uowFactoryMock.On("Begin", mock.Anything).Return(s.uowMock, nil)
	s.uowMock.On("DepositRepository").Return(s.depositRepoMock)
	s.uowMock.On("OutboxRepository").Return(s.outboxRepoMock)
	s.uowMock.On("Rollback", mock.Anything).Return(nil)
	s.depositRepoMock.On("Save", mock.Anything, mock.Anything).Return(persisted, nil)
	s.outboxRepoMock.On("Save", mock.Anything, mock.Anything).Return(nil)
	s.uowMock.On("Complete", mock.Anything).Return(nil)

	// Act
	output, err := s.sut.Execute(ctx, input)

	// Assert
	s.Require().NoError(err)
	s.Equal(uint64(42), output.DepositID)
	s.Equal("held", output.Status)
}

func (s *CreateDepositCommandTestSuite) TestExecute_PaymentFailure_ReturnsExternalFailure() {
	// Arrange
	ctx := context.Background()
	input := command.CreateDepositCommandInput{
		UserID:        commandUserID,
		AuctionID:     commandAuctionID,
		AmountInCents: commandAmount,
		Currency:      "CNY",
	}

	s.paymentMock.On("Hold", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return("", errors.New("provider down"))

	// Act
	output, err := s.sut.Execute(ctx, input)

	// Assert
	s.Require().ErrorIs(err, errs.ErrDepositExternalFailure)
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

	saveErr := errors.New("db write failed")
	s.paymentMock.On("Hold", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return("ext-ref", nil)
	s.uowFactoryMock.On("Begin", mock.Anything).Return(s.uowMock, nil)
	s.uowMock.On("DepositRepository").Return(s.depositRepoMock)
	s.uowMock.On("Rollback", mock.Anything).Return(nil)
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
