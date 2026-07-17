package command_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"auction/internal/modules/listing/application/command"
	"auction/internal/modules/listing/domain/errs"
	"auction/internal/modules/listing/domain/model"
	"auction/tests/mocks"
)

type UpdateCategoryCommandTestSuite struct {
	suite.Suite
	sut                    *command.UpdateCategoryCommand
	categoryRepositoryMock *mocks.MockCategoryRepository
	loggerMock             *mocks.MockLogger
	mockCategory           model.CategoryModel
}

func (s *UpdateCategoryCommandTestSuite) SetupTest() {
	s.categoryRepositoryMock = mocks.NewMockCategoryRepository(s.T())
	s.loggerMock = mocks.NewMockLogger(s.T())

	s.sut = command.NewUpdateCategoryCommand(
		s.categoryRepositoryMock,
		s.loggerMock,
	)

	now := time.Now().UTC()
	s.mockCategory, _ = model.RestoreCategoryModel(1, "数码", nil, 0, 1, now, now)
}

func TestUpdateCategoryCommandSuite(t *testing.T) {
	suite.Run(t, new(UpdateCategoryCommandTestSuite))
}

func (s *UpdateCategoryCommandTestSuite) TestExecute_ValidInput_ReturnsUpdatedCategory() {
	// Arrange
	ctx := context.Background()
	input := command.UpdateCategoryCommandInput{ID: 1, Name: "电子产品", SortOrder: 5}

	s.categoryRepositoryMock.On("FindByID", mock.Anything, uint64(1)).Return(s.mockCategory, nil)
	s.categoryRepositoryMock.
		On("Update", mock.Anything, mock.AnythingOfType("model.CategoryModel")).
		Return(nil)

	// Act
	output, err := s.sut.Execute(ctx, input)

	// Assert
	s.Require().NoError(err)
	s.Equal("电子产品", output.Name)
	s.Equal(int32(5), output.SortOrder)
}

func (s *UpdateCategoryCommandTestSuite) TestExecute_CategoryNotFound_ReturnsError() {
	// Arrange
	ctx := context.Background()
	input := command.UpdateCategoryCommandInput{ID: 99, Name: "电子产品"}

	s.categoryRepositoryMock.
		On("FindByID", mock.Anything, uint64(99)).
		Return(model.CategoryModel{}, errs.ErrCategoryNotFound)
	s.loggerMock.On("Error").Return(nil)

	// Act
	output, err := s.sut.Execute(ctx, input)

	// Assert
	s.Require().ErrorIs(err, errs.ErrCategoryNotFound)
	s.Equal(command.UpdateCategoryCommandOutput{}, output)
}

func (s *UpdateCategoryCommandTestSuite) TestExecute_RepositoryError_ReturnsError() {
	// Arrange
	ctx := context.Background()
	input := command.UpdateCategoryCommandInput{ID: 1, Name: "电子产品"}
	repositoryErr := errors.New("repository error")

	s.categoryRepositoryMock.On("FindByID", mock.Anything, uint64(1)).Return(s.mockCategory, nil)
	s.categoryRepositoryMock.
		On("Update", mock.Anything, mock.AnythingOfType("model.CategoryModel")).
		Return(repositoryErr)
	s.loggerMock.On("Error").Return(nil)

	// Act
	output, err := s.sut.Execute(ctx, input)

	// Assert
	s.Require().ErrorIs(err, repositoryErr)
	s.Equal(command.UpdateCategoryCommandOutput{}, output)
}
