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

type UpdateSpuCommandTestSuite struct {
	suite.Suite
	sut                    *command.UpdateSpuCommand
	spuRepositoryMock      *mocks.MockSpuRepository
	categoryRepositoryMock *mocks.MockCategoryRepository
	loggerMock             *mocks.MockLogger
	mockDraftSpu           model.SpuModel
	mockPublishedSpu       model.SpuModel
}

func (s *UpdateSpuCommandTestSuite) SetupTest() {
	s.spuRepositoryMock = mocks.NewMockSpuRepository(s.T())
	s.categoryRepositoryMock = mocks.NewMockCategoryRepository(s.T())
	s.loggerMock = mocks.NewMockLogger(s.T())

	s.sut = command.NewUpdateSpuCommand(
		s.spuRepositoryMock,
		s.categoryRepositoryMock,
		s.loggerMock,
	)

	now := time.Now().UTC()
	draftStatus, _ := enum.NewListingStatusEnum(enum.EnumListingStatusDraft)
	publishedStatus, _ := enum.NewListingStatusEnum(enum.EnumListingStatusPublished)
	s.mockDraftSpu, _ = model.RestoreSpuModel(1, "iPhone 15", "", 1, nil, nil, draftStatus, 1, now, now)
	s.mockPublishedSpu, _ = model.RestoreSpuModel(1, "iPhone 15", "", 1, nil, nil, publishedStatus, 1, now, now)
}

func TestUpdateSpuCommandSuite(t *testing.T) {
	suite.Run(t, new(UpdateSpuCommandTestSuite))
}

func (s *UpdateSpuCommandTestSuite) TestExecute_ValidInput_ReturnsUpdatedSpu() {
	// Arrange
	ctx := context.Background()
	input := command.UpdateSpuCommandInput{ID: 1, Title: "iPhone 15 Pro", CategoryID: 1}

	s.spuRepositoryMock.On("FindByID", mock.Anything, uint64(1)).Return(s.mockDraftSpu, nil)
	s.spuRepositoryMock.
		On("Update", mock.Anything, mock.AnythingOfType("model.SpuModel")).
		Return(nil)

	// Act
	output, err := s.sut.Execute(ctx, input)

	// Assert
	s.Require().NoError(err)
	s.Equal("iPhone 15 Pro", output.Title)
}

func (s *UpdateSpuCommandTestSuite) TestExecute_PublishedSpu_ReturnsError() {
	// Arrange
	ctx := context.Background()
	input := command.UpdateSpuCommandInput{ID: 1, Title: "iPhone 15 Pro", CategoryID: 1}

	s.spuRepositoryMock.On("FindByID", mock.Anything, uint64(1)).Return(s.mockPublishedSpu, nil)
	s.loggerMock.On("Error").Return(nil)

	// Act
	output, err := s.sut.Execute(ctx, input)

	// Assert
	s.Require().ErrorIs(err, errs.ErrSpuNotEditable)
	s.Equal(command.UpdateSpuCommandOutput{}, output)
	s.spuRepositoryMock.AssertNotCalled(s.T(), "Update")
}

func (s *UpdateSpuCommandTestSuite) TestExecute_CategoryChanged_ValidatesNewCategory() {
	// Arrange
	ctx := context.Background()
	input := command.UpdateSpuCommandInput{ID: 1, Title: "iPhone 15 Pro", CategoryID: 2}

	s.spuRepositoryMock.On("FindByID", mock.Anything, uint64(1)).Return(s.mockDraftSpu, nil)
	s.categoryRepositoryMock.
		On("FindByID", mock.Anything, uint64(2)).
		Return(model.CategoryModel{}, errs.ErrCategoryNotFound)
	s.loggerMock.On("Error").Return(nil)

	// Act
	output, err := s.sut.Execute(ctx, input)

	// Assert
	s.Require().ErrorIs(err, errs.ErrCategoryNotFound)
	s.Equal(command.UpdateSpuCommandOutput{}, output)
	s.spuRepositoryMock.AssertNotCalled(s.T(), "Update")
}

func (s *UpdateSpuCommandTestSuite) TestExecute_SpuNotFound_ReturnsError() {
	// Arrange
	ctx := context.Background()
	input := command.UpdateSpuCommandInput{ID: 99, Title: "iPhone 15 Pro", CategoryID: 1}

	s.spuRepositoryMock.
		On("FindByID", mock.Anything, uint64(99)).
		Return(model.SpuModel{}, errs.ErrSpuNotFound)
	s.loggerMock.On("Error").Return(nil)

	// Act
	output, err := s.sut.Execute(ctx, input)

	// Assert
	s.Require().ErrorIs(err, errs.ErrSpuNotFound)
	s.Equal(command.UpdateSpuCommandOutput{}, output)
}
