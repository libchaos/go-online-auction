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
	"auction/tests/mocks"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type CreateAuctionCommandTestSuite struct {
	suite.Suite
	sut                    *command.CreateAuctionCommand
	auctionRepositoryMock  *mocks.MockAuctionRepository
	listingValidatorMock   *mocks.MockListingValidator
	loggerMock             *mocks.MockLogger
	auctionRepositoryErr   error
	validListingID         uint64
	validEndTime           time.Time
	invalidEndTime         time.Time
	mockPersistedAuction   model.AuctionModel
	mockPersistedAuctionID uint64
	mockCreatedAt          time.Time
	mockAuctionState       enum.AuctionStateEnum
}

func (s *CreateAuctionCommandTestSuite) SetupTest() {
	s.auctionRepositoryMock = mocks.NewMockAuctionRepository(s.T())
	s.listingValidatorMock = mocks.NewMockListingValidator(s.T())
	s.loggerMock = mocks.NewMockLogger(s.T())

	s.sut = command.NewCreateAuctionCommand(
		s.auctionRepositoryMock,
		s.listingValidatorMock,
		s.loggerMock,
	)

	s.validListingID = 123
	s.validEndTime = time.Now().UTC().Add(24 * time.Hour)
	s.invalidEndTime = time.Now().UTC().Add(-1 * time.Hour)
	s.mockPersistedAuctionID = 456
	s.mockCreatedAt = time.Now().UTC()
	s.auctionRepositoryErr = errors.New("repository error")

	draftState, _ := enum.NewAuctionStateEnum(enum.EnumAuctionStateDraft)
	s.mockAuctionState = draftState

	s.mockPersistedAuction, _ = model.RestoreAuctionModel(
		s.mockPersistedAuctionID,
		s.validListingID,
		nil,
		s.validEndTime,
		s.mockAuctionState,
		nil,
		1,
		s.mockCreatedAt,
		s.mockCreatedAt,
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

	s.loggerMock.On("Error").Return(nil)

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

	s.loggerMock.On("Error").Return(nil)

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

	s.loggerMock.On("Error").Return(nil)

	// Act
	output, err := s.sut.Execute(ctx, input)

	// Assert
	s.Require().ErrorIs(err, s.auctionRepositoryErr)
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
	s.auctionRepositoryMock.AssertNotCalled(s.T(), "Create")
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

	s.loggerMock.On("Error").Return(nil)

	// Act
	output, err := s.sut.Execute(ctx, input)

	// Assert
	s.Require().ErrorIs(err, validatorErr)
	s.Equal(command.CreateAuctionCommandOutput{}, output)
	s.auctionRepositoryMock.AssertNotCalled(s.T(), "Create")
}
