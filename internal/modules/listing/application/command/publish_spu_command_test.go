package command_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"auction/internal/modules/listing/application/command"
	"auction/internal/modules/listing/domain/enum"
	"auction/internal/modules/listing/domain/errs"
	"auction/internal/modules/listing/domain/event"
	"auction/internal/modules/listing/domain/model"
	"auction/internal/modules/listing/ports"
	"auction/tests/mocks"
)

type PublishSpuCommandTestSuite struct {
	suite.Suite
	sut               *command.PublishSpuCommand
	uowFactoryMock    *mocks.MockListingUnitOfWorkFactory
	uowMock           *mocks.MockListingUnitOfWork
	spuRepositoryMock *mocks.MockSpuRepository
	outboxRepoMock    *mocks.MockListingOutboxRepository
	loggerMock        *mocks.MockLogger
	mockDraftSpu      model.SpuModel
}

func (s *PublishSpuCommandTestSuite) SetupTest() {
	s.uowFactoryMock = mocks.NewMockListingUnitOfWorkFactory(s.T())
	s.uowMock = mocks.NewMockListingUnitOfWork(s.T())
	s.spuRepositoryMock = mocks.NewMockSpuRepository(s.T())
	s.outboxRepoMock = mocks.NewMockListingOutboxRepository(s.T())
	s.loggerMock = mocks.NewMockLogger(s.T())

	s.sut = command.NewPublishSpuCommand(
		s.uowFactoryMock,
		s.loggerMock,
	)

	now := time.Now().UTC()
	draftStatus, _ := enum.NewListingStatusEnum(enum.EnumListingStatusDraft)
	s.mockDraftSpu, _ = model.RestoreSpuModel(1, "iPhone 15", "", 1, nil, nil, draftStatus, 1, now, now)
}

func TestPublishSpuCommandSuite(t *testing.T) {
	suite.Run(t, new(PublishSpuCommandTestSuite))
}

func (s *PublishSpuCommandTestSuite) TestExecute_DraftSpu_PublishesAndSavesOutboxEvent() {
	// Arrange
	ctx := context.Background()
	input := command.PublishSpuCommandInput{ID: 1}

	s.uowFactoryMock.On("Begin", mock.Anything).Return(s.uowMock, nil)
	s.uowMock.On("SpuRepository").Return(s.spuRepositoryMock)
	s.uowMock.On("OutboxRepository").Return(s.outboxRepoMock)
	s.uowMock.On("Complete", mock.Anything).Return(nil)
	s.uowMock.On("Rollback", mock.Anything).Return(nil)
	s.spuRepositoryMock.On("FindByIDForUpdate", mock.Anything, uint64(1)).Return(s.mockDraftSpu, nil)
	s.spuRepositoryMock.On("Update", mock.Anything, mock.AnythingOfType("model.SpuModel")).Return(nil)
	s.outboxRepoMock.
		On("Save", mock.Anything, mock.MatchedBy(func(evt ports.OutboxEvent) bool {
			return evt.EventType == event.SpuPublishedEventType && evt.Subject == "listing.evt.spu.1"
		})).
		Return(nil)

	// Act
	output, err := s.sut.Execute(ctx, input)

	// Assert
	s.Require().NoError(err)
	s.Equal(uint64(1), output.ID)
	s.Equal(enum.EnumListingStatusPublished, output.Status)
}

func (s *PublishSpuCommandTestSuite) TestExecute_AlreadyPublished_ReturnsError() {
	// Arrange
	ctx := context.Background()
	input := command.PublishSpuCommandInput{ID: 1}

	now := time.Now().UTC()
	publishedStatus, _ := enum.NewListingStatusEnum(enum.EnumListingStatusPublished)
	publishedSpu, _ := model.RestoreSpuModel(1, "iPhone 15", "", 1, nil, nil, publishedStatus, 1, now, now)

	s.uowFactoryMock.On("Begin", mock.Anything).Return(s.uowMock, nil)
	s.uowMock.On("SpuRepository").Return(s.spuRepositoryMock)
	s.uowMock.On("Rollback", mock.Anything).Return(nil)
	s.spuRepositoryMock.On("FindByIDForUpdate", mock.Anything, uint64(1)).Return(publishedSpu, nil)
	s.loggerMock.On("Error").Return(nil)

	// Act
	output, err := s.sut.Execute(ctx, input)

	// Assert
	s.Require().ErrorIs(err, errs.ErrSpuAlreadyPublished)
	s.Equal(command.PublishSpuCommandOutput{}, output)
	s.uowMock.AssertNotCalled(s.T(), "Complete", mock.Anything)
}

func (s *PublishSpuCommandTestSuite) TestExecute_OutboxSaveFails_RollsBack() {
	// Arrange
	ctx := context.Background()
	input := command.PublishSpuCommandInput{ID: 1}
	outboxErr := errors.New("outbox save failed")

	s.uowFactoryMock.On("Begin", mock.Anything).Return(s.uowMock, nil)
	s.uowMock.On("SpuRepository").Return(s.spuRepositoryMock)
	s.uowMock.On("OutboxRepository").Return(s.outboxRepoMock)
	s.uowMock.On("Rollback", mock.Anything).Return(nil)
	s.spuRepositoryMock.On("FindByIDForUpdate", mock.Anything, uint64(1)).Return(s.mockDraftSpu, nil)
	s.spuRepositoryMock.On("Update", mock.Anything, mock.AnythingOfType("model.SpuModel")).Return(nil)
	s.outboxRepoMock.On("Save", mock.Anything, mock.AnythingOfType("ports.OutboxEvent")).Return(outboxErr)
	s.loggerMock.On("Error").Return(nil)

	// Act
	output, err := s.sut.Execute(ctx, input)

	// Assert: transaction is rolled back, the event is never delivered
	s.Require().ErrorIs(err, outboxErr)
	s.Equal(command.PublishSpuCommandOutput{}, output)
	s.uowMock.AssertNotCalled(s.T(), "Complete", mock.Anything)
}

func (s *PublishSpuCommandTestSuite) TestExecute_BeginFails_ReturnsError() {
	// Arrange
	ctx := context.Background()
	input := command.PublishSpuCommandInput{ID: 1}
	beginErr := errors.New("begin failed")

	s.uowFactoryMock.On("Begin", mock.Anything).Return(nil, beginErr)
	s.loggerMock.On("Error").Return(nil)

	// Act
	output, err := s.sut.Execute(ctx, input)

	// Assert
	s.Require().ErrorIs(err, beginErr)
	s.Equal(command.PublishSpuCommandOutput{}, output)
}
