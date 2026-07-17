package query_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"auction/internal/modules/users/application/query"
	"auction/internal/modules/users/domain/enum"
	"auction/internal/modules/users/domain/model"
	"auction/tests/mocks"
)

type ListUsersQueryTestSuite struct {
	suite.Suite
	sut                *query.ListUsersQuery
	userRepositoryMock *mocks.MockUserRepository
	loggerMock         *mocks.MockLogger
	mockUsers          []model.UserModel
}

func (s *ListUsersQueryTestSuite) SetupTest() {
	s.userRepositoryMock = mocks.NewMockUserRepository(s.T())
	s.loggerMock = mocks.NewMockLogger(s.T())

	s.sut = query.NewListUsersQuery(
		s.userRepositoryMock,
		s.loggerMock,
	)

	role, _ := enum.NewRoleEnum(enum.EnumRoleBidder)
	status, _ := enum.NewUserStatusEnum(enum.EnumUserStatusActive)
	passwordHash := "hashed-password"
	now := time.Now().UTC()
	user, _ := model.RestoreUserModel(
		1, "John Doe", "john@example.com", &passwordHash, role, status, nil, nil, 1, now, now,
	)
	s.mockUsers = []model.UserModel{user}
}

func TestListUsersQuerySuite(t *testing.T) {
	suite.Run(t, new(ListUsersQueryTestSuite))
}

func (s *ListUsersQueryTestSuite) TestExecute_ValidInput_ReturnsUsers() {
	// Arrange
	ctx := context.Background()
	input := query.ListUsersQueryInput{Limit: 10, Offset: 0}

	s.userRepositoryMock.On("FindAllPaginated", mock.Anything, 10, 0).Return(s.mockUsers, nil)
	s.userRepositoryMock.On("Count", mock.Anything).Return(uint64(1), nil)

	// Act
	output, err := s.sut.Execute(ctx, input)

	// Assert
	s.Require().NoError(err)
	s.Len(output.Users, 1)
	s.Equal(uint64(1), output.TotalCount)
	s.Equal(10, output.Limit)
}

func (s *ListUsersQueryTestSuite) TestExecute_ZeroLimit_UsesDefaultLimit() {
	// Arrange
	ctx := context.Background()
	input := query.ListUsersQueryInput{Limit: 0, Offset: -5}

	s.userRepositoryMock.On("FindAllPaginated", mock.Anything, 20, 0).Return(s.mockUsers, nil)
	s.userRepositoryMock.On("Count", mock.Anything).Return(uint64(1), nil)

	// Act
	output, err := s.sut.Execute(ctx, input)

	// Assert
	s.Require().NoError(err)
	s.Equal(20, output.Limit)
	s.Equal(0, output.Offset)
}

func (s *ListUsersQueryTestSuite) TestExecute_LimitAboveMax_IsClamped() {
	// Arrange
	ctx := context.Background()
	input := query.ListUsersQueryInput{Limit: 500, Offset: 0}

	s.userRepositoryMock.On("FindAllPaginated", mock.Anything, 100, 0).Return(s.mockUsers, nil)
	s.userRepositoryMock.On("Count", mock.Anything).Return(uint64(1), nil)

	// Act
	output, err := s.sut.Execute(ctx, input)

	// Assert
	s.Require().NoError(err)
	s.Equal(100, output.Limit)
}
