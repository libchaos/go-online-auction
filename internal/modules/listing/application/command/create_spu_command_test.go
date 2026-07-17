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

type CreateSpuCommandTestSuite struct {
	suite.Suite
	sut                    *command.CreateSpuCommand
	spuRepositoryMock      *mocks.MockSpuRepository
	categoryRepositoryMock *mocks.MockCategoryRepository
	loggerMock             *mocks.MockLogger
	mockCategory           model.CategoryModel
	mockPersistedSpu       model.SpuModel
}

func (s *CreateSpuCommandTestSuite) SetupTest() {
	s.spuRepositoryMock = mocks.NewMockSpuRepository(s.T())
	s.categoryRepositoryMock = mocks.NewMockCategoryRepository(s.T())
	s.loggerMock = mocks.NewMockLogger(s.T())

	s.sut = command.NewCreateSpuCommand(
		s.spuRepositoryMock,
		s.categoryRepositoryMock,
		s.loggerMock,
	)

	now := time.Now().UTC()
	s.mockCategory, _ = model.RestoreCategoryModel(1, "数码", nil, 0, 1, now, now)

	draftStatus, _ := enum.NewListingStatusEnum(enum.EnumListingStatusDraft)
	s.mockPersistedSpu, _ = model.RestoreSpuModel(
		1, "iPhone 15", "旗舰手机", 1, nil, nil, draftStatus, 1, now, now,
	)
}

func TestCreateSpuCommandSuite(t *testing.T) {
	suite.Run(t, new(CreateSpuCommandTestSuite))
}

func (s *CreateSpuCommandTestSuite) TestExecute_ValidInput_ReturnsCreatedSpu() {
	// Arrange
	ctx := context.Background()
	input := command.CreateSpuCommandInput{
		Title:       "iPhone 15",
		Description: "旗舰手机",
		CategoryID:  1,
	}

	s.categoryRepositoryMock.On("FindByID", mock.Anything, uint64(1)).Return(s.mockCategory, nil)
	s.spuRepositoryMock.
		On("Create", mock.Anything, mock.AnythingOfType("model.SpuModel")).
		Return(s.mockPersistedSpu, nil)

	// Act
	output, err := s.sut.Execute(ctx, input)

	// Assert
	s.Require().NoError(err)
	s.Equal(uint64(1), output.ID)
	s.Equal("iPhone 15", output.Title)
	s.Equal(enum.EnumListingStatusDraft, output.Status)
}

func (s *CreateSpuCommandTestSuite) TestExecute_CategoryNotFound_ReturnsError() {
	// Arrange
	ctx := context.Background()
	input := command.CreateSpuCommandInput{Title: "iPhone 15", CategoryID: 99}

	s.categoryRepositoryMock.
		On("FindByID", mock.Anything, uint64(99)).
		Return(model.CategoryModel{}, errs.ErrCategoryNotFound)

	// Act
	output, err := s.sut.Execute(ctx, input)

	// Assert
	s.Require().ErrorIs(err, errs.ErrCategoryNotFound)
	s.Equal(command.CreateSpuCommandOutput{}, output)
	s.spuRepositoryMock.AssertNotCalled(s.T(), "Create")
}

func (s *CreateSpuCommandTestSuite) TestExecute_ZeroCategoryID_ReturnsError() {
	// Arrange
	ctx := context.Background()
	input := command.CreateSpuCommandInput{Title: "iPhone 15", CategoryID: 0}

	// Act
	output, err := s.sut.Execute(ctx, input)

	// Assert
	s.Require().ErrorIs(err, errs.ErrSpuCategoryRequired)
	s.Equal(command.CreateSpuCommandOutput{}, output)
	s.spuRepositoryMock.AssertNotCalled(s.T(), "Create")
}

func (s *CreateSpuCommandTestSuite) TestExecute_RepositoryError_ReturnsError() {
	// Arrange
	ctx := context.Background()
	input := command.CreateSpuCommandInput{Title: "iPhone 15", CategoryID: 1}
	repositoryErr := errors.New("repository error")

	s.categoryRepositoryMock.On("FindByID", mock.Anything, uint64(1)).Return(s.mockCategory, nil)
	s.spuRepositoryMock.
		On("Create", mock.Anything, mock.AnythingOfType("model.SpuModel")).
		Return(model.SpuModel{}, repositoryErr)
	s.loggerMock.On("Error").Return(nil)

	// Act
	output, err := s.sut.Execute(ctx, input)

	// Assert
	s.Require().ErrorIs(err, repositoryErr)
	s.Equal(command.CreateSpuCommandOutput{}, output)
}
