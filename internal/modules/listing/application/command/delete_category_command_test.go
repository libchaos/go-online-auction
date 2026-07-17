package command_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"auction/internal/modules/listing/application/command"
	"auction/internal/modules/listing/domain/errs"
	"auction/internal/modules/listing/domain/model"
	"auction/tests/mocks"
)

type DeleteCategoryCommandTestSuite struct {
	suite.Suite
	sut                    *command.DeleteCategoryCommand
	categoryRepositoryMock *mocks.MockCategoryRepository
	loggerMock             *mocks.MockLogger
	mockCategory           model.CategoryModel
}

func (s *DeleteCategoryCommandTestSuite) SetupTest() {
	s.categoryRepositoryMock = mocks.NewMockCategoryRepository(s.T())
	s.loggerMock = mocks.NewMockLogger(s.T())

	s.sut = command.NewDeleteCategoryCommand(
		s.categoryRepositoryMock,
		s.loggerMock,
	)

	now := time.Now().UTC()
	s.mockCategory, _ = model.RestoreCategoryModel(1, "数码", nil, 0, 1, now, now)
}

func TestDeleteCategoryCommandSuite(t *testing.T) {
	suite.Run(t, new(DeleteCategoryCommandTestSuite))
}

func (s *DeleteCategoryCommandTestSuite) TestExecute_ValidInput_DeletesCategory() {
	// Arrange
	ctx := context.Background()
	input := command.DeleteCategoryCommandInput{ID: 1}

	s.categoryRepositoryMock.On("FindByID", mock.Anything, uint64(1)).Return(s.mockCategory, nil)
	s.categoryRepositoryMock.On("CountChildren", mock.Anything, uint64(1)).Return(uint64(0), nil)
	s.categoryRepositoryMock.On("CountSpusByCategory", mock.Anything, uint64(1)).Return(uint64(0), nil)
	s.categoryRepositoryMock.On("Delete", mock.Anything, uint64(1)).Return(nil)

	// Act
	err := s.sut.Execute(ctx, input)

	// Assert
	s.Require().NoError(err)
}

func (s *DeleteCategoryCommandTestSuite) TestExecute_HasChildren_ReturnsError() {
	// Arrange
	ctx := context.Background()
	input := command.DeleteCategoryCommandInput{ID: 1}

	s.categoryRepositoryMock.On("FindByID", mock.Anything, uint64(1)).Return(s.mockCategory, nil)
	s.categoryRepositoryMock.On("CountChildren", mock.Anything, uint64(1)).Return(uint64(2), nil)

	// Act
	err := s.sut.Execute(ctx, input)

	// Assert
	s.Require().ErrorIs(err, errs.ErrCategoryHasChildren)
	s.categoryRepositoryMock.AssertNotCalled(s.T(), "Delete")
}

func (s *DeleteCategoryCommandTestSuite) TestExecute_CategoryInUse_ReturnsError() {
	// Arrange
	ctx := context.Background()
	input := command.DeleteCategoryCommandInput{ID: 1}

	s.categoryRepositoryMock.On("FindByID", mock.Anything, uint64(1)).Return(s.mockCategory, nil)
	s.categoryRepositoryMock.On("CountChildren", mock.Anything, uint64(1)).Return(uint64(0), nil)
	s.categoryRepositoryMock.On("CountSpusByCategory", mock.Anything, uint64(1)).Return(uint64(3), nil)

	// Act
	err := s.sut.Execute(ctx, input)

	// Assert
	s.Require().ErrorIs(err, errs.ErrCategoryInUse)
	s.categoryRepositoryMock.AssertNotCalled(s.T(), "Delete")
}

func (s *DeleteCategoryCommandTestSuite) TestExecute_CategoryNotFound_ReturnsError() {
	// Arrange
	ctx := context.Background()
	input := command.DeleteCategoryCommandInput{ID: 99}

	s.categoryRepositoryMock.
		On("FindByID", mock.Anything, uint64(99)).
		Return(model.CategoryModel{}, errs.ErrCategoryNotFound)
	s.loggerMock.On("Error").Return(nil)

	// Act
	err := s.sut.Execute(ctx, input)

	// Assert
	s.Require().ErrorIs(err, errs.ErrCategoryNotFound)
	s.categoryRepositoryMock.AssertNotCalled(s.T(), "Delete")
}
