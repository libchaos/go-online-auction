package query_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"auction/internal/modules/users/application/query"
	"auction/internal/modules/users/domain/enum"
	"auction/internal/modules/users/domain/errs"
	"auction/internal/modules/users/domain/model"
	"auction/tests/mocks"
)

type GetUserByIDQueryTestSuite struct {
	suite.Suite
	sut                *query.GetUserByIDQuery
	userRepositoryMock *mocks.MockUserRepository
	loggerMock         *mocks.MockLogger
	mockUser           model.UserModel
}

func (s *GetUserByIDQueryTestSuite) SetupTest() {
	s.userRepositoryMock = mocks.NewMockUserRepository(s.T())
	s.loggerMock = mocks.NewMockLogger(s.T())

	s.sut = query.NewGetUserByIDQuery(
		s.userRepositoryMock,
		s.loggerMock,
	)

	role, _ := enum.NewRoleEnum(enum.EnumRoleSeller)
	status, _ := enum.NewUserStatusEnum(enum.EnumUserStatusActive)
	passwordHash := "hashed-password"
	now := time.Now().UTC()
	s.mockUser, _ = model.RestoreUserModel(
		1, "John Doe", "john@example.com", &passwordHash, role, status, nil, nil, 1, now, now,
	)
}

func TestGetUserByIDQuerySuite(t *testing.T) {
	suite.Run(t, new(GetUserByIDQueryTestSuite))
}

func (s *GetUserByIDQueryTestSuite) TestExecute_ValidID_ReturnsUser() {
	// Arrange
	ctx := context.Background()
	input := query.GetUserByIDQueryInput{UserID: 1}

	s.userRepositoryMock.On("FindByID", mock.Anything, uint64(1)).Return(s.mockUser, nil)

	// Act
	output, err := s.sut.Execute(ctx, input)

	// Assert
	s.Require().NoError(err)
	s.Equal(uint64(1), output.ID)
	s.Equal("John Doe", output.Name)
	s.Equal("john@example.com", output.Email)
	s.Equal(enum.EnumRoleSeller, output.Role)
}

func (s *GetUserByIDQueryTestSuite) TestExecute_ZeroID_ReturnsError() {
	// Arrange
	ctx := context.Background()
	input := query.GetUserByIDQueryInput{UserID: 0}

	// Act
	output, err := s.sut.Execute(ctx, input)

	// Assert
	s.Require().ErrorIs(err, errs.ErrUserIDRequired)
	s.Equal(query.GetUserByIDQueryOutput{}, output)
}

func (s *GetUserByIDQueryTestSuite) TestExecute_UserNotFound_ReturnsError() {
	// Arrange
	ctx := context.Background()
	input := query.GetUserByIDQueryInput{UserID: 99}

	emptyUser := model.UserModel{}
	s.userRepositoryMock.
		On("FindByID", mock.Anything, uint64(99)).
		Return(emptyUser, errs.ErrUserNotFound)
	s.loggerMock.On("Error").Return(nil)

	// Act
	output, err := s.sut.Execute(ctx, input)

	// Assert
	s.Require().ErrorIs(err, errs.ErrUserNotFound)
	s.Equal(query.GetUserByIDQueryOutput{}, output)
}
