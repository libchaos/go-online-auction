package command_test

import (
	"context"
	"testing"
	"time"

	"auction/internal/modules/notification/application/command"
	"auction/internal/modules/notification/domain/enum"
	"auction/internal/modules/notification/domain/model"
	"auction/tests/mocks"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

const (
	createUserID         = uint64(100)
	createIdempotencyKey = "evt-1:100:in_app"
)

type CreateNotificationCommandTestSuite struct {
	suite.Suite
	sut               *command.CreateNotificationCommand
	uowFactoryMock    *mocks.MockNotificationUnitOfWorkFactory
	uowMock           *mocks.MockNotificationUnitOfWork
	notificationsMock *mocks.MockNotificationRepository
	outboxMock        *mocks.MockNotificationOutboxRepository
	loggerMock        *mocks.MockLogger
}

func (s *CreateNotificationCommandTestSuite) SetupTest() {
	s.uowFactoryMock = mocks.NewMockNotificationUnitOfWorkFactory(s.T())
	s.uowMock = mocks.NewMockNotificationUnitOfWork(s.T())
	s.notificationsMock = mocks.NewMockNotificationRepository(s.T())
	s.outboxMock = mocks.NewMockNotificationOutboxRepository(s.T())
	s.loggerMock = mocks.NewMockLogger(s.T())
	nopLogger := zerolog.Nop()
	s.loggerMock.On("Info").Return(nopLogger.Info()).Maybe()
	s.loggerMock.On("Error").Return(nopLogger.Error()).Maybe()

	s.sut = command.NewCreateNotificationCommand(s.uowFactoryMock, s.loggerMock)
}

func TestCreateNotificationCommandSuite(t *testing.T) {
	suite.Run(t, new(CreateNotificationCommandTestSuite))
}

func (s *CreateNotificationCommandTestSuite) validInput() command.CreateNotificationCommandInput {
	return command.CreateNotificationCommandInput{
		UserID:         createUserID,
		Type:           enum.EnumNotificationTypeRechargeSuccess,
		Title:          "Recharge successful",
		Body:           "Your recharge was successful.",
		Payload:        []byte(`{"amount_in_cents":9900}`),
		Channels:       []string{enum.EnumNotificationChannelInApp},
		IdempotencyKey: createIdempotencyKey,
	}
}

func (s *CreateNotificationCommandTestSuite) persisted(id uint64, readAt *time.Time) model.NotificationModel {
	built, err := model.RestoreNotificationModel(
		id,
		createUserID,
		enum.EnumNotificationCategoryPayment,
		enum.EnumNotificationTypeRechargeSuccess,
		"Recharge successful",
		"Your recharge was successful.",
		[]byte(`{"amount_in_cents":9900}`),
		[]string{enum.EnumNotificationChannelInApp},
		createIdempotencyKey,
		readAt,
		time.Now().UTC(),
	)
	s.Require().NoError(err)

	return built
}

func (s *CreateNotificationCommandTestSuite) TestExecute_NewRow_WritesOutboxAndCommits() {
	// Arrange
	ctx := context.Background()
	s.uowFactoryMock.EXPECT().Begin(ctx).Return(s.uowMock, nil)
	s.uowMock.EXPECT().NotificationRepository().Return(s.notificationsMock)
	s.notificationsMock.EXPECT().Save(ctx, mock.Anything).Return(s.persisted(55, nil), true, nil)
	s.uowMock.EXPECT().OutboxRepository().Return(s.outboxMock)
	s.outboxMock.EXPECT().Save(ctx, mock.Anything).Return(nil)
	s.uowMock.EXPECT().Complete(ctx).Return(nil)
	s.uowMock.EXPECT().Rollback(ctx).Return(nil).Maybe()

	// Act
	output, err := s.sut.Execute(ctx, s.validInput())

	// Assert
	s.Require().NoError(err)
	s.True(output.Created)
	s.Equal(uint64(55), output.NotificationID)
}

func (s *CreateNotificationCommandTestSuite) TestExecute_DuplicateRow_SkipsOutboxAndDoesNotCommit() {
	// Arrange
	ctx := context.Background()
	s.uowFactoryMock.EXPECT().Begin(ctx).Return(s.uowMock, nil)
	s.uowMock.EXPECT().NotificationRepository().Return(s.notificationsMock)
	s.notificationsMock.EXPECT().Save(ctx, mock.Anything).Return(s.persisted(77, nil), false, nil)
	s.uowMock.EXPECT().Rollback(ctx).Return(nil).Maybe()

	// Act
	output, err := s.sut.Execute(ctx, s.validInput())

	// Assert
	s.Require().NoError(err)
	s.False(output.Created)
	s.Equal(uint64(77), output.NotificationID)
}

func (s *CreateNotificationCommandTestSuite) TestExecute_InvalidType_ReturnsErrorWithoutTransaction() {
	// Arrange
	ctx := context.Background()
	input := s.validInput()
	input.Type = "not_a_real_type"

	// Act
	output, err := s.sut.Execute(ctx, input)

	// Assert
	s.Require().Error(err)
	s.False(output.Created)
}
