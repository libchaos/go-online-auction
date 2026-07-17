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
	"auction/internal/modules/listing/domain/model"
	"auction/tests/mocks"
)

type UpdateSkuCommandTestSuite struct {
	suite.Suite
	sut               *command.UpdateSkuCommand
	skuRepositoryMock *mocks.MockSkuRepository
	loggerMock        *mocks.MockLogger
	mockDraftSku      model.SkuModel
	mockPublishedSku  model.SkuModel
	validSpecValues   map[string]string
}

func (s *UpdateSkuCommandTestSuite) SetupTest() {
	s.skuRepositoryMock = mocks.NewMockSkuRepository(s.T())
	s.loggerMock = mocks.NewMockLogger(s.T())

	s.sut = command.NewUpdateSkuCommand(
		s.skuRepositoryMock,
		s.loggerMock,
	)

	now := time.Now().UTC()
	s.validSpecValues = map[string]string{"颜色": "红"}
	draftStatus, _ := enum.NewListingStatusEnum(enum.EnumListingStatusDraft)
	publishedStatus, _ := enum.NewListingStatusEnum(enum.EnumListingStatusPublished)
	s.mockDraftSku, _ = model.RestoreSkuModel(10, 1, s.validSpecValues, 19900, 5, draftStatus, 1, now, now)
	s.mockPublishedSku, _ = model.RestoreSkuModel(10, 1, s.validSpecValues, 19900, 5, publishedStatus, 1, now, now)
}

func TestUpdateSkuCommandSuite(t *testing.T) {
	suite.Run(t, new(UpdateSkuCommandTestSuite))
}

func (s *UpdateSkuCommandTestSuite) TestExecute_ValidInput_ReturnsUpdatedSku() {
	// Arrange
	ctx := context.Background()
	input := command.UpdateSkuCommandInput{
		ID:           10,
		SpecValues:   map[string]string{"颜色": "蓝"},
		PriceInCents: 29900,
		Quantity:     10,
	}

	s.skuRepositoryMock.On("FindByID", mock.Anything, uint64(10)).Return(s.mockDraftSku, nil)
	s.skuRepositoryMock.On("Update", mock.Anything, mock.AnythingOfType("model.SkuModel")).Return(nil)

	// Act
	output, err := s.sut.Execute(ctx, input)

	// Assert
	s.Require().NoError(err)
	s.Equal(uint64(29900), output.PriceInCents)
	s.Equal(uint64(10), output.Quantity)
}

func (s *UpdateSkuCommandTestSuite) TestExecute_PublishedSku_ReturnsError() {
	// Arrange
	ctx := context.Background()
	input := command.UpdateSkuCommandInput{
		ID:           10,
		SpecValues:   s.validSpecValues,
		PriceInCents: 29900,
	}

	s.skuRepositoryMock.On("FindByID", mock.Anything, uint64(10)).Return(s.mockPublishedSku, nil)
	s.loggerMock.On("Error").Return(nil)

	// Act
	output, err := s.sut.Execute(ctx, input)

	// Assert
	s.Require().ErrorIs(err, errs.ErrSkuNotEditable)
	s.Equal(command.UpdateSkuCommandOutput{}, output)
	s.skuRepositoryMock.AssertNotCalled(s.T(), "Update")
}

func (s *UpdateSkuCommandTestSuite) TestExecute_SkuNotFound_ReturnsError() {
	// Arrange
	ctx := context.Background()
	input := command.UpdateSkuCommandInput{ID: 99, SpecValues: s.validSpecValues, PriceInCents: 100}

	s.skuRepositoryMock.
		On("FindByID", mock.Anything, uint64(99)).
		Return(model.SkuModel{}, errs.ErrSkuNotFound)
	s.loggerMock.On("Error").Return(nil)

	// Act
	output, err := s.sut.Execute(ctx, input)

	// Assert
	s.Require().ErrorIs(err, errs.ErrSkuNotFound)
	s.Equal(command.UpdateSkuCommandOutput{}, output)
}
