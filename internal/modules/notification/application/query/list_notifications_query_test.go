package query_test

import (
	"context"
	"testing"

	"auction/internal/modules/notification/application/query"
	"auction/internal/modules/notification/domain/model"
	"auction/tests/mocks"
	"github.com/stretchr/testify/suite"
)

const listQueryUserID = uint64(100)

type ListNotificationsQueryTestSuite struct {
	suite.Suite
	sut               *query.ListNotificationsQuery
	notificationsMock *mocks.MockNotificationRepository
}

func (s *ListNotificationsQueryTestSuite) SetupTest() {
	s.notificationsMock = mocks.NewMockNotificationRepository(s.T())
	s.sut = query.NewListNotificationsQuery(s.notificationsMock)
}

func TestListNotificationsQuerySuite(t *testing.T) {
	suite.Run(t, new(ListNotificationsQueryTestSuite))
}

func (s *ListNotificationsQueryTestSuite) TestExecute_DefaultsLimitWhenNonPositive() {
	// Arrange
	ctx := context.Background()
	s.notificationsMock.EXPECT().
		ListByUser(ctx, listQueryUserID, 20, 0).
		Return([]model.NotificationModel{}, nil)

	// Act
	_, err := s.sut.Execute(ctx, query.ListNotificationsQueryInput{UserID: listQueryUserID, Limit: 0, Offset: 0})

	// Assert
	s.Require().NoError(err)
}

func (s *ListNotificationsQueryTestSuite) TestExecute_ClampsLimitToMaximum() {
	// Arrange
	ctx := context.Background()
	s.notificationsMock.EXPECT().
		ListByUser(ctx, listQueryUserID, 100, 0).
		Return([]model.NotificationModel{}, nil)

	// Act
	_, err := s.sut.Execute(ctx, query.ListNotificationsQueryInput{UserID: listQueryUserID, Limit: 5000, Offset: 0})

	// Assert
	s.Require().NoError(err)
}

func (s *ListNotificationsQueryTestSuite) TestExecute_NormalizesNegativeOffset() {
	// Arrange
	ctx := context.Background()
	s.notificationsMock.EXPECT().
		ListByUser(ctx, listQueryUserID, 20, 0).
		Return([]model.NotificationModel{}, nil)

	// Act
	_, err := s.sut.Execute(ctx, query.ListNotificationsQueryInput{UserID: listQueryUserID, Limit: 0, Offset: -10})

	// Assert
	s.Require().NoError(err)
}

func (s *ListNotificationsQueryTestSuite) TestExecute_UnreadOnly_UsesUnreadRepository() {
	// Arrange
	ctx := context.Background()
	s.notificationsMock.EXPECT().
		ListUnreadByUser(ctx, listQueryUserID, 20, 0).
		Return([]model.NotificationModel{}, nil)

	// Act
	_, err := s.sut.Execute(ctx, query.ListNotificationsQueryInput{UserID: listQueryUserID, UnreadOnly: true})

	// Assert
	s.Require().NoError(err)
}
