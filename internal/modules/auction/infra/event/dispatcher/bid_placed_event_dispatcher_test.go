package dispatcher_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/cristiano-pacheco/go-online-auction/internal/modules/auction/domain/event"
	"github.com/cristiano-pacheco/go-online-auction/internal/modules/auction/domain/model"
	"github.com/cristiano-pacheco/go-online-auction/internal/modules/auction/infra/event/dispatcher"
	"github.com/cristiano-pacheco/go-online-auction/tests/mocks"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type RedisBidPlacedEventDispatcherTestSuite struct {
	suite.Suite
	sut         *dispatcher.RedisBidPlacedEventDispatcher
	redisClient *mocks.MockUniversalClient
	logger      *mocks.MockLogger
}

func (s *RedisBidPlacedEventDispatcherTestSuite) SetupTest() {
	s.redisClient = mocks.NewMockUniversalClient(s.T())
	s.logger = mocks.NewMockLogger(s.T())
	s.sut = dispatcher.NewRedisBidPlacedEventDispatcher(s.redisClient, s.logger)
}

func TestRedisBidPlacedEventDispatcherSuite(t *testing.T) {
	suite.Run(t, new(RedisBidPlacedEventDispatcherTestSuite))
}

func (s *RedisBidPlacedEventDispatcherTestSuite) TestDispatch_Success() {
	// Arrange
	ctx := context.Background()
	auctionID := uint64(123)
	bidID := uint64(456)
	userID := uint64(789)
	amountInCents := uint64(10000)
	amount := model.NewMoneyModel(amountInCents)
	evt := event.NewBidPlacedEvent(bidID, auctionID, userID, amount)
	channel := dispatcher.BuildAuctionEventChannel(auctionID)
	expectedPayloadData := dispatcher.BidPlacedPayloadData{
		BidID:  bidID,
		UserID: userID,
		Amount: dispatcher.MoneyPayload{
			AmountInCents: amountInCents,
		},
	}
	expectedPayload := dispatcher.BidPlacedPayload{
		EventType: event.BidPlacedEventType,
		EventID:   evt.EventID(),
		Timestamp: evt.Timestamp(),
		AuctionID: auctionID,
		Data:      expectedPayloadData,
	}
	expectedJSON, marshalErr := json.Marshal(expectedPayload)
	s.Require().NoError(marshalErr)
	cmd := redis.NewIntCmd(ctx)
	cmd.SetVal(1)
	s.redisClient.On("Publish", mock.Anything, channel, expectedJSON).Return(cmd)
	nopLogger := zerolog.Nop()
	s.logger.On("Debug").Return(nopLogger.Debug())

	// Act
	err := s.sut.Dispatch(ctx, evt)

	// Assert
	s.Require().NoError(err)
}

func (s *RedisBidPlacedEventDispatcherTestSuite) TestDispatch_RedisPublishError_ReturnsError() {
	// Arrange
	ctx := context.Background()
	auctionID := uint64(321)
	bidID := uint64(654)
	userID := uint64(987)
	amountInCents := uint64(20000)
	amount := model.NewMoneyModel(amountInCents)
	evt := event.NewBidPlacedEvent(bidID, auctionID, userID, amount)
	channel := dispatcher.BuildAuctionEventChannel(auctionID)
	publishError := errors.New("redis publish failed")
	expectedPayloadData := dispatcher.BidPlacedPayloadData{
		BidID:  bidID,
		UserID: userID,
		Amount: dispatcher.MoneyPayload{
			AmountInCents: amountInCents,
		},
	}
	expectedPayload := dispatcher.BidPlacedPayload{
		EventType: event.BidPlacedEventType,
		EventID:   evt.EventID(),
		Timestamp: evt.Timestamp(),
		AuctionID: auctionID,
		Data:      expectedPayloadData,
	}
	expectedJSON, _ := json.Marshal(expectedPayload)
	cmd := redis.NewIntCmd(ctx)
	cmd.SetErr(publishError)
	s.redisClient.On("Publish", mock.Anything, channel, expectedJSON).Return(cmd)
	nopLogger := zerolog.Nop()
	s.logger.On("Error").Return(nopLogger.Error())

	// Act
	err := s.sut.Dispatch(ctx, evt)

	// Assert
	s.Require().ErrorIs(err, publishError)
}

func (s *RedisBidPlacedEventDispatcherTestSuite) TestDispatch_ValidatesPayloadStructure() {
	// Arrange
	ctx := context.Background()
	auctionID := uint64(999)
	bidID := uint64(888)
	userID := uint64(777)
	amountInCents := uint64(15000)
	amount := model.NewMoneyModel(amountInCents)
	evt := event.NewBidPlacedEvent(bidID, auctionID, userID, amount)
	channel := dispatcher.BuildAuctionEventChannel(auctionID)
	var capturedPayload []byte
	cmd := redis.NewIntCmd(ctx)
	cmd.SetVal(1)
	s.redisClient.
		On("Publish", mock.Anything, channel, mock.MatchedBy(func(data []byte) bool {
			capturedPayload = data
			return true
		})).
		Return(cmd)
	nopLogger := zerolog.Nop()
	s.logger.On("Debug").Return(nopLogger.Debug())

	// Act
	err := s.sut.Dispatch(ctx, evt)

	// Assert
	s.Require().NoError(err)
	var payload dispatcher.BidPlacedPayload
	unmarshalErr := json.Unmarshal(capturedPayload, &payload)
	s.Require().NoError(unmarshalErr)
	s.Equal(event.BidPlacedEventType, payload.EventType)
	s.Equal(evt.EventID(), payload.EventID)
	s.Equal(auctionID, payload.AuctionID)
	s.Equal(bidID, payload.Data.BidID)
	s.Equal(userID, payload.Data.UserID)
	s.Equal(amountInCents, payload.Data.Amount.AmountInCents)
	s.WithinDuration(time.Now(), payload.Timestamp, 5*time.Second)
}
