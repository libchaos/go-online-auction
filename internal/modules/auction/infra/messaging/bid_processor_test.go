package messaging_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"auction/internal/modules/auction/domain/enum"
	"auction/internal/modules/auction/domain/errs"
	"auction/internal/modules/auction/domain/model"
	"auction/internal/modules/auction/domain/strategy"
	"auction/internal/modules/auction/infra/messaging"
	"auction/internal/modules/auction/ports"
	depositerrs "auction/internal/modules/deposit/domain/errs"
	depositmocks "auction/internal/modules/deposit/testmocks"
	"auction/tests/mocks"
)

type BidProcessorTestSuite struct {
	suite.Suite
	sut                   *messaging.BidProcessor
	uowFactoryMock        *mocks.MockAuctionUnitOfWorkFactory
	uowMock               *mocks.MockAuctionUnitOfWork
	auctionRepositoryMock *mocks.MockAuctionRepository
	bidRepositoryMock     *mocks.MockBidRepository
	outboxRepositoryMock  *mocks.MockOutboxRepository
	depositGuardMock      *depositmocks.MockDepositGuard
	loggerMock            *mocks.MockLogger
}

func (s *BidProcessorTestSuite) SetupTest() {
	s.uowFactoryMock = mocks.NewMockAuctionUnitOfWorkFactory(s.T())
	s.uowMock = mocks.NewMockAuctionUnitOfWork(s.T())
	s.auctionRepositoryMock = mocks.NewMockAuctionRepository(s.T())
	s.bidRepositoryMock = mocks.NewMockBidRepository(s.T())
	s.outboxRepositoryMock = mocks.NewMockOutboxRepository(s.T())
	s.depositGuardMock = depositmocks.NewMockDepositGuard(s.T())
	s.loggerMock = mocks.NewMockLogger(s.T())

	// js is only used by Start/sendToDLQ, which these tests never reach.
	s.sut = messaging.NewBidProcessor(nil, s.uowFactoryMock, strategy.GetResolver(), s.depositGuardMock, s.loggerMock)
}

func TestBidProcessorSuite(t *testing.T) {
	suite.Run(t, new(BidProcessorTestSuite))
}

func (s *BidProcessorTestSuite) activeAuction(auctionID uint64) model.AuctionModel {
	activeState, _ := enum.NewAuctionStateEnum(enum.EnumAuctionStateActive)
	futureTime := time.Now().Add(24 * time.Hour)
	auction, _ := model.RestoreAuctionModel(
		auctionID,
		1,
		nil,
		futureTime,
		activeState,
		nil,
		1,
		time.Now(),
		time.Now(),
	)
	return auction
}

func (s *BidProcessorTestSuite) command(auctionID uint64) ports.BidCommand {
	return ports.BidCommand{
		IdempotencyKey: "key-1",
		AuctionID:      auctionID,
		UserID:         100,
		AmountInCents:  5000,
		IssuedAt:       time.Now().UTC(),
	}
}

func (s *BidProcessorTestSuite) TestProcessBid_Success_PersistsAndDispatches() {
	// Arrange
	ctx := context.Background()
	auctionID := uint64(1)
	auction := s.activeAuction(auctionID)
	money := model.NewMoneyModel(5000)
	persistedBid, _ := model.RestoreBidModel(10, auctionID, 100, money, time.Now(), time.Now())

	s.depositGuardMock.On("EnsureEligible", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	s.uowFactoryMock.On("Begin", mock.Anything).Return(s.uowMock, nil)
	s.uowMock.On("Rollback", mock.Anything).Return(nil)
	s.uowMock.On("AuctionRepository").Return(s.auctionRepositoryMock)
	s.uowMock.On("BidRepository").Return(s.bidRepositoryMock)
	s.auctionRepositoryMock.On("FindByID", mock.Anything, auctionID).Return(auction, nil)
	s.bidRepositoryMock.On("Create", mock.Anything, mock.AnythingOfType("model.BidModel"), "key-1").
		Return(persistedBid, nil)
	s.auctionRepositoryMock.On("Update", mock.Anything, mock.AnythingOfType("model.AuctionModel")).Return(nil)
	s.uowMock.On("OutboxRepository").Return(s.outboxRepositoryMock)
	s.outboxRepositoryMock.On("Save", mock.Anything, mock.Anything).Return(nil)
	s.uowMock.On("Complete", mock.Anything).Return(nil)

	// Act
	err := s.sut.ProcessBid(ctx, s.command(auctionID))

	// Assert
	s.Require().NoError(err)
}

func (s *BidProcessorTestSuite) TestProcessBid_DuplicateIdempotencyKey_ReturnsDuplicateError() {
	// Arrange
	ctx := context.Background()
	auctionID := uint64(1)
	auction := s.activeAuction(auctionID)

	s.depositGuardMock.On("EnsureEligible", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	s.uowFactoryMock.On("Begin", mock.Anything).Return(s.uowMock, nil)
	s.uowMock.On("Rollback", mock.Anything).Return(nil)
	s.uowMock.On("AuctionRepository").Return(s.auctionRepositoryMock)
	s.uowMock.On("BidRepository").Return(s.bidRepositoryMock)
	s.auctionRepositoryMock.On("FindByID", mock.Anything, auctionID).Return(auction, nil)
	s.bidRepositoryMock.On("Create", mock.Anything, mock.AnythingOfType("model.BidModel"), "key-1").
		Return(model.BidModel{}, errs.ErrBidDuplicateIdempotencyKey)

	// Act
	err := s.sut.ProcessBid(ctx, s.command(auctionID))

	// Assert
	s.Require().ErrorIs(err, errs.ErrBidDuplicateIdempotencyKey)
}

func (s *BidProcessorTestSuite) TestProcessBid_TransientError_ReturnsError() {
	// Arrange
	ctx := context.Background()
	auctionID := uint64(1)
	transientErr := errors.New("db timeout")

	s.depositGuardMock.On("EnsureEligible", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	s.uowFactoryMock.On("Begin", mock.Anything).Return(s.uowMock, nil)
	s.uowMock.On("Rollback", mock.Anything).Return(nil)
	s.uowMock.On("AuctionRepository").Return(s.auctionRepositoryMock)
	s.auctionRepositoryMock.On("FindByID", mock.Anything, auctionID).Return(model.AuctionModel{}, transientErr)

	// Act
	err := s.sut.ProcessBid(ctx, s.command(auctionID))

	// Assert
	s.Require().ErrorIs(err, transientErr)
}

func (s *BidProcessorTestSuite) TestProcessBid_IneligibleDeposit_ReturnsGuardError() {
	// Arrange
	ctx := context.Background()
	auctionID := uint64(1)

	s.depositGuardMock.On("EnsureEligible", mock.Anything, mock.Anything, mock.Anything).
		Return(depositerrs.ErrDepositNotHeld)

	// Act
	err := s.sut.ProcessBid(ctx, s.command(auctionID))

	// Assert
	s.Require().ErrorIs(err, depositerrs.ErrDepositNotHeld)
}

func TestIsPermanentBidError(t *testing.T) {
	t.Run("permanent domain errors are classified as permanent", func(t *testing.T) {
		// Arrange
		permanentErrors := []error{
			errs.ErrBidMustExceedHighest,
			errs.ErrFirstBidMustBePositive,
			errs.ErrBidsOnlyOnActiveAuctions,
			errs.ErrAuctionExpired,
			errs.ErrAuctionNotFound,
			errs.ErrInvalidAuctionState,
			errs.ErrDutchBidMustMatchPrice,
			errs.ErrDutchPriceNotAvailable,
			errs.ErrFixedPriceMismatch,
			errs.ErrFixedPriceNotConfigured,
			errs.ErrProxyMaxTooLow,
			errs.ErrStartingPriceRequired,
			depositerrs.ErrDepositRequired,
			depositerrs.ErrDepositInsufficient,
			depositerrs.ErrDepositNotHeld,
		}

		// Act / Assert
		for _, permanentErr := range permanentErrors {
			require.True(t, messaging.IsPermanentBidError(permanentErr), permanentErr.Error())
		}
	})

	t.Run("transient errors are not permanent", func(t *testing.T) {
		// Arrange
		transientErr := errors.New("connection reset")

		// Act
		result := messaging.IsPermanentBidError(transientErr)

		// Assert
		require.False(t, result)
	})
}
