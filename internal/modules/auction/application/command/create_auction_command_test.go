package command_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"auction/internal/modules/auction/application/command"
	"auction/internal/modules/auction/domain/enum"
	"auction/internal/modules/auction/domain/errs"
	"auction/internal/modules/auction/domain/model"
	"auction/internal/modules/auction/domain/strategy"
	"auction/tests/mocks"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type CreateAuctionCommandTestSuite struct {
	suite.Suite
	sut                   *command.CreateAuctionCommand
	uowFactoryMock        *mocks.MockAuctionUnitOfWorkFactory
	uowMock               *mocks.MockAuctionUnitOfWork
	auctionRepositoryMock *mocks.MockAuctionRepository
	outboxRepositoryMock  *mocks.MockOutboxRepository
	listingValidatorMock  *mocks.MockListingValidator
	loggerMock            *mocks.MockLogger
	validListingID        uint64
	validEndTime          time.Time
	invalidEndTime        time.Time
	mockPersistedAuction  model.AuctionModel
	mockPersistedAuctionID uint64
	mockCreatedAt         time.Time
	auctionRepositoryErr  error
}

func (s *CreateAuctionCommandTestSuite) SetupTest() {
	s.uowFactoryMock = mocks.NewMockAuctionUnitOfWorkFactory(s.T())
	s.uowMock = mocks.NewMockAuctionUnitOfWork(s.T())
	s.auctionRepositoryMock = mocks.NewMockAuctionRepository(s.T())
	s.outboxRepositoryMock = mocks.NewMockOutboxRepository(s.T())
	s.listingValidatorMock = mocks.NewMockListingValidator(s.T())
	s.loggerMock = mocks.NewMockLogger(s.T())

	// The unit-of-work method expectations are optional at the suite level:
	// not every test reaches Begin/Commit/Rollback, so mark them Maybe to
	// avoid failing AssertExpectations on paths that never open a transaction.
	s.uowFactoryMock.On("Begin", mock.Anything).Return(s.uowMock, nil).Maybe()
	s.uowMock.On("AuctionRepository").Return(s.auctionRepositoryMock).Maybe()
	s.uowMock.On("OutboxRepository").Return(s.outboxRepositoryMock).Maybe()
	s.uowMock.On("Rollback", mock.Anything).Return(nil).Maybe()
	s.uowMock.On("Complete", mock.Anything).Return(nil).Maybe()

	s.sut = command.NewCreateAuctionCommand(
		s.uowFactoryMock,
		s.listingValidatorMock,
		strategy.NewDefaultResolver(),
		s.loggerMock,
	)

	s.validListingID = 123
	s.validEndTime = time.Now().UTC().Add(24 * time.Hour)
	s.invalidEndTime = time.Now().UTC().Add(-1 * time.Hour)
	s.mockPersistedAuctionID = 456
	s.mockCreatedAt = time.Now().UTC()
	s.auctionRepositoryErr = errors.New("repository error")

	draftState, _ := enum.NewAuctionStateEnum(enum.EnumAuctionStateDraft)
	tradingMode, _ := enum.NewTradingModeEnum(enum.EnumTradingModeEnglish)
	s.mockPersistedAuction, _ = model.RestoreAuctionModelWithMode(
		s.mockPersistedAuctionID,
		s.validListingID,
		nil,
		s.validEndTime,
		draftState,
		tradingMode,
		nil, nil, nil, nil, nil, nil, nil, nil,
		false,
		0,
		1,
		s.mockCreatedAt,
		s.mockCreatedAt,
		strategy.NewDefaultResolver(),
	)
}

func TestCreateAuctionCommandSuite(t *testing.T) {
	suite.Run(t, new(CreateAuctionCommandTestSuite))
}

func (s *CreateAuctionCommandTestSuite) TestExecute_ValidInput_ReturnsCreatedAuction() {
	// Arrange
	ctx := context.Background()
	input := command.CreateAuctionCommandInput{
		ListingID:   s.validListingID,
		EndTime:     s.validEndTime,
		TradingMode: "english",
	}

	s.listingValidatorMock.
		On("IsAuctionable", mock.Anything, s.validListingID).
		Return(true, nil)

	s.auctionRepositoryMock.
		On("Create", mock.Anything, mock.AnythingOfType("model.AuctionModel")).
		Return(s.mockPersistedAuction, nil)

	s.outboxRepositoryMock.
		On("Save", mock.Anything, mock.Anything).
		Return(nil)

	// Act
	output, err := s.sut.Execute(ctx, input)

	// Assert
	s.Require().NoError(err)
	s.Equal(s.mockPersistedAuctionID, output.ID)
	s.Equal(s.validListingID, output.ListingID)
	s.Equal(enum.EnumAuctionStateDraft, output.State)
	s.Equal("english", output.TradingMode)
	s.Equal(s.validEndTime.Unix(), output.EndTime.Unix())
	s.Equal(s.mockCreatedAt.Unix(), output.CreatedAt.Unix())
	s.outboxRepositoryMock.AssertCalled(s.T(), "Save", mock.Anything, mock.Anything)
	s.uowMock.AssertCalled(s.T(), "Complete", mock.Anything)
}

func (s *CreateAuctionCommandTestSuite) TestExecute_InvalidListingID_ReturnsError() {
	// Arrange
	ctx := context.Background()
	input := command.CreateAuctionCommandInput{
		ListingID:   0,
		EndTime:     s.validEndTime,
		TradingMode: "english",
	}

	s.listingValidatorMock.
		On("IsAuctionable", mock.Anything, uint64(0)).
		Return(true, nil)

	s.loggerMock.On("Error").Return(nil).Maybe()

	// Act
	output, err := s.sut.Execute(ctx, input)

	// Assert
	s.Require().ErrorIs(err, errs.ErrListingIDRequired)
	s.Equal(command.CreateAuctionCommandOutput{}, output)
}

func (s *CreateAuctionCommandTestSuite) TestExecute_InvalidEndTime_ReturnsError() {
	// Arrange
	ctx := context.Background()
	input := command.CreateAuctionCommandInput{
		ListingID:   s.validListingID,
		EndTime:     s.invalidEndTime,
		TradingMode: "english",
	}

	s.listingValidatorMock.
		On("IsAuctionable", mock.Anything, s.validListingID).
		Return(true, nil)

	s.loggerMock.On("Error").Return(nil).Maybe()

	// Act
	output, err := s.sut.Execute(ctx, input)

	// Assert
	s.Require().ErrorIs(err, errs.ErrEndTimeMustBeInFuture)
	s.Equal(command.CreateAuctionCommandOutput{}, output)
}

func (s *CreateAuctionCommandTestSuite) TestExecute_RepositoryError_ReturnsError() {
	// Arrange
	ctx := context.Background()
	input := command.CreateAuctionCommandInput{
		ListingID:   s.validListingID,
		EndTime:     s.validEndTime,
		TradingMode: "english",
	}

	s.listingValidatorMock.
		On("IsAuctionable", mock.Anything, s.validListingID).
		Return(true, nil)

	emptyAuction := model.AuctionModel{}
	s.auctionRepositoryMock.
		On("Create", mock.Anything, mock.AnythingOfType("model.AuctionModel")).
		Return(emptyAuction, s.auctionRepositoryErr)

	s.loggerMock.On("Error").Return(nil).Maybe()

	// Act
	output, err := s.sut.Execute(ctx, input)

	// Assert
	s.Require().ErrorIs(err, s.auctionRepositoryErr)
	s.Equal(command.CreateAuctionCommandOutput{}, output)
	s.outboxRepositoryMock.AssertNotCalled(s.T(), "Save", mock.Anything, mock.Anything)
}

func (s *CreateAuctionCommandTestSuite) TestExecute_OutboxSaveError_ReturnsError() {
	// Arrange
	ctx := context.Background()
	input := command.CreateAuctionCommandInput{
		ListingID:   s.validListingID,
		EndTime:     s.validEndTime,
		TradingMode: "english",
	}

	s.listingValidatorMock.
		On("IsAuctionable", mock.Anything, s.validListingID).
		Return(true, nil)

	s.auctionRepositoryMock.
		On("Create", mock.Anything, mock.AnythingOfType("model.AuctionModel")).
		Return(s.mockPersistedAuction, nil)

	outboxErr := errors.New("outbox error")
	s.outboxRepositoryMock.
		On("Save", mock.Anything, mock.Anything).
		Return(outboxErr)

	s.loggerMock.On("Error").Return(nil).Maybe()

	// Act
	output, err := s.sut.Execute(ctx, input)

	// Assert
	s.Require().ErrorIs(err, outboxErr)
	s.Equal(command.CreateAuctionCommandOutput{}, output)
}

func (s *CreateAuctionCommandTestSuite) TestExecute_ListingNotAuctionable_ReturnsError() {
	// Arrange
	ctx := context.Background()
	input := command.CreateAuctionCommandInput{
		ListingID:   s.validListingID,
		EndTime:     s.validEndTime,
		TradingMode: "english",
	}

	s.listingValidatorMock.
		On("IsAuctionable", mock.Anything, s.validListingID).
		Return(false, nil)

	// Act
	output, err := s.sut.Execute(ctx, input)

	// Assert
	s.Require().ErrorIs(err, errs.ErrListingNotAvailable)
	s.Equal(command.CreateAuctionCommandOutput{}, output)
	s.uowFactoryMock.AssertNotCalled(s.T(), "Begin", mock.Anything)
}

func (s *CreateAuctionCommandTestSuite) TestExecute_ListingValidatorError_ReturnsError() {
	// Arrange
	ctx := context.Background()
	input := command.CreateAuctionCommandInput{
		ListingID:   s.validListingID,
		EndTime:     s.validEndTime,
		TradingMode: "english",
	}

	validatorErr := errors.New("validator error")
	s.listingValidatorMock.
		On("IsAuctionable", mock.Anything, s.validListingID).
		Return(false, validatorErr)

	s.loggerMock.On("Error").Return(nil).Maybe()

	// Act
	output, err := s.sut.Execute(ctx, input)

	// Assert
	s.Require().ErrorIs(err, validatorErr)
	s.Equal(command.CreateAuctionCommandOutput{}, output)
	s.uowFactoryMock.AssertNotCalled(s.T(), "Begin", mock.Anything)
}
