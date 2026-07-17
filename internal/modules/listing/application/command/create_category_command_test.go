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

type CreateCategoryCommandTestSuite struct {
	suite.Suite
	sut                    *command.CreateCategoryCommand
	categoryRepositoryMock *mocks.MockCategoryRepository
	loggerMock             *mocks.MockLogger
	mockPersistedCategory  model.CategoryModel
	mockCreatedAt          time.Time
}

func (s *CreateCategoryCommandTestSuite) SetupTest() {
	s.categoryRepositoryMock = mocks.NewMockCategoryRepository(s.T())
	s.loggerMock = mocks.NewMockLogger(s.T())

	s.sut = command.NewCreateCategoryCommand(
		s.categoryRepositoryMock,
		s.loggerMock,
	)

	s.mockCreatedAt = time.Now().UTC()
	s.mockPersistedCategory, _ = model.RestoreCategoryModel(
		1, "数码", nil, 0, "", 0, 1, s.mockCreatedAt, s.mockCreatedAt,
	)
}

func TestCreateCategoryCommandSuite(t *testing.T) {
	suite.Run(t, new(CreateCategoryCommandTestSuite))
}

func (s *CreateCategoryCommandTestSuite) TestExecute_ValidInput_ReturnsCreatedCategory() {
	// Arrange
	ctx := context.Background()
	input := command.CreateCategoryCommandInput{Name: "数码"}

	s.categoryRepositoryMock.
		On("Create", mock.Anything, mock.AnythingOfType("model.CategoryModel")).
		Return(s.mockPersistedCategory, nil)

	// Act
	output, err := s.sut.Execute(ctx, input)

	// Assert
	s.Require().NoError(err)
	s.Equal(uint64(1), output.ID)
	s.Equal("数码", output.Name)
}

func (s *CreateCategoryCommandTestSuite) TestExecute_ValidParent_ValidatesParentExists() {
	// Arrange
	ctx := context.Background()
	parentID := uint64(1)
	input := command.CreateCategoryCommandInput{Name: "手机", ParentID: &parentID}

	childCategory, _ := model.RestoreCategoryModel(
		2, "手机", &parentID, 0, "", 0, 1, s.mockCreatedAt, s.mockCreatedAt,
	)

	s.categoryRepositoryMock.
		On("FindByID", mock.Anything, parentID).
		Return(s.mockPersistedCategory, nil)
	s.categoryRepositoryMock.
		On("Create", mock.Anything, mock.AnythingOfType("model.CategoryModel")).
		Return(childCategory, nil)

	// Act
	output, err := s.sut.Execute(ctx, input)

	// Assert
	s.Require().NoError(err)
	s.Equal(uint64(2), output.ID)
	s.Require().NotNil(output.ParentID)
	s.Equal(parentID, *output.ParentID)
}

func (s *CreateCategoryCommandTestSuite) TestExecute_ParentNotFound_ReturnsError() {
	// Arrange
	ctx := context.Background()
	parentID := uint64(99)
	input := command.CreateCategoryCommandInput{Name: "手机", ParentID: &parentID}

	s.categoryRepositoryMock.
		On("FindByID", mock.Anything, parentID).
		Return(model.CategoryModel{}, errs.ErrCategoryNotFound)

	// Act
	output, err := s.sut.Execute(ctx, input)

	// Assert
	s.Require().ErrorIs(err, errs.ErrCategoryParentNotFound)
	s.Equal(command.CreateCategoryCommandOutput{}, output)
	s.categoryRepositoryMock.AssertNotCalled(s.T(), "Create")
}

func (s *CreateCategoryCommandTestSuite) TestExecute_BlankName_ReturnsError() {
	// Arrange
	ctx := context.Background()
	input := command.CreateCategoryCommandInput{Name: "  "}

	s.loggerMock.On("Error").Return(nil)

	// Act
	output, err := s.sut.Execute(ctx, input)

	// Assert
	s.Require().ErrorIs(err, errs.ErrCategoryNameRequired)
	s.Equal(command.CreateCategoryCommandOutput{}, output)
}

func (s *CreateCategoryCommandTestSuite) TestExecute_RepositoryError_ReturnsError() {
	// Arrange
	ctx := context.Background()
	input := command.CreateCategoryCommandInput{Name: "数码"}
	repositoryErr := errors.New("repository error")

	s.categoryRepositoryMock.
		On("Create", mock.Anything, mock.AnythingOfType("model.CategoryModel")).
		Return(model.CategoryModel{}, repositoryErr)

	s.loggerMock.On("Error").Return(nil)

	// Act
	output, err := s.sut.Execute(ctx, input)

	// Assert
	s.Require().ErrorIs(err, repositoryErr)
	s.Equal(command.CreateCategoryCommandOutput{}, output)
}

func (s *CreateCategoryCommandTestSuite) TestExecute_ParentAtMaxDepth_ReturnsDepthExceeded() {
	// Arrange
	ctx := context.Background()
	parentID := uint64(1)
	input := command.CreateCategoryCommandInput{Name: "手机壳", ParentID: &parentID}

	deepParent, _ := model.RestoreCategoryModel(
		parentID, "手机", nil, model.MaxCategoryDepth, "/1", 0, 1, s.mockCreatedAt, s.mockCreatedAt,
	)

	s.categoryRepositoryMock.
		On("FindByID", mock.Anything, parentID).
		Return(deepParent, nil)
	s.loggerMock.On("Error").Return(nil)

	// Act
	output, err := s.sut.Execute(ctx, input)

	// Assert
	s.Require().ErrorIs(err, errs.ErrCategoryDepthExceeded)
	s.Equal(command.CreateCategoryCommandOutput{}, output)
	s.categoryRepositoryMock.AssertNotCalled(s.T(), "Create")
}
