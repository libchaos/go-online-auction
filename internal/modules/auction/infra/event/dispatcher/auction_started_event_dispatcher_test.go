package dispatcher_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/cristiano-pacheco/go-online-auction/internal/modules/auction/domain/event"
	"github.com/cristiano-pacheco/go-online-auction/internal/modules/auction/infra/event/dispatcher"
	"github.com/cristiano-pacheco/go-online-auction/tests/mocks"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type RedisAuctionStartedEventDispatcherTestSuite struct {
	suite.Suite
	sut         *dispatcher.RedisAuctionStartedEventDispatcher
	redisClient *mocks.MockUniversalClient
	logger      *mocks.MockLogger
}

func (s *RedisAuctionStartedEventDispatcherTestSuite) SetupTest() {
	s.redisClient = mocks.NewMockUniversalClient(s.T())
	s.logger = mocks.NewMockLogger(s.T())
	s.sut = dispatcher.NewRedisAuctionStartedEventDispatcher(s.redisClient, s.logger)
}

func TestRedisAuctionStartedEventDispatcherSuite(t *testing.T) {
	suite.Run(t, new(RedisAuctionStartedEventDispatcherTestSuite))
}

func (s *RedisAuctionStartedEventDispatcherTestSuite) TestDispatch_Success() {
	// Arrange
	ctx := context.Background()
	auctionID := uint64(123)
	listingID := uint64(456)
	startTime := time.Now()
	endTime := startTime.Add(24 * time.Hour)
	evt := event.NewAuctionStartedEvent(auctionID, listingID, &startTime, endTime)
	channel := dispatcher.BuildAuctionEventChannel(auctionID)
	expectedPayloadData := dispatcher.AuctionStartedPayloadData{
		ListingID: listingID,
		StartTime: startTime,
		EndTime:   endTime,
	}
	expectedPayload := dispatcher.AuctionStartedPayload{
		EventType: event.AuctionStartedEventType,
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

func (s *RedisAuctionStartedEventDispatcherTestSuite) TestDispatch_WithNilStartTime_Success() {
	// Arrange
	ctx := context.Background()
	auctionID := uint64(789)
	listingID := uint64(987)
	endTime := time.Now().Add(48 * time.Hour)
	evt := event.NewAuctionStartedEvent(auctionID, listingID, nil, endTime)
	channel := dispatcher.BuildAuctionEventChannel(auctionID)
	expectedPayloadData := dispatcher.AuctionStartedPayloadData{
		ListingID: listingID,
		StartTime: time.Time{},
		EndTime:   endTime,
	}
	expectedPayload := dispatcher.AuctionStartedPayload{
		EventType: event.AuctionStartedEventType,
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

func (s *RedisAuctionStartedEventDispatcherTestSuite) TestDispatch_RedisPublishError_ReturnsError() {
	// Arrange
	ctx := context.Background()
	auctionID := uint64(321)
	listingID := uint64(654)
	startTime := time.Now()
	endTime := startTime.Add(12 * time.Hour)
	evt := event.NewAuctionStartedEvent(auctionID, listingID, &startTime, endTime)
	channel := dispatcher.BuildAuctionEventChannel(auctionID)
	publishError := errors.New("redis publish failed")
	expectedPayloadData := dispatcher.AuctionStartedPayloadData{
		ListingID: listingID,
		StartTime: startTime,
		EndTime:   endTime,
	}
	expectedPayload := dispatcher.AuctionStartedPayload{
		EventType: event.AuctionStartedEventType,
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

func (s *RedisAuctionStartedEventDispatcherTestSuite) TestDispatch_ValidatesPayloadStructure() {
	// Arrange
	ctx := context.Background()
	auctionID := uint64(999)
	listingID := uint64(888)
	startTime := time.Now()
	endTime := startTime.Add(36 * time.Hour)
	evt := event.NewAuctionStartedEvent(auctionID, listingID, &startTime, endTime)
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
	var payload dispatcher.AuctionStartedPayload
	unmarshalErr := json.Unmarshal(capturedPayload, &payload)
	s.Require().NoError(unmarshalErr)
	s.Equal(event.AuctionStartedEventType, payload.EventType)
	s.Equal(evt.EventID(), payload.EventID)
	s.Equal(auctionID, payload.AuctionID)
	s.Equal(listingID, payload.Data.ListingID)
	s.WithinDuration(startTime, payload.Data.StartTime, 1*time.Second)
	s.WithinDuration(endTime, payload.Data.EndTime, 1*time.Second)
	s.WithinDuration(time.Now(), payload.Timestamp, 5*time.Second)
}
