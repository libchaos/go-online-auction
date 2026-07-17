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

type OffShelfSkuCommandTestSuite struct {
	suite.Suite
	sut               *command.OffShelfSkuCommand
	uowFactoryMock    *mocks.MockListingUnitOfWorkFactory
	uowMock           *mocks.MockListingUnitOfWork
	skuRepositoryMock *mocks.MockSkuRepository
	outboxRepoMock    *mocks.MockListingOutboxRepository
	loggerMock        *mocks.MockLogger
	mockPublishedSku  model.SkuModel
}

func (s *OffShelfSkuCommandTestSuite) SetupTest() {
	s.uowFactoryMock = mocks.NewMockListingUnitOfWorkFactory(s.T())
	s.uowMock = mocks.NewMockListingUnitOfWork(s.T())
	s.skuRepositoryMock = mocks.NewMockSkuRepository(s.T())
	s.outboxRepoMock = mocks.NewMockListingOutboxRepository(s.T())
	s.loggerMock = mocks.NewMockLogger(s.T())

	s.sut = command.NewOffShelfSkuCommand(
		s.uowFactoryMock,
		s.loggerMock,
	)

	now := time.Now().UTC()
	publishedStatus, _ := enum.NewListingStatusEnum(enum.EnumListingStatusPublished)
	specValues := map[string]string{"颜色": "红"}
	s.mockPublishedSku, _ = model.RestoreSkuModel(10, 1, specValues, 19900, 5, publishedStatus, 1, now, now)
}

func TestOffShelfSkuCommandSuite(t *testing.T) {
	suite.Run(t, new(OffShelfSkuCommandTestSuite))
}

func (s *OffShelfSkuCommandTestSuite) TestExecute_PublishedSku_GoesOffShelf() {
	// Arrange
	ctx := context.Background()
	input := command.OffShelfSkuCommandInput{ID: 10}

	s.uowFactoryMock.On("Begin", mock.Anything).Return(s.uowMock, nil)
	s.uowMock.On("SkuRepository").Return(s.skuRepositoryMock)
	s.uowMock.On("OutboxRepository").Return(s.outboxRepoMock)
	s.uowMock.On("Complete", mock.Anything).Return(nil)
	s.uowMock.On("Rollback", mock.Anything).Return(nil)
	s.skuRepositoryMock.On("FindByIDForUpdate", mock.Anything, uint64(10)).Return(s.mockPublishedSku, nil)
	s.skuRepositoryMock.On("Update", mock.Anything, mock.AnythingOfType("model.SkuModel")).Return(nil)
	s.outboxRepoMock.
		On("Save", mock.Anything, mock.MatchedBy(func(evt ports.OutboxEvent) bool {
			return evt.EventType == event.SkuOffShelfEventType && evt.Subject == "listing.evt.sku.10"
		})).
		Return(nil)

	// Act
	output, err := s.sut.Execute(ctx, input)

	// Assert
	s.Require().NoError(err)
	s.Equal(uint64(10), output.ID)
	s.Equal(enum.EnumListingStatusOffShelf, output.Status)
}

func (s *OffShelfSkuCommandTestSuite) TestExecute_DraftSku_ReturnsError() {
	// Arrange
	ctx := context.Background()
	input := command.OffShelfSkuCommandInput{ID: 10}

	now := time.Now().UTC()
	draftStatus, _ := enum.NewListingStatusEnum(enum.EnumListingStatusDraft)
	specValues := map[string]string{"颜色": "红"}
	draftSku, _ := model.RestoreSkuModel(10, 1, specValues, 19900, 5, draftStatus, 1, now, now)

	s.uowFactoryMock.On("Begin", mock.Anything).Return(s.uowMock, nil)
	s.uowMock.On("SkuRepository").Return(s.skuRepositoryMock)
	s.uowMock.On("Rollback", mock.Anything).Return(nil)
	s.skuRepositoryMock.On("FindByIDForUpdate", mock.Anything, uint64(10)).Return(draftSku, nil)
	s.loggerMock.On("Error").Return(nil)

	// Act
	output, err := s.sut.Execute(ctx, input)

	// Assert
	s.Require().ErrorIs(err, errs.ErrSkuNotPublished)
	s.Equal(command.OffShelfSkuCommandOutput{}, output)
	s.uowMock.AssertNotCalled(s.T(), "Complete", mock.Anything)
}
