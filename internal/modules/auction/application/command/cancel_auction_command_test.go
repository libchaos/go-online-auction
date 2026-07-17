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

type CancelAuctionCommandTestSuite struct {
	suite.Suite
	uowFactoryMock        *mocks.MockAuctionUnitOfWorkFactory
	uowMock               *mocks.MockAuctionUnitOfWork
	auctionRepositoryMock *mocks.MockAuctionRepository
	loggerMock            *mocks.MockLogger
	sut                   *command.CancelAuctionCommand
}

func (s *CancelAuctionCommandTestSuite) SetupTest() {
	s.uowFactoryMock = mocks.NewMockAuctionUnitOfWorkFactory(s.T())
	s.uowMock = mocks.NewMockAuctionUnitOfWork(s.T())
	s.auctionRepositoryMock = mocks.NewMockAuctionRepository(s.T())
	s.loggerMock = mocks.NewMockLogger(s.T())

	s.sut = command.NewCancelAuctionCommand(s.uowFactoryMock, s.loggerMock)
}

func TestCancelAuctionCommandSuite(t *testing.T) {
	suite.Run(t, new(CancelAuctionCommandTestSuite))
}

func (s *CancelAuctionCommandTestSuite) TestExecute_ValidInput_CancelsAuction() {
	// Arrange
	ctx := context.Background()
	input := command.CancelAuctionCommandInput{
		AuctionID: 10,
	}

	listingID := uint64(100)
	now := time.Now().UTC()
	startTime := now.Add(-1 * time.Hour)
	endTime := now.Add(24 * time.Hour)
	state, _ := enum.NewAuctionStateEnum(enum.EnumAuctionStateActive)

	// Use RestoreAuctionModel to have ID set
	auction, _ := model.RestoreAuctionModel(
		input.AuctionID, listingID, &startTime, endTime, state, nil, 1, now, now,
	)

	s.uowFactoryMock.On("Begin", ctx).Return(s.uowMock, nil)
	s.uowMock.On("Rollback", ctx).Return(nil)
	s.uowMock.On("AuctionRepository").Return(s.auctionRepositoryMock)
	s.auctionRepositoryMock.On("FindByIDForUpdate", ctx, input.AuctionID).Return(auction, nil)
	s.auctionRepositoryMock.On("Update", ctx, mock.MatchedBy(func(a model.AuctionModel) bool {
		state := a.State()
		return state.String() == enum.EnumAuctionStateCancelled
	})).Return(nil)
	s.uowMock.On("Complete", ctx).Return(nil)

	// Act
	output, err := s.sut.Execute(ctx, input)

	// Assert
	s.NoError(err)
	s.Equal(input.AuctionID, output.ID)
	s.Equal(listingID, output.ListingID)
	s.Equal(enum.EnumAuctionStateCancelled, output.State)
	s.NotNil(output.StartTime)
	s.Equal(startTime, *output.StartTime)
	s.Equal(endTime, output.EndTime)
	s.True(output.UpdatedAt.After(now) || output.UpdatedAt.Equal(now), "UpdatedAt should be updated")
}

func (s *CancelAuctionCommandTestSuite) TestExecute_BeginUOWFails_ReturnsError() {
	// Arrange
	ctx := context.Background()
	input := command.CancelAuctionCommandInput{
		AuctionID: 10,
	}
	expectedErr := errors.New("begin error")

	s.uowFactoryMock.On("Begin", ctx).Return(s.uowMock, expectedErr)
	s.loggerMock.On("Error").Return(nil)

	// Act
	_, err := s.sut.Execute(ctx, input)

	// Assert
	s.ErrorIs(err, expectedErr)
}

func (s *CancelAuctionCommandTestSuite) TestExecute_FindAuctionFails_ReturnsError() {
	// Arrange
	ctx := context.Background()
	input := command.CancelAuctionCommandInput{
		AuctionID: 10,
	}
	expectedErr := errors.New("db error")

	s.uowFactoryMock.On("Begin", ctx).Return(s.uowMock, nil)
	s.uowMock.On("Rollback", ctx).Return(nil)
	s.uowMock.On("AuctionRepository").Return(s.auctionRepositoryMock)
	s.auctionRepositoryMock.On("FindByIDForUpdate", ctx, input.AuctionID).Return(model.AuctionModel{}, expectedErr)
	s.loggerMock.On("Error").Return(nil)

	// Act
	_, err := s.sut.Execute(ctx, input)

	// Assert
	s.ErrorIs(err, expectedErr)
}

func (s *CancelAuctionCommandTestSuite) TestExecute_CancelFails_ReturnsError() {
	// Arrange
	ctx := context.Background()
	input := command.CancelAuctionCommandInput{
		AuctionID: 10,
	}

	listingID := uint64(100)
	now := time.Now().UTC()
	startTime := now.Add(-1 * time.Hour)
	endTime := now.Add(24 * time.Hour)
	state, _ := enum.NewAuctionStateEnum(enum.EnumAuctionStateClosed)

	// Auction is already closed
	auction, _ := model.RestoreAuctionModel(
		input.AuctionID, listingID, &startTime, endTime, state, nil, 1, now, now,
	)

	s.uowFactoryMock.On("Begin", ctx).Return(s.uowMock, nil)
	s.uowMock.On("Rollback", ctx).Return(nil)
	s.uowMock.On("AuctionRepository").Return(s.auctionRepositoryMock)
	s.auctionRepositoryMock.On("FindByIDForUpdate", ctx, input.AuctionID).Return(auction, nil)
	s.loggerMock.On("Error").Return(nil)

	// Act
	_, err := s.sut.Execute(ctx, input)

	// Assert
	s.ErrorIs(err, errs.ErrAuctionCanOnlyCancelFromDraftOrActive)
}

func (s *CancelAuctionCommandTestSuite) TestExecute_UpdateFails_ReturnsError() {
	// Arrange
	ctx := context.Background()
	input := command.CancelAuctionCommandInput{
		AuctionID: 10,
	}
	expectedErr := errors.New("update error")

	listingID := uint64(100)
	now := time.Now().UTC()
	startTime := now.Add(-1 * time.Hour)
	endTime := now.Add(24 * time.Hour)
	state, _ := enum.NewAuctionStateEnum(enum.EnumAuctionStateActive)

	// Use RestoreAuctionModel
	auction, _ := model.RestoreAuctionModel(
		input.AuctionID, listingID, &startTime, endTime, state, nil, 1, now, now,
	)

	s.uowFactoryMock.On("Begin", ctx).Return(s.uowMock, nil)
	s.uowMock.On("Rollback", ctx).Return(nil)
	s.uowMock.On("AuctionRepository").Return(s.auctionRepositoryMock)
	s.auctionRepositoryMock.On("FindByIDForUpdate", ctx, input.AuctionID).Return(auction, nil)
	s.auctionRepositoryMock.On("Update", ctx, mock.Anything).Return(expectedErr)
	s.loggerMock.On("Error").Return(nil)

	// Act
	_, err := s.sut.Execute(ctx, input)

	// Assert
	s.ErrorIs(err, expectedErr)
}

func (s *CancelAuctionCommandTestSuite) TestExecute_CompleteFails_ReturnsError() {
	// Arrange
	ctx := context.Background()
	input := command.CancelAuctionCommandInput{
		AuctionID: 10,
	}
	expectedErr := errors.New("commit error")

	listingID := uint64(100)
	now := time.Now().UTC()
	startTime := now.Add(-1 * time.Hour)
	endTime := now.Add(24 * time.Hour)
	state, _ := enum.NewAuctionStateEnum(enum.EnumAuctionStateActive)

	// Use RestoreAuctionModel
	auction, _ := model.RestoreAuctionModel(
		input.AuctionID, listingID, &startTime, endTime, state, nil, 1, now, now,
	)

	s.uowFactoryMock.On("Begin", ctx).Return(s.uowMock, nil)
	s.uowMock.On("Rollback", ctx).Return(nil)
	s.uowMock.On("AuctionRepository").Return(s.auctionRepositoryMock)
	s.auctionRepositoryMock.On("FindByIDForUpdate", ctx, input.AuctionID).Return(auction, nil)
	s.auctionRepositoryMock.On("Update", ctx, mock.Anything).Return(nil)
	s.uowMock.On("Complete", ctx).Return(expectedErr)
	s.loggerMock.On("Error").Return(nil)

	// Act
	_, err := s.sut.Execute(ctx, input)

	// Assert
	s.ErrorIs(err, expectedErr)
}
