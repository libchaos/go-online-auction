package dispatcher_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/cristiano-pacheco/go-online-auction/internal/modules/auction/domain/event"
	"github.com/cristiano-pacheco/go-online-auction/internal/modules/auction/domain/model"
	"github.com/cristiano-pacheco/go-online-auction/internal/modules/auction/infra/event/dispatcher"
	"github.com/cristiano-pacheco/go-online-auction/tests/mocks"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/suite"
)

type AuctionEndedEventDispatcherTestSuite struct {
	suite.Suite
	redisClient *mocks.MockUniversalClient
	logger      *mocks.MockLogger
	sut         *dispatcher.RedisAuctionEndedEventDispatcher
}

func (s *AuctionEndedEventDispatcherTestSuite) SetupTest() {
	s.redisClient = new(mocks.MockUniversalClient)
	s.logger = new(mocks.MockLogger)
	s.sut = dispatcher.NewRedisAuctionEndedEventDispatcher(s.redisClient, s.logger)
}

func TestAuctionEndedEventDispatcherTestSuite(t *testing.T) {
	suite.Run(t, new(AuctionEndedEventDispatcherTestSuite))
}

func (s *AuctionEndedEventDispatcherTestSuite) TestDispatch_Success_WithWinner() {
	// Arrange
	ctx := context.Background()
	auctionID := uint64(123)
	winningBidID := uint64(456)
	amountInCents := uint64(1000)
	finalAmount := model.NewMoneyModel(amountInCents)

	evt := event.NewAuctionEndedEvent(auctionID, &winningBidID, &finalAmount)
	channel := dispatcher.BuildAuctionEventChannel(auctionID)

	expectedPayloadData := dispatcher.AuctionEndedPayloadData{
		WinningBidID: &winningBidID,
		FinalAmount: &dispatcher.MoneyPayload{
			AmountInCents: amountInCents,
		},
	}
	expectedPayload := dispatcher.AuctionEndedPayload{
		EventType: event.AuctionEndedEventType,
		EventID:   evt.EventID(),
		Timestamp: evt.Timestamp(),
		AuctionID: auctionID,
		Data:      expectedPayloadData,
	}

	expectedJSON, err := json.Marshal(expectedPayload)
	s.Require().NoError(err)

	cmd := redis.NewIntCmd(ctx)
	cmd.SetVal(1)
	s.redisClient.On("Publish", ctx, channel, expectedJSON).Return(cmd)

	nopLogger := zerolog.Nop()
	s.logger.On("Debug").Return(nopLogger.Debug())

	// Act
	err = s.sut.Dispatch(ctx, evt)

	// Assert
	s.Require().NoError(err)
}

func (s *AuctionEndedEventDispatcherTestSuite) TestDispatch_Success_NoWinner() {
	// Arrange
	ctx := context.Background()
	auctionID := uint64(123)
	// No winner, no final amount

	evt := event.NewAuctionEndedEvent(auctionID, nil, nil)
	channel := dispatcher.BuildAuctionEventChannel(auctionID)

	expectedPayloadData := dispatcher.AuctionEndedPayloadData{
		WinningBidID: nil,
		FinalAmount:  nil,
	}
	expectedPayload := dispatcher.AuctionEndedPayload{
		EventType: event.AuctionEndedEventType,
		EventID:   evt.EventID(),
		Timestamp: evt.Timestamp(),
		AuctionID: auctionID,
		Data:      expectedPayloadData,
	}

	expectedJSON, err := json.Marshal(expectedPayload)
	s.Require().NoError(err)

	cmd := redis.NewIntCmd(ctx)
	cmd.SetVal(1)
	s.redisClient.On("Publish", ctx, channel, expectedJSON).Return(cmd)

	nopLogger := zerolog.Nop()
	s.logger.On("Debug").Return(nopLogger.Debug())

	// Act
	err = s.sut.Dispatch(ctx, evt)

	// Assert
	s.Require().NoError(err)
}

func (s *AuctionEndedEventDispatcherTestSuite) TestDispatch_Error_RedisPublish() {
	// Arrange
	ctx := context.Background()
	auctionID := uint64(123)
	evt := event.NewAuctionEndedEvent(auctionID, nil, nil)
	channel := dispatcher.BuildAuctionEventChannel(auctionID)

	expectedErr := errors.New("redis error")

	cmd := redis.NewIntCmd(ctx)
	cmd.SetErr(expectedErr)

	expectedPayloadData := dispatcher.AuctionEndedPayloadData{
		WinningBidID: nil,
		FinalAmount:  nil,
	}
	expectedPayload := dispatcher.AuctionEndedPayload{
		EventType: event.AuctionEndedEventType,
		EventID:   evt.EventID(),
		Timestamp: evt.Timestamp(),
		AuctionID: auctionID,
		Data:      expectedPayloadData,
	}
	expectedJSON, _ := json.Marshal(expectedPayload)

	s.redisClient.On("Publish", ctx, channel, expectedJSON).Return(cmd)

	nopLogger := zerolog.Nop()
	s.logger.On("Error").Return(nopLogger.Error())

	// Act
	err := s.sut.Dispatch(ctx, evt)

	// Assert
	s.Require().ErrorIs(err, expectedErr)
}
