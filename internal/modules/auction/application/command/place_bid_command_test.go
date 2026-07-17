package command_test

import (
	"context"
	"errors"
	"testing"

	"auction/internal/modules/auction/application/command"
	"auction/internal/modules/auction/domain/errs"
	"auction/internal/modules/auction/ports"
	"auction/tests/mocks"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type PlaceBidCommandTestSuite struct {
	suite.Suite
	sut                     *command.PlaceBidCommand
	bidCommandPublisherMock *mocks.MockBidCommandPublisher
	loggerMock              *mocks.MockLogger
}

func (s *PlaceBidCommandTestSuite) SetupTest() {
	s.bidCommandPublisherMock = mocks.NewMockBidCommandPublisher(s.T())
	s.loggerMock = mocks.NewMockLogger(s.T())

	s.sut = command.NewPlaceBidCommand(
		s.bidCommandPublisherMock,
		s.loggerMock,
	)
}

func TestPlaceBidCommandSuite(t *testing.T) {
	suite.Run(t, new(PlaceBidCommandTestSuite))
}

func (s *PlaceBidCommandTestSuite) TestExecute_WithProvidedIdempotencyKey_PublishesAndReturnsAccepted() {
	// Arrange
	ctx := context.Background()
	input := command.PlaceBidCommandInput{
		AuctionID:      1,
		UserID:         100,
		AmountInCents:  5000,
		IdempotencyKey: "client-key-123",
	}

	var capturedCommand ports.BidCommand
	s.bidCommandPublisherMock.
		On("Publish", mock.Anything, mock.MatchedBy(func(cmd ports.BidCommand) bool {
			capturedCommand = cmd
			return true
		})).
		Return(ports.BidCommandAck{IdempotencyKey: "client-key-123"}, nil)

	// Act
	result, err := s.sut.Execute(ctx, input)

	// Assert
	s.Require().NoError(err)
	s.Equal("client-key-123", result.IdempotencyKey)
	s.Equal("accepted", result.Status)
	s.Equal(input.AuctionID, capturedCommand.AuctionID)
	s.Equal(input.UserID, capturedCommand.UserID)
	s.Equal(input.AmountInCents, capturedCommand.AmountInCents)
	s.Equal("client-key-123", capturedCommand.IdempotencyKey)
	s.False(capturedCommand.IssuedAt.IsZero())
}

func (s *PlaceBidCommandTestSuite) TestExecute_WithoutIdempotencyKey_GeneratesKey() {
	// Arrange
	ctx := context.Background()
	input := command.PlaceBidCommandInput{
		AuctionID:     2,
		UserID:        200,
		AmountInCents: 7500,
	}

	var capturedCommand ports.BidCommand
	s.bidCommandPublisherMock.
		On("Publish", mock.Anything, mock.MatchedBy(func(cmd ports.BidCommand) bool {
			capturedCommand = cmd
			return true
		})).
		Return(ports.BidCommandAck{IdempotencyKey: "generated"}, nil).
		Run(func(args mock.Arguments) {
			cmd := args.Get(1).(ports.BidCommand)
			s.Require().NotEmpty(cmd.IdempotencyKey)
		})

	// Act
	result, err := s.sut.Execute(ctx, input)

	// Assert
	s.Require().NoError(err)
	s.Equal("accepted", result.Status)
	s.NotEmpty(capturedCommand.IdempotencyKey)
}

func (s *PlaceBidCommandTestSuite) TestExecute_ZeroAmount_ReturnsError() {
	// Arrange
	ctx := context.Background()
	input := command.PlaceBidCommandInput{
		AuctionID:     3,
		UserID:        300,
		AmountInCents: 0,
	}

	// Act
	result, err := s.sut.Execute(ctx, input)

	// Assert
	s.Require().ErrorIs(err, errs.ErrFirstBidMustBePositive)
	s.Empty(result.IdempotencyKey)
}

func (s *PlaceBidCommandTestSuite) TestExecute_PublishFails_ReturnsError() {
	// Arrange
	ctx := context.Background()
	input := command.PlaceBidCommandInput{
		AuctionID:     4,
		UserID:        400,
		AmountInCents: 9000,
	}

	publishErr := errors.New("publish failed")
	s.bidCommandPublisherMock.On("Publish", mock.Anything, mock.Anything).
		Return(ports.BidCommandAck{}, publishErr)
	nopLogger := zerolog.Nop()
	s.loggerMock.On("Error").Return(nopLogger.Error())

	// Act
	result, err := s.sut.Execute(ctx, input)

	// Assert
	s.Require().ErrorIs(err, publishErr)
	s.Empty(result.IdempotencyKey)
}
