package service_test

import (
	"context"
	"testing"

	"auction/internal/modules/notification/application/command"
	"auction/internal/modules/notification/application/service"
	"auction/internal/modules/notification/domain/enum"
	"auction/internal/modules/notification/domain/model"
	"auction/tests/mocks"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

const (
	serviceUserID   = uint64(100)
	serviceAuctionID = uint64(9)
)

type NotificationApplicationServiceTestSuite struct {
	suite.Suite
	sut               *service.NotificationApplicationService
	uowFactoryMock    *mocks.MockNotificationUnitOfWorkFactory
	uowMock           *mocks.MockNotificationUnitOfWork
	notificationsMock *mocks.MockNotificationRepository
	outboxMock        *mocks.MockNotificationOutboxRepository
	preferencesMock   *mocks.MockPreferenceRepository
	auctionReadMock   *mocks.MockAuctionReadPort
	watchlistMock     *mocks.MockWatchlistRepository
	listingReadMock   *mocks.MockListingReadPort
	userEmailMock     *mocks.MockUserEmailResolver
	loggerMock        *mocks.MockLogger
}

func (s *NotificationApplicationServiceTestSuite) SetupTest() {
	s.uowFactoryMock = mocks.NewMockNotificationUnitOfWorkFactory(s.T())
	s.uowMock = mocks.NewMockNotificationUnitOfWork(s.T())
	s.notificationsMock = mocks.NewMockNotificationRepository(s.T())
	s.outboxMock = mocks.NewMockNotificationOutboxRepository(s.T())
	s.preferencesMock = mocks.NewMockPreferenceRepository(s.T())
	s.auctionReadMock = mocks.NewMockAuctionReadPort(s.T())
	s.watchlistMock = mocks.NewMockWatchlistRepository(s.T())
	s.listingReadMock = mocks.NewMockListingReadPort(s.T())
	s.userEmailMock = mocks.NewMockUserEmailResolver(s.T())
	s.loggerMock = mocks.NewMockLogger(s.T())
	nopLogger := zerolog.Nop()
	s.loggerMock.On("Info").Return(nopLogger.Info()).Maybe()
	s.loggerMock.On("Error").Return(nopLogger.Error()).Maybe()

	createCommand := command.NewCreateNotificationCommand(s.uowFactoryMock, s.loggerMock)
	emailCommand := command.NewCreateEmailRequestCommand(s.outboxMock, s.loggerMock)
	s.sut = service.NewNotificationApplicationService(
		createCommand,
		emailCommand,
		s.preferencesMock,
		s.auctionReadMock,
		s.watchlistMock,
		s.listingReadMock,
		s.userEmailMock,
		s.loggerMock,
	)
}

func TestNotificationApplicationServiceSuite(t *testing.T) {
	suite.Run(t, new(NotificationApplicationServiceTestSuite))
}

func (s *NotificationApplicationServiceTestSuite) preferenceWith(inAppEnabled bool) model.NotificationPreferenceModel {
	settings := model.PreferenceSettings{
		enum.EnumNotificationCategoryPayment: {enum.EnumNotificationChannelInApp: inAppEnabled},
		enum.EnumNotificationCategoryAuction: {enum.EnumNotificationChannelInApp: inAppEnabled},
	}
	preference, err := model.NewNotificationPreference(serviceUserID, settings)
	s.Require().NoError(err)

	return preference
}

func (s *NotificationApplicationServiceTestSuite) expectNotificationPersisted() {
	s.uowFactoryMock.EXPECT().Begin(mock.Anything).Return(s.uowMock, nil)
	s.uowMock.EXPECT().NotificationRepository().Return(s.notificationsMock)
	s.notificationsMock.EXPECT().Save(mock.Anything, mock.Anything).
		Return(model.NotificationModel{}, false, nil)
	s.uowMock.EXPECT().Rollback(mock.Anything).Return(nil).Maybe()
}

func (s *NotificationApplicationServiceTestSuite) expectEmailEnqueued(to string) {
	s.userEmailMock.EXPECT().ResolveEmail(mock.Anything, serviceUserID).Return(to, true, nil)
	s.outboxMock.EXPECT().SaveEmailRequest(mock.Anything, mock.Anything).Return(nil)
}

func (s *NotificationApplicationServiceTestSuite) preferenceEmail(enabled bool) model.NotificationPreferenceModel {
	settings := model.PreferenceSettings{
		enum.EnumNotificationCategoryPayment: {
			enum.EnumNotificationChannelInApp: true,
			enum.EnumNotificationChannelEmail: enabled,
		},
	}
	preference, err := model.NewNotificationPreference(serviceUserID, settings)
	s.Require().NoError(err)

	return preference
}

func (s *NotificationApplicationServiceTestSuite) TestHandleRechargeSuccess_PreferenceEnabled_Persists() {
	// Arrange
	ctx := context.Background()
	s.preferencesMock.EXPECT().Get(ctx, serviceUserID).Return(s.preferenceWith(true), nil)
	s.expectNotificationPersisted()

	// Act
	err := s.sut.HandleRechargeSuccess(ctx, service.RechargeSuccessInput{
		SourceEventID: "evt-1",
		UserID:        serviceUserID,
		PaymentID:     7,
		AmountInCents: 9900,
		Currency:      "CNY",
	})

	// Assert
	s.Require().NoError(err)
}

func (s *NotificationApplicationServiceTestSuite) TestHandleRechargeSuccess_PreferenceDisabled_Skips() {
	// Arrange
	ctx := context.Background()
	s.preferencesMock.EXPECT().Get(ctx, serviceUserID).Return(s.preferenceWith(false), nil)

	// Act
	err := s.sut.HandleRechargeSuccess(ctx, service.RechargeSuccessInput{
		SourceEventID: "evt-1",
		UserID:        serviceUserID,
		AmountInCents: 9900,
		Currency:      "CNY",
	})

	// Assert
	s.Require().NoError(err)
	s.uowFactoryMock.AssertNotCalled(s.T(), "Begin", mock.Anything)
}

func (s *NotificationApplicationServiceTestSuite) TestHandleWithdrawalCompleted_PreferenceEnabled_Persists() {
	// Arrange
	ctx := context.Background()
	s.preferencesMock.EXPECT().Get(ctx, serviceUserID).Return(s.preferenceWith(true), nil)
	s.expectNotificationPersisted()

	// Act
	err := s.sut.HandleWithdrawalCompleted(ctx, service.WithdrawalCompletedInput{
		SourceEventID: "evt-w1",
		UserID:        serviceUserID,
		WithdrawalID:  55,
		AmountInCents: 30000,
		Currency:      "CNY",
	})

	// Assert
	s.Require().NoError(err)
}

func (s *NotificationApplicationServiceTestSuite) TestHandleWithdrawalCompleted_PreferenceDisabled_Skips() {
	// Arrange
	ctx := context.Background()
	s.preferencesMock.EXPECT().Get(ctx, serviceUserID).Return(s.preferenceWith(false), nil)

	// Act
	err := s.sut.HandleWithdrawalCompleted(ctx, service.WithdrawalCompletedInput{
		SourceEventID: "evt-w1",
		UserID:        serviceUserID,
		WithdrawalID:  55,
		AmountInCents: 30000,
		Currency:      "CNY",
	})

	// Assert
	s.Require().NoError(err)
	s.uowFactoryMock.AssertNotCalled(s.T(), "Begin", mock.Anything)
}

func (s *NotificationApplicationServiceTestSuite) TestHandleWithdrawalFailed_PreferenceEnabled_Persists() {
	// Arrange
	ctx := context.Background()
	s.preferencesMock.EXPECT().Get(ctx, serviceUserID).Return(s.preferenceWith(true), nil)
	s.expectNotificationPersisted()

	// Act
	err := s.sut.HandleWithdrawalFailed(ctx, service.WithdrawalFailedInput{
		SourceEventID: "evt-w2",
		UserID:        serviceUserID,
		WithdrawalID:  56,
		AmountInCents: 30000,
		Currency:      "CNY",
		FailReason:    "alipay rejected",
	})

	// Assert
	s.Require().NoError(err)
}

func (s *NotificationApplicationServiceTestSuite) TestHandleBidPlaced_NoPreviousBidder_Skips() {
	// Arrange
	ctx := context.Background()
	s.auctionReadMock.EXPECT().
		FindPreviousHighestBidderID(ctx, serviceAuctionID, serviceUserID).
		Return(uint64(0), false, nil)

	// Act
	err := s.sut.HandleBidPlaced(ctx, service.BidPlacedInput{
		SourceEventID: "evt-2",
		AuctionID:     serviceAuctionID,
		NewBidderID:   serviceUserID,
		AmountInCents: 12000,
	})

	// Assert
	s.Require().NoError(err)
	s.preferencesMock.AssertNotCalled(s.T(), "Get", mock.Anything, mock.Anything)
}

func (s *NotificationApplicationServiceTestSuite) TestHandleBidPlaced_PreviousBidder_NotifiesLoser() {
	// Arrange
	ctx := context.Background()
	previousBidderID := uint64(200)
	s.auctionReadMock.EXPECT().
		FindPreviousHighestBidderID(ctx, serviceAuctionID, serviceUserID).
		Return(previousBidderID, true, nil)
	preference, err := model.NewNotificationPreference(previousBidderID, model.PreferenceSettings{
		enum.EnumNotificationCategoryAuction: {enum.EnumNotificationChannelInApp: true},
	})
	s.Require().NoError(err)
	s.preferencesMock.EXPECT().Get(ctx, previousBidderID).Return(preference, nil)
	s.uowFactoryMock.EXPECT().Begin(mock.Anything).Return(s.uowMock, nil)
	s.uowMock.EXPECT().NotificationRepository().Return(s.notificationsMock)
	s.notificationsMock.EXPECT().Save(mock.Anything, mock.Anything).
		Return(model.NotificationModel{}, false, nil)
	s.uowMock.EXPECT().Rollback(mock.Anything).Return(nil).Maybe()

	// Act
	err = s.sut.HandleBidPlaced(ctx, service.BidPlacedInput{
		SourceEventID: "evt-2",
		AuctionID:     serviceAuctionID,
		NewBidderID:   serviceUserID,
		AmountInCents: 12000,
	})

	// Assert
	s.Require().NoError(err)
}

func (s *NotificationApplicationServiceTestSuite) TestHandleAuctionEnded_NoWinningBid_Skips() {
	// Arrange
	ctx := context.Background()

	// Act
	err := s.sut.HandleAuctionEnded(ctx, service.AuctionEndedInput{
		SourceEventID: "evt-3",
		AuctionID:     serviceAuctionID,
		WinningBidID:  nil,
	})

	// Assert
	s.Require().NoError(err)
	s.auctionReadMock.AssertNotCalled(s.T(), "FindBidderIDByBidID", mock.Anything, mock.Anything)
}

func (s *NotificationApplicationServiceTestSuite) TestHandleDepositEvent_UnknownEventType_Skips() {
	// Arrange
	ctx := context.Background()

	// Act
	err := s.sut.HandleDepositEvent(ctx, service.DepositEventInput{
		SourceEventID: "evt-4",
		EventType:     "deposit_unknown",
		UserID:        serviceUserID,
		AuctionID:     serviceAuctionID,
		AmountInCents: 5000,
		Currency:      "CNY",
	})

	// Assert
	s.Require().NoError(err)
	s.preferencesMock.AssertNotCalled(s.T(), "Get", mock.Anything, mock.Anything)
}

const (
	listingSpuID      = uint64(42)
	listingWatcherOne = uint64(300)
	listingWatcherTwo = uint64(301)
)

func (s *NotificationApplicationServiceTestSuite) preferenceListing(inAppEnabled bool) model.NotificationPreferenceModel {
	settings := model.PreferenceSettings{
		enum.EnumNotificationCategoryListing: {enum.EnumNotificationChannelInApp: inAppEnabled},
	}
	preference, err := model.NewNotificationPreference(listingWatcherOne, settings)
	s.Require().NoError(err)

	return preference
}

func (s *NotificationApplicationServiceTestSuite) TestHandleListingEvent_UnknownEventType_Skips() {
	// Arrange
	ctx := context.Background()

	// Act
	err := s.sut.HandleListingEvent(ctx, service.ListingEventInput{
		SourceEventID: "evt-l0",
		SpuID:         listingSpuID,
		EventType:     "listing.unknown",
	})

	// Assert
	s.Require().NoError(err)
	s.listingReadMock.AssertNotCalled(s.T(), "FindSpuTitleByID", mock.Anything, mock.Anything)
	s.watchlistMock.AssertNotCalled(s.T(), "FindWatcherIDsBySpuID", mock.Anything, mock.Anything)
}

func (s *NotificationApplicationServiceTestSuite) TestHandleListingEvent_NoWatchers_Skips() {
	// Arrange
	ctx := context.Background()
	s.listingReadMock.EXPECT().FindSpuTitleByID(ctx, listingSpuID).Return("Vintage Watch", true, nil)
	s.watchlistMock.EXPECT().FindWatcherIDsBySpuID(ctx, listingSpuID).Return([]uint64{}, nil)

	// Act
	err := s.sut.HandleListingEvent(ctx, service.ListingEventInput{
		SourceEventID: "evt-l1",
		SpuID:         listingSpuID,
		EventType:     "listing.spu.published",
	})

	// Assert
	s.Require().NoError(err)
	s.uowFactoryMock.AssertNotCalled(s.T(), "Begin", mock.Anything)
}

func (s *NotificationApplicationServiceTestSuite) TestHandleListingEvent_Published_NotifiesAllWatchers() {
	// Arrange
	ctx := context.Background()
	watchers := []uint64{listingWatcherOne, listingWatcherTwo}
	s.listingReadMock.EXPECT().FindSpuTitleByID(ctx, listingSpuID).Return("Vintage Watch", true, nil)
	s.watchlistMock.EXPECT().FindWatcherIDsBySpuID(ctx, listingSpuID).Return(watchers, nil)

	prefOne, err := model.NewNotificationPreference(listingWatcherOne, model.PreferenceSettings{
		enum.EnumNotificationCategoryListing: {enum.EnumNotificationChannelInApp: true},
	})
	s.Require().NoError(err)
	prefTwo, err := model.NewNotificationPreference(listingWatcherTwo, model.PreferenceSettings{
		enum.EnumNotificationCategoryListing: {enum.EnumNotificationChannelInApp: true},
	})
	s.Require().NoError(err)
	s.preferencesMock.EXPECT().Get(ctx, listingWatcherOne).Return(prefOne, nil)
	s.preferencesMock.EXPECT().Get(ctx, listingWatcherTwo).Return(prefTwo, nil)
	s.uowFactoryMock.EXPECT().Begin(mock.Anything).Return(s.uowMock, nil).Times(2)
	s.uowMock.EXPECT().NotificationRepository().Return(s.notificationsMock).Times(2)
	s.notificationsMock.EXPECT().Save(mock.Anything, mock.Anything).
		Return(model.NotificationModel{}, false, nil).Times(2)
	s.uowMock.EXPECT().Rollback(mock.Anything).Return(nil).Maybe()

	// Act
	err = s.sut.HandleListingEvent(ctx, service.ListingEventInput{
		SourceEventID: "evt-l2",
		SpuID:         listingSpuID,
		EventType:     "listing.spu.published",
	})

	// Assert
	s.Require().NoError(err)
}

func (s *NotificationApplicationServiceTestSuite) TestHandleListingEvent_SkuOffShelf_NotifiesWatchers() {
	// Arrange
	ctx := context.Background()
	s.listingReadMock.EXPECT().FindSpuTitleByID(ctx, listingSpuID).Return("Vintage Watch", true, nil)
	s.watchlistMock.EXPECT().FindWatcherIDsBySpuID(ctx, listingSpuID).Return([]uint64{listingWatcherOne}, nil)
	s.preferencesMock.EXPECT().Get(ctx, listingWatcherOne).Return(s.preferenceListing(true), nil)
	s.uowFactoryMock.EXPECT().Begin(mock.Anything).Return(s.uowMock, nil)
	s.uowMock.EXPECT().NotificationRepository().Return(s.notificationsMock)
	s.notificationsMock.EXPECT().Save(mock.Anything, mock.Anything).
		Return(model.NotificationModel{}, false, nil)
	s.uowMock.EXPECT().Rollback(mock.Anything).Return(nil).Maybe()

	// Act
	err := s.sut.HandleListingEvent(ctx, service.ListingEventInput{
		SourceEventID: "evt-l3",
		SpuID:         listingSpuID,
		EventType:     "listing.sku.off_shelf",
	})

	// Assert
	s.Require().NoError(err)
}

func (s *NotificationApplicationServiceTestSuite) TestHandleRechargeSuccess_EmailDisabled_SkipsEmail() {
	// Arrange
	ctx := context.Background()
	s.preferencesMock.EXPECT().Get(ctx, serviceUserID).Return(s.preferenceEmail(false), nil)
	s.expectNotificationPersisted()

	// Act
	err := s.sut.HandleRechargeSuccess(ctx, service.RechargeSuccessInput{
		SourceEventID: "evt-email-off",
		UserID:        serviceUserID,
		PaymentID:     11,
		AmountInCents: 5000,
		Currency:      "CNY",
	})

	// Assert
	s.Require().NoError(err)
	s.userEmailMock.AssertNotCalled(s.T(), "ResolveEmail", mock.Anything, mock.Anything)
	s.outboxMock.AssertNotCalled(s.T(), "SaveEmailRequest", mock.Anything, mock.Anything)
}

func (s *NotificationApplicationServiceTestSuite) TestHandleRechargeSuccess_EmailEnabled_EnqueuesEmail() {
	// Arrange
	ctx := context.Background()
	s.preferencesMock.EXPECT().Get(ctx, serviceUserID).Return(s.preferenceEmail(true), nil)
	s.expectNotificationPersisted()
	s.expectEmailEnqueued("buyer@example.com")

	// Act
	err := s.sut.HandleRechargeSuccess(ctx, service.RechargeSuccessInput{
		SourceEventID: "evt-email-on",
		UserID:        serviceUserID,
		PaymentID:     14,
		AmountInCents: 5000,
		Currency:      "CNY",
	})

	// Assert
	s.Require().NoError(err)
	s.outboxMock.AssertExpectations(s.T())
}

func (s *NotificationApplicationServiceTestSuite) TestHandleRechargeSuccess_EmailResolutionFails_PersistsInApp() {
	// Arrange
	ctx := context.Background()
	s.preferencesMock.EXPECT().Get(ctx, serviceUserID).Return(s.preferenceEmail(true), nil)
	s.expectNotificationPersisted()
	s.userEmailMock.EXPECT().ResolveEmail(mock.Anything, serviceUserID).
		Return("", false, context.DeadlineExceeded)

	// Act
	err := s.sut.HandleRechargeSuccess(ctx, service.RechargeSuccessInput{
		SourceEventID: "evt-email-err",
		UserID:        serviceUserID,
		PaymentID:     12,
		AmountInCents: 5000,
		Currency:      "CNY",
	})

	// Assert
	s.Require().NoError(err)
	s.outboxMock.AssertNotCalled(s.T(), "SaveEmailRequest", mock.Anything, mock.Anything)
}

func (s *NotificationApplicationServiceTestSuite) TestHandleRechargeSuccess_UnknownEmail_SkipsEmail() {
	// Arrange
	ctx := context.Background()
	s.preferencesMock.EXPECT().Get(ctx, serviceUserID).Return(s.preferenceEmail(true), nil)
	s.expectNotificationPersisted()
	s.userEmailMock.EXPECT().ResolveEmail(mock.Anything, serviceUserID).Return("", false, nil)

	// Act
	err := s.sut.HandleRechargeSuccess(ctx, service.RechargeSuccessInput{
		SourceEventID: "evt-email-missing",
		UserID:        serviceUserID,
		PaymentID:     13,
		AmountInCents: 5000,
		Currency:      "CNY",
	})

	// Assert
	s.Require().NoError(err)
	s.outboxMock.AssertNotCalled(s.T(), "SaveEmailRequest", mock.Anything, mock.Anything)
}
