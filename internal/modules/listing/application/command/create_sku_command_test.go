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
	"auction/internal/modules/listing/domain/model"
	"auction/tests/mocks"
)

type CreateSkuCommandTestSuite struct {
	suite.Suite
	sut               *command.CreateSkuCommand
	skuRepositoryMock *mocks.MockSkuRepository
	spuRepositoryMock *mocks.MockSpuRepository
	loggerMock        *mocks.MockLogger
	mockSpu           model.SpuModel
	mockPersistedSku  model.SkuModel
	validSpecValues   map[string]string
}

func (s *CreateSkuCommandTestSuite) SetupTest() {
	s.skuRepositoryMock = mocks.NewMockSkuRepository(s.T())
	s.spuRepositoryMock = mocks.NewMockSpuRepository(s.T())
	s.loggerMock = mocks.NewMockLogger(s.T())

	s.sut = command.NewCreateSkuCommand(
		s.skuRepositoryMock,
		s.spuRepositoryMock,
		s.loggerMock,
	)

	now := time.Now().UTC()
	draftStatus, _ := enum.NewListingStatusEnum(enum.EnumListingStatusDraft)
	s.mockSpu, _ = model.RestoreSpuModel(1, "iPhone 15", "", 1, nil, nil, draftStatus, 1, now, now)
	s.validSpecValues = map[string]string{"颜色": "红", "尺寸": "L"}
	s.mockPersistedSku, _ = model.RestoreSkuModel(
		10, 1, s.validSpecValues, 19900, 5, draftStatus, 1, now, now,
	)
}

func TestCreateSkuCommandSuite(t *testing.T) {
	suite.Run(t, new(CreateSkuCommandTestSuite))
}

func (s *CreateSkuCommandTestSuite) TestExecute_ValidInput_ReturnsCreatedSku() {
	// Arrange
	ctx := context.Background()
	input := command.CreateSkuCommandInput{
		SpuID:        1,
		SpecValues:   s.validSpecValues,
		PriceInCents: 19900,
		Quantity:     5,
	}

	s.spuRepositoryMock.On("FindByID", mock.Anything, uint64(1)).Return(s.mockSpu, nil)
	s.skuRepositoryMock.
		On("Create", mock.Anything, mock.AnythingOfType("model.SkuModel")).
		Return(s.mockPersistedSku, nil)

	// Act
	output, err := s.sut.Execute(ctx, input)

	// Assert
	s.Require().NoError(err)
	s.Equal(uint64(10), output.ID)
	s.Equal(uint64(1), output.SpuID)
	s.Equal(s.validSpecValues, output.SpecValues)
	s.Equal(enum.EnumListingStatusDraft, output.Status)
}

func (s *CreateSkuCommandTestSuite) TestExecute_SpuNotFound_ReturnsError() {
	// Arrange
	ctx := context.Background()
	input := command.CreateSkuCommandInput{
		SpuID:        99,
		SpecValues:   s.validSpecValues,
		PriceInCents: 19900,
	}

	s.spuRepositoryMock.
		On("FindByID", mock.Anything, uint64(99)).
		Return(model.SpuModel{}, errs.ErrSpuNotFound)
	s.loggerMock.On("Error").Return(nil)

	// Act
	output, err := s.sut.Execute(ctx, input)

	// Assert
	s.Require().ErrorIs(err, errs.ErrSpuNotFound)
	s.Equal(command.CreateSkuCommandOutput{}, output)
	s.skuRepositoryMock.AssertNotCalled(s.T(), "Create")
}

func (s *CreateSkuCommandTestSuite) TestExecute_EmptySpecValues_ReturnsError() {
	// Arrange
	ctx := context.Background()
	input := command.CreateSkuCommandInput{SpuID: 1, PriceInCents: 19900}

	s.spuRepositoryMock.On("FindByID", mock.Anything, uint64(1)).Return(s.mockSpu, nil)
	s.loggerMock.On("Error").Return(nil)

	// Act
	output, err := s.sut.Execute(ctx, input)

	// Assert
	s.Require().ErrorIs(err, errs.ErrSkuSpecValuesRequired)
	s.Equal(command.CreateSkuCommandOutput{}, output)
	s.skuRepositoryMock.AssertNotCalled(s.T(), "Create")
}

func (s *CreateSkuCommandTestSuite) TestExecute_RepositoryError_ReturnsError() {
	// Arrange
	ctx := context.Background()
	input := command.CreateSkuCommandInput{
		SpuID:        1,
		SpecValues:   s.validSpecValues,
		PriceInCents: 19900,
	}
	repositoryErr := errors.New("repository error")

	s.spuRepositoryMock.On("FindByID", mock.Anything, uint64(1)).Return(s.mockSpu, nil)
	s.skuRepositoryMock.
		On("Create", mock.Anything, mock.AnythingOfType("model.SkuModel")).
		Return(model.SkuModel{}, repositoryErr)
	s.loggerMock.On("Error").Return(nil)

	// Act
	output, err := s.sut.Execute(ctx, input)

	// Assert
	s.Require().ErrorIs(err, repositoryErr)
	s.Equal(command.CreateSkuCommandOutput{}, output)
}
