package command_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/cristiano-pacheco/go-online-auction/internal/modules/auction/application/command"
	"github.com/cristiano-pacheco/go-online-auction/internal/modules/auction/domain/enum"
	"github.com/cristiano-pacheco/go-online-auction/internal/modules/auction/domain/model"
	"github.com/cristiano-pacheco/go-online-auction/tests/mocks"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type StartAuctionCommandTestSuite struct {
	suite.Suite
	sut                               *command.StartAuctionCommand
	uowFactoryMock                    *mocks.MockAuctionUnitOfWorkFactory
	uowMock                           *mocks.MockAuctionUnitOfWork
	auctionRepositoryMock             *mocks.MockAuctionRepository
	auctionStartedEventDispatcherMock *mocks.MockAuctionStartedEventDispatcher
	loggerMock                        *mocks.MockLogger
}

func (s *StartAuctionCommandTestSuite) SetupTest() {
	s.uowFactoryMock = mocks.NewMockAuctionUnitOfWorkFactory(s.T())
	s.uowMock = mocks.NewMockAuctionUnitOfWork(s.T())
	s.auctionRepositoryMock = mocks.NewMockAuctionRepository(s.T())
	s.auctionStartedEventDispatcherMock = mocks.NewMockAuctionStartedEventDispatcher(s.T())
	s.loggerMock = mocks.NewMockLogger(s.T())

	s.sut = command.NewStartAuctionCommand(
		s.uowFactoryMock,
		s.auctionStartedEventDispatcherMock,
		s.loggerMock,
	)
}

func TestStartAuctionCommandSuite(t *testing.T) {
	suite.Run(t, new(StartAuctionCommandTestSuite))
}

func (s *StartAuctionCommandTestSuite) TestExecute_SuccessfulAuctionStart_ReturnsOutput() {
	// Arrange
	ctx := context.Background()
	auctionID := uint64(1)
	listingID := uint64(100)

	input := command.StartAuctionCommandInput{
		AuctionID: auctionID,
	}

	draftState, _ := enum.NewAuctionStateEnum(enum.EnumAuctionStateDraft)
	futureTime := time.Now().Add(24 * time.Hour)
	auction, _ := model.RestoreAuctionModel(
		auctionID,
		listingID,
		nil,
		futureTime,
		draftState,
		nil,
		1,
		time.Now(),
		time.Now(),
	)

	s.uowFactoryMock.On("Begin", mock.Anything).Return(s.uowMock, nil)
	s.uowMock.On("Rollback", mock.Anything).Return(nil)
	s.uowMock.On("AuctionRepository").Return(s.auctionRepositoryMock)
	s.auctionRepositoryMock.On("FindByIDForUpdate", mock.Anything, auctionID).Return(auction, nil)
	s.auctionRepositoryMock.On("Update", mock.Anything, mock.AnythingOfType("model.AuctionModel")).
		Return(nil)
	s.uowMock.On("Complete", mock.Anything).Return(nil)
	s.auctionStartedEventDispatcherMock.On("Dispatch", mock.Anything, mock.Anything).Return(nil)

	// Act
	result, err := s.sut.Execute(ctx, input)

	// Assert
	s.Require().NoError(err)
	s.Equal(auctionID, result.ID)
	s.Equal(listingID, result.ListingID)
	s.Equal(enum.EnumAuctionStateActive, result.State)
	s.NotNil(result.StartTime)
}

func (s *StartAuctionCommandTestSuite) TestExecute_BeginUnitOfWorkFails_ReturnsError() {
	// Arrange
	ctx := context.Background()
	input := command.StartAuctionCommandInput{
		AuctionID: 1,
	}

	expectedErr := errors.New("begin uow error")
	s.uowFactoryMock.On("Begin", mock.Anything).Return(nil, expectedErr)
	s.loggerMock.On("Error").Return(nil)

	// Act
	_, err := s.sut.Execute(ctx, input)

	// Assert
	s.Require().ErrorIs(err, expectedErr)
}

func (s *StartAuctionCommandTestSuite) TestExecute_FindAuctionForUpdateFails_ReturnsError() {
	// Arrange
	ctx := context.Background()
	auctionID := uint64(1)
	input := command.StartAuctionCommandInput{
		AuctionID: auctionID,
	}

	expectedErr := errors.New("auction not found")
	s.uowFactoryMock.On("Begin", mock.Anything).Return(s.uowMock, nil)
	s.uowMock.On("Rollback", mock.Anything).Return(nil)
	s.uowMock.On("AuctionRepository").Return(s.auctionRepositoryMock)
	s.auctionRepositoryMock.On("FindByIDForUpdate", mock.Anything, auctionID).
		Return(model.AuctionModel{}, expectedErr)
	s.loggerMock.On("Error").Return(nil)

	// Act
	_, err := s.sut.Execute(ctx, input)

	// Assert
	s.Require().ErrorIs(err, expectedErr)
}

func (s *StartAuctionCommandTestSuite) TestExecute_StartAuctionFails_ReturnsError() {
	// Arrange
	ctx := context.Background()
	auctionID := uint64(1)
	listingID := uint64(100)

	input := command.StartAuctionCommandInput{
		AuctionID: auctionID,
	}

	activeState, _ := enum.NewAuctionStateEnum(enum.EnumAuctionStateActive)
	futureTime := time.Now().Add(24 * time.Hour)
	now := time.Now()
	auction, _ := model.RestoreAuctionModel(
		auctionID,
		listingID,
		&now,
		futureTime,
		activeState,
		nil,
		1,
		time.Now(),
		time.Now(),
	)

	s.uowFactoryMock.On("Begin", mock.Anything).Return(s.uowMock, nil)
	s.uowMock.On("Rollback", mock.Anything).Return(nil)
	s.uowMock.On("AuctionRepository").Return(s.auctionRepositoryMock)
	s.auctionRepositoryMock.On("FindByIDForUpdate", mock.Anything, auctionID).Return(auction, nil)
	s.loggerMock.On("Error").Return(nil)

	// Act
	_, err := s.sut.Execute(ctx, input)

	// Assert
	s.Require().Error(err)
}

func (s *StartAuctionCommandTestSuite) TestExecute_UpdateAuctionFails_ReturnsError() {
	// Arrange
	ctx := context.Background()
	auctionID := uint64(1)
	listingID := uint64(100)

	input := command.StartAuctionCommandInput{
		AuctionID: auctionID,
	}

	draftState, _ := enum.NewAuctionStateEnum(enum.EnumAuctionStateDraft)
	futureTime := time.Now().Add(24 * time.Hour)
	auction, _ := model.RestoreAuctionModel(
		auctionID,
		listingID,
		nil,
		futureTime,
		draftState,
		nil,
		1,
		time.Now(),
		time.Now(),
	)

	expectedErr := errors.New("update auction error")
	s.uowFactoryMock.On("Begin", mock.Anything).Return(s.uowMock, nil)
	s.uowMock.On("Rollback", mock.Anything).Return(nil)
	s.uowMock.On("AuctionRepository").Return(s.auctionRepositoryMock)
	s.auctionRepositoryMock.On("FindByIDForUpdate", mock.Anything, auctionID).Return(auction, nil)
	s.auctionRepositoryMock.On("Update", mock.Anything, mock.AnythingOfType("model.AuctionModel")).
		Return(expectedErr)
	s.loggerMock.On("Error").Return(nil)

	// Act
	_, err := s.sut.Execute(ctx, input)

	// Assert
	s.Require().ErrorIs(err, expectedErr)
}

func (s *StartAuctionCommandTestSuite) TestExecute_CompleteUnitOfWorkFails_ReturnsError() {
	// Arrange
	ctx := context.Background()
	auctionID := uint64(1)
	listingID := uint64(100)

	input := command.StartAuctionCommandInput{
		AuctionID: auctionID,
	}

	draftState, _ := enum.NewAuctionStateEnum(enum.EnumAuctionStateDraft)
	futureTime := time.Now().Add(24 * time.Hour)
	auction, _ := model.RestoreAuctionModel(
		auctionID,
		listingID,
		nil,
		futureTime,
		draftState,
		nil,
		1,
		time.Now(),
		time.Now(),
	)

	expectedErr := errors.New("complete uow error")
	s.uowFactoryMock.On("Begin", mock.Anything).Return(s.uowMock, nil)
	s.uowMock.On("Rollback", mock.Anything).Return(nil)
	s.uowMock.On("AuctionRepository").Return(s.auctionRepositoryMock)
	s.auctionRepositoryMock.On("FindByIDForUpdate", mock.Anything, auctionID).Return(auction, nil)
	s.auctionRepositoryMock.On("Update", mock.Anything, mock.AnythingOfType("model.AuctionModel")).
		Return(nil)
	s.uowMock.On("Complete", mock.Anything).Return(expectedErr)
	s.loggerMock.On("Error").Return(nil)

	// Act
	_, err := s.sut.Execute(ctx, input)

	// Assert
	s.Require().ErrorIs(err, expectedErr)
}

func (s *StartAuctionCommandTestSuite) TestExecute_DispatchEventFails_ReturnsError() {
	// Arrange
	ctx := context.Background()
	auctionID := uint64(1)
	listingID := uint64(100)

	input := command.StartAuctionCommandInput{
		AuctionID: auctionID,
	}

	draftState, _ := enum.NewAuctionStateEnum(enum.EnumAuctionStateDraft)
	futureTime := time.Now().Add(24 * time.Hour)
	auction, _ := model.RestoreAuctionModel(
		auctionID,
		listingID,
		nil,
		futureTime,
		draftState,
		nil,
		1,
		time.Now(),
		time.Now(),
	)

	expectedErr := errors.New("dispatch event error")
	s.uowFactoryMock.On("Begin", mock.Anything).Return(s.uowMock, nil)
	s.uowMock.On("Rollback", mock.Anything).Return(nil)
	s.uowMock.On("AuctionRepository").Return(s.auctionRepositoryMock)
	s.auctionRepositoryMock.On("FindByIDForUpdate", mock.Anything, auctionID).Return(auction, nil)
	s.auctionRepositoryMock.On("Update", mock.Anything, mock.AnythingOfType("model.AuctionModel")).
		Return(nil)
	s.uowMock.On("Complete", mock.Anything).Return(nil)
	s.auctionStartedEventDispatcherMock.On("Dispatch", mock.Anything, mock.Anything).Return(expectedErr)
	s.loggerMock.On("Error").Return(nil)

	// Act
	_, err := s.sut.Execute(ctx, input)

	// Assert
	s.Require().ErrorIs(err, expectedErr)
}
