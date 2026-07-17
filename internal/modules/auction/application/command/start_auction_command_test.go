package command_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"auction/internal/modules/auction/application/command"
	"auction/internal/modules/auction/domain/enum"
	"auction/internal/modules/auction/domain/event"
	"auction/internal/modules/auction/domain/model"
	"auction/internal/modules/auction/ports"
	"auction/tests/mocks"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type StartAuctionCommandTestSuite struct {
	suite.Suite
	sut                   *command.StartAuctionCommand
	uowFactoryMock        *mocks.MockAuctionUnitOfWorkFactory
	uowMock               *mocks.MockAuctionUnitOfWork
	auctionRepositoryMock *mocks.MockAuctionRepository
	outboxRepositoryMock  *mocks.MockOutboxRepository
	loggerMock            *mocks.MockLogger
}

func (s *StartAuctionCommandTestSuite) SetupTest() {
	s.uowFactoryMock = mocks.NewMockAuctionUnitOfWorkFactory(s.T())
	s.uowMock = mocks.NewMockAuctionUnitOfWork(s.T())
	s.auctionRepositoryMock = mocks.NewMockAuctionRepository(s.T())
	s.outboxRepositoryMock = mocks.NewMockOutboxRepository(s.T())
	s.loggerMock = mocks.NewMockLogger(s.T())

	s.sut = command.NewStartAuctionCommand(
		s.uowFactoryMock,
		s.loggerMock,
	)
}

func TestStartAuctionCommandSuite(t *testing.T) {
	suite.Run(t, new(StartAuctionCommandTestSuite))
}

func (s *StartAuctionCommandTestSuite) draftAuction(auctionID, listingID uint64) model.AuctionModel {
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
	return auction
}

func (s *StartAuctionCommandTestSuite) TestExecute_SuccessfulAuctionStart_ReturnsOutput() {
	// Arrange
	ctx := context.Background()
	auctionID := uint64(1)
	listingID := uint64(100)

	input := command.StartAuctionCommandInput{
		AuctionID: auctionID,
	}

	auction := s.draftAuction(auctionID, listingID)

	s.uowFactoryMock.On("Begin", mock.Anything).Return(s.uowMock, nil)
	s.uowMock.On("Rollback", mock.Anything).Return(nil)
	s.uowMock.On("AuctionRepository").Return(s.auctionRepositoryMock)
	s.uowMock.On("OutboxRepository").Return(s.outboxRepositoryMock)
	s.auctionRepositoryMock.On("FindByIDForUpdate", mock.Anything, auctionID).Return(auction, nil)
	s.auctionRepositoryMock.On("Update", mock.Anything, mock.AnythingOfType("model.AuctionModel")).
		Return(nil)
	s.outboxRepositoryMock.On("Save", mock.Anything, mock.MatchedBy(func(evt ports.OutboxEvent) bool {
		return evt.EventType == event.AuctionStartedEventType
	})).Return(nil)
	s.uowMock.On("Complete", mock.Anything).Return(nil)

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
	auction := s.draftAuction(auctionID, 100)

	input := command.StartAuctionCommandInput{
		AuctionID: auctionID,
	}

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

func (s *StartAuctionCommandTestSuite) TestExecute_SaveOutboxFails_ReturnsError() {
	// Arrange
	ctx := context.Background()
	auctionID := uint64(1)
	auction := s.draftAuction(auctionID, 100)

	input := command.StartAuctionCommandInput{
		AuctionID: auctionID,
	}

	expectedErr := errors.New("save outbox error")
	s.uowFactoryMock.On("Begin", mock.Anything).Return(s.uowMock, nil)
	s.uowMock.On("Rollback", mock.Anything).Return(nil)
	s.uowMock.On("AuctionRepository").Return(s.auctionRepositoryMock)
	s.uowMock.On("OutboxRepository").Return(s.outboxRepositoryMock)
	s.auctionRepositoryMock.On("FindByIDForUpdate", mock.Anything, auctionID).Return(auction, nil)
	s.auctionRepositoryMock.On("Update", mock.Anything, mock.AnythingOfType("model.AuctionModel")).
		Return(nil)
	s.outboxRepositoryMock.On("Save", mock.Anything, mock.Anything).Return(expectedErr)
	s.loggerMock.On("Error").Return(nil)

	// Act
	_, err := s.sut.Execute(ctx, input)

	// Assert: transaction is rolled back, so the event is never delivered
	s.Require().ErrorIs(err, expectedErr)
	s.uowMock.AssertNotCalled(s.T(), "Complete", mock.Anything)
}

func (s *StartAuctionCommandTestSuite) TestExecute_CompleteUnitOfWorkFails_ReturnsError() {
	// Arrange
	ctx := context.Background()
	auctionID := uint64(1)
	auction := s.draftAuction(auctionID, 100)

	input := command.StartAuctionCommandInput{
		AuctionID: auctionID,
	}

	expectedErr := errors.New("complete uow error")
	s.uowFactoryMock.On("Begin", mock.Anything).Return(s.uowMock, nil)
	s.uowMock.On("Rollback", mock.Anything).Return(nil)
	s.uowMock.On("AuctionRepository").Return(s.auctionRepositoryMock)
	s.uowMock.On("OutboxRepository").Return(s.outboxRepositoryMock)
	s.auctionRepositoryMock.On("FindByIDForUpdate", mock.Anything, auctionID).Return(auction, nil)
	s.auctionRepositoryMock.On("Update", mock.Anything, mock.AnythingOfType("model.AuctionModel")).
		Return(nil)
	s.outboxRepositoryMock.On("Save", mock.Anything, mock.Anything).Return(nil)
	s.uowMock.On("Complete", mock.Anything).Return(expectedErr)
	s.loggerMock.On("Error").Return(nil)

	// Act
	_, err := s.sut.Execute(ctx, input)

	// Assert
	s.Require().ErrorIs(err, expectedErr)
}
