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

type PublishSkuCommandTestSuite struct {
	suite.Suite
	sut               *command.PublishSkuCommand
	uowFactoryMock    *mocks.MockListingUnitOfWorkFactory
	uowMock           *mocks.MockListingUnitOfWork
	spuRepositoryMock *mocks.MockSpuRepository
	skuRepositoryMock *mocks.MockSkuRepository
	outboxRepoMock    *mocks.MockListingOutboxRepository
	loggerMock        *mocks.MockLogger
	mockPublishedSpu  model.SpuModel
	mockDraftSku      model.SkuModel
	validSpecValues   map[string]string
}

func (s *PublishSkuCommandTestSuite) SetupTest() {
	s.uowFactoryMock = mocks.NewMockListingUnitOfWorkFactory(s.T())
	s.uowMock = mocks.NewMockListingUnitOfWork(s.T())
	s.spuRepositoryMock = mocks.NewMockSpuRepository(s.T())
	s.skuRepositoryMock = mocks.NewMockSkuRepository(s.T())
	s.outboxRepoMock = mocks.NewMockListingOutboxRepository(s.T())
	s.loggerMock = mocks.NewMockLogger(s.T())

	s.sut = command.NewPublishSkuCommand(
		s.uowFactoryMock,
		s.loggerMock,
	)

	now := time.Now().UTC()
	s.validSpecValues = map[string]string{"颜色": "红"}
	publishedStatus, _ := enum.NewListingStatusEnum(enum.EnumListingStatusPublished)
	draftStatus, _ := enum.NewListingStatusEnum(enum.EnumListingStatusDraft)
	s.mockPublishedSpu, _ = model.RestoreSpuModel(1, "iPhone 15", "", 1, nil, nil, publishedStatus, 1, now, now)
	s.mockDraftSku, _ = model.RestoreSkuModel(10, 1, s.validSpecValues, 19900, 5, draftStatus, 1, now, now)
}

func TestPublishSkuCommandSuite(t *testing.T) {
	suite.Run(t, new(PublishSkuCommandTestSuite))
}

func (s *PublishSkuCommandTestSuite) TestExecute_ValidInput_PublishesAndSavesOutboxEvent() {
	// Arrange
	ctx := context.Background()
	input := command.PublishSkuCommandInput{ID: 10}

	s.uowFactoryMock.On("Begin", mock.Anything).Return(s.uowMock, nil)
	s.uowMock.On("SpuRepository").Return(s.spuRepositoryMock)
	s.uowMock.On("SkuRepository").Return(s.skuRepositoryMock)
	s.uowMock.On("OutboxRepository").Return(s.outboxRepoMock)
	s.uowMock.On("Complete", mock.Anything).Return(nil)
	s.uowMock.On("Rollback", mock.Anything).Return(nil)
	s.skuRepositoryMock.On("FindByID", mock.Anything, uint64(10)).Return(s.mockDraftSku, nil)
	s.spuRepositoryMock.On("FindByIDForUpdate", mock.Anything, uint64(1)).Return(s.mockPublishedSpu, nil)
	s.skuRepositoryMock.On("FindByIDForUpdate", mock.Anything, uint64(10)).Return(s.mockDraftSku, nil)
	s.skuRepositoryMock.On("Update", mock.Anything, mock.AnythingOfType("model.SkuModel")).Return(nil)
	s.outboxRepoMock.
		On("Save", mock.Anything, mock.MatchedBy(func(evt ports.OutboxEvent) bool {
			return evt.EventType == event.SkuPublishedEventType && evt.Subject == "listing.evt.sku.10"
		})).
		Return(nil)

	// Act
	output, err := s.sut.Execute(ctx, input)

	// Assert
	s.Require().NoError(err)
	s.Equal(uint64(10), output.ID)
	s.Equal(enum.EnumListingStatusPublished, output.Status)
}

func (s *PublishSkuCommandTestSuite) TestExecute_SpuNotPublished_ReturnsError() {
	// Arrange
	ctx := context.Background()
	input := command.PublishSkuCommandInput{ID: 10}

	now := time.Now().UTC()
	draftStatus, _ := enum.NewListingStatusEnum(enum.EnumListingStatusDraft)
	draftSpu, _ := model.RestoreSpuModel(1, "iPhone 15", "", 1, nil, nil, draftStatus, 1, now, now)

	s.uowFactoryMock.On("Begin", mock.Anything).Return(s.uowMock, nil)
	s.uowMock.On("SpuRepository").Return(s.spuRepositoryMock)
	s.uowMock.On("SkuRepository").Return(s.skuRepositoryMock)
	s.uowMock.On("Rollback", mock.Anything).Return(nil)
	s.skuRepositoryMock.On("FindByID", mock.Anything, uint64(10)).Return(s.mockDraftSku, nil)
	s.spuRepositoryMock.On("FindByIDForUpdate", mock.Anything, uint64(1)).Return(draftSpu, nil)

	// Act
	output, err := s.sut.Execute(ctx, input)

	// Assert
	s.Require().ErrorIs(err, errs.ErrSpuMustBePublished)
	s.Equal(command.PublishSkuCommandOutput{}, output)
	s.skuRepositoryMock.AssertNotCalled(s.T(), "Update")
	s.uowMock.AssertNotCalled(s.T(), "Complete", mock.Anything)
}

func (s *PublishSkuCommandTestSuite) TestExecute_SkuNotFound_ReturnsError() {
	// Arrange
	ctx := context.Background()
	input := command.PublishSkuCommandInput{ID: 99}

	s.uowFactoryMock.On("Begin", mock.Anything).Return(s.uowMock, nil)
	s.uowMock.On("SkuRepository").Return(s.skuRepositoryMock)
	s.uowMock.On("Rollback", mock.Anything).Return(nil)
	s.skuRepositoryMock.
		On("FindByID", mock.Anything, uint64(99)).
		Return(model.SkuModel{}, errs.ErrSkuNotFound)
	s.loggerMock.On("Error").Return(nil)

	// Act
	output, err := s.sut.Execute(ctx, input)

	// Assert
	s.Require().ErrorIs(err, errs.ErrSkuNotFound)
	s.Equal(command.PublishSkuCommandOutput{}, output)
}

func (s *PublishSkuCommandTestSuite) TestExecute_OutboxSaveFails_RollsBack() {
	// Arrange
	ctx := context.Background()
	input := command.PublishSkuCommandInput{ID: 10}
	outboxErr := errors.New("outbox save failed")

	s.uowFactoryMock.On("Begin", mock.Anything).Return(s.uowMock, nil)
	s.uowMock.On("SpuRepository").Return(s.spuRepositoryMock)
	s.uowMock.On("SkuRepository").Return(s.skuRepositoryMock)
	s.uowMock.On("OutboxRepository").Return(s.outboxRepoMock)
	s.uowMock.On("Rollback", mock.Anything).Return(nil)
	s.skuRepositoryMock.On("FindByID", mock.Anything, uint64(10)).Return(s.mockDraftSku, nil)
	s.spuRepositoryMock.On("FindByIDForUpdate", mock.Anything, uint64(1)).Return(s.mockPublishedSpu, nil)
	s.skuRepositoryMock.On("FindByIDForUpdate", mock.Anything, uint64(10)).Return(s.mockDraftSku, nil)
	s.skuRepositoryMock.On("Update", mock.Anything, mock.AnythingOfType("model.SkuModel")).Return(nil)
	s.outboxRepoMock.On("Save", mock.Anything, mock.AnythingOfType("ports.OutboxEvent")).Return(outboxErr)
	s.loggerMock.On("Error").Return(nil)

	// Act
	output, err := s.sut.Execute(ctx, input)

	// Assert: transaction is rolled back, the event is never delivered
	s.Require().ErrorIs(err, outboxErr)
	s.Equal(command.PublishSkuCommandOutput{}, output)
	s.uowMock.AssertNotCalled(s.T(), "Complete", mock.Anything)
}
