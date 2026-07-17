package command_test

import (
	"context"
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

type OffShelfSpuCommandTestSuite struct {
	suite.Suite
	sut               *command.OffShelfSpuCommand
	uowFactoryMock    *mocks.MockListingUnitOfWorkFactory
	uowMock           *mocks.MockListingUnitOfWork
	spuRepositoryMock *mocks.MockSpuRepository
	skuRepositoryMock *mocks.MockSkuRepository
	outboxRepoMock    *mocks.MockListingOutboxRepository
	loggerMock        *mocks.MockLogger
	mockPublishedSpu  model.SpuModel
}

func (s *OffShelfSpuCommandTestSuite) SetupTest() {
	s.uowFactoryMock = mocks.NewMockListingUnitOfWorkFactory(s.T())
	s.uowMock = mocks.NewMockListingUnitOfWork(s.T())
	s.spuRepositoryMock = mocks.NewMockSpuRepository(s.T())
	s.skuRepositoryMock = mocks.NewMockSkuRepository(s.T())
	s.outboxRepoMock = mocks.NewMockListingOutboxRepository(s.T())
	s.loggerMock = mocks.NewMockLogger(s.T())

	s.sut = command.NewOffShelfSpuCommand(
		s.uowFactoryMock,
		s.loggerMock,
	)

	now := time.Now().UTC()
	publishedStatus, _ := enum.NewListingStatusEnum(enum.EnumListingStatusPublished)
	s.mockPublishedSpu, _ = model.RestoreSpuModel(1, "iPhone 15", "", 1, nil, nil, publishedStatus, 1, now, now)
}

func TestOffShelfSpuCommandSuite(t *testing.T) {
	suite.Run(t, new(OffShelfSpuCommandTestSuite))
}

func (s *OffShelfSpuCommandTestSuite) TestExecute_CascadesToPublishedSkus() {
	// Arrange
	ctx := context.Background()
	input := command.OffShelfSpuCommandInput{ID: 1}

	now := time.Now().UTC()
	publishedStatus, _ := enum.NewListingStatusEnum(enum.EnumListingStatusPublished)
	specValues := map[string]string{"颜色": "红"}
	sku1, _ := model.RestoreSkuModel(10, 1, specValues, 19900, 5, publishedStatus, 1, now, now)
	sku2, _ := model.RestoreSkuModel(11, 1, specValues, 29900, 3, publishedStatus, 1, now, now)

	s.uowFactoryMock.On("Begin", mock.Anything).Return(s.uowMock, nil)
	s.uowMock.On("SpuRepository").Return(s.spuRepositoryMock)
	s.uowMock.On("SkuRepository").Return(s.skuRepositoryMock)
	s.uowMock.On("OutboxRepository").Return(s.outboxRepoMock)
	s.uowMock.On("Complete", mock.Anything).Return(nil)
	s.uowMock.On("Rollback", mock.Anything).Return(nil)
	s.spuRepositoryMock.On("FindByIDForUpdate", mock.Anything, uint64(1)).Return(s.mockPublishedSpu, nil)
	s.spuRepositoryMock.On("Update", mock.Anything, mock.AnythingOfType("model.SpuModel")).Return(nil)
	s.skuRepositoryMock.
		On("FindPublishedBySpuIDForUpdate", mock.Anything, uint64(1)).
		Return([]model.SkuModel{sku1, sku2}, nil)
	s.skuRepositoryMock.On("Update", mock.Anything, mock.AnythingOfType("model.SkuModel")).Return(nil).Twice()
	s.outboxRepoMock.
		On("Save", mock.Anything, mock.MatchedBy(func(evt ports.OutboxEvent) bool {
			return evt.EventType == event.SkuOffShelfEventType
		})).
		Return(nil).Twice()
	s.outboxRepoMock.
		On("Save", mock.Anything, mock.MatchedBy(func(evt ports.OutboxEvent) bool {
			return evt.EventType == event.SpuOffShelfEventType && evt.Subject == "listing.evt.spu.1"
		})).
		Return(nil).Once()

	// Act
	output, err := s.sut.Execute(ctx, input)

	// Assert
	s.Require().NoError(err)
	s.Equal(uint64(1), output.ID)
	s.Equal(enum.EnumListingStatusOffShelf, output.Status)
	s.Equal(2, output.OffShelfSkuCount)
}

func (s *OffShelfSpuCommandTestSuite) TestExecute_NoPublishedSkus_EmitsOnlySpuEvent() {
	// Arrange
	ctx := context.Background()
	input := command.OffShelfSpuCommandInput{ID: 1}

	s.uowFactoryMock.On("Begin", mock.Anything).Return(s.uowMock, nil)
	s.uowMock.On("SpuRepository").Return(s.spuRepositoryMock)
	s.uowMock.On("SkuRepository").Return(s.skuRepositoryMock)
	s.uowMock.On("OutboxRepository").Return(s.outboxRepoMock)
	s.uowMock.On("Complete", mock.Anything).Return(nil)
	s.uowMock.On("Rollback", mock.Anything).Return(nil)
	s.spuRepositoryMock.On("FindByIDForUpdate", mock.Anything, uint64(1)).Return(s.mockPublishedSpu, nil)
	s.spuRepositoryMock.On("Update", mock.Anything, mock.AnythingOfType("model.SpuModel")).Return(nil)
	s.skuRepositoryMock.
		On("FindPublishedBySpuIDForUpdate", mock.Anything, uint64(1)).
		Return([]model.SkuModel{}, nil)
	s.outboxRepoMock.
		On("Save", mock.Anything, mock.MatchedBy(func(evt ports.OutboxEvent) bool {
			return evt.EventType == event.SpuOffShelfEventType
		})).
		Return(nil).Once()

	// Act
	output, err := s.sut.Execute(ctx, input)

	// Assert
	s.Require().NoError(err)
	s.Equal(0, output.OffShelfSkuCount)
}

func (s *OffShelfSpuCommandTestSuite) TestExecute_DraftSpu_ReturnsError() {
	// Arrange
	ctx := context.Background()
	input := command.OffShelfSpuCommandInput{ID: 1}

	now := time.Now().UTC()
	draftStatus, _ := enum.NewListingStatusEnum(enum.EnumListingStatusDraft)
	draftSpu, _ := model.RestoreSpuModel(1, "iPhone 15", "", 1, nil, nil, draftStatus, 1, now, now)

	s.uowFactoryMock.On("Begin", mock.Anything).Return(s.uowMock, nil)
	s.uowMock.On("SpuRepository").Return(s.spuRepositoryMock)
	s.uowMock.On("Rollback", mock.Anything).Return(nil)
	s.spuRepositoryMock.On("FindByIDForUpdate", mock.Anything, uint64(1)).Return(draftSpu, nil)
	s.loggerMock.On("Error").Return(nil)

	// Act
	output, err := s.sut.Execute(ctx, input)

	// Assert
	s.Require().ErrorIs(err, errs.ErrSpuNotPublished)
	s.Equal(command.OffShelfSpuCommandOutput{}, output)
	s.uowMock.AssertNotCalled(s.T(), "Complete", mock.Anything)
}
