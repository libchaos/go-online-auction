package mapper_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"auction/internal/modules/users/domain/enum"
	"auction/internal/modules/users/domain/model"
	"auction/internal/modules/users/infra/mapper"
	"auction/internal/modules/users/infra/sqlcgen"
)

type UserMapperTestSuite struct {
	suite.Suite
	sut *mapper.UserMapper
}

func (s *UserMapperTestSuite) SetupTest() {
	s.sut = mapper.NewUserMapper()
}

func TestUserMapperSuite(t *testing.T) {
	suite.Run(t, new(UserMapperTestSuite))
}

func (s *UserMapperTestSuite) TestToDomain_ValidRow_ReturnsUserModel() {
	// Arrange
	now := time.Now().UTC()
	passwordHash := "hashed-password"
	u := sqlcgen.User{
		ID:           1,
		Name:         "John Doe",
		Email:        "john@example.com",
		PasswordHash: &passwordHash,
		Role:         sqlcgen.UserRoleBidder,
		Status:       sqlcgen.UserStatusActive,
		Version:      3,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	// Act
	result, err := s.sut.ToDomain(u)

	// Assert
	s.Require().NoError(err)
	s.Equal(uint64(u.ID), result.ID())
	s.Equal(u.Name, result.Name())
	s.Equal(u.Email, result.Email())
	s.Equal(passwordHash, *result.PasswordHash())
	role := result.Role()
	s.Equal(string(u.Role), role.String())
	status := result.Status()
	s.Equal(string(u.Status), status.String())
	s.Equal(uint64(u.Version), result.Version())
	s.Equal(now, result.CreatedAt())
	s.Equal(now, result.UpdatedAt())
}

func (s *UserMapperTestSuite) TestToDomain_InvalidRole_ReturnsError() {
	// Arrange
	now := time.Now().UTC()
	passwordHash := "hashed-password"
	u := sqlcgen.User{
		ID:           1,
		Name:         "John Doe",
		Email:        "john@example.com",
		PasswordHash: &passwordHash,
		Role:         "invalid_role",
		Status:       sqlcgen.UserStatusActive,
		Version:      1,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	// Act
	result, err := s.sut.ToDomain(u)

	// Assert
	s.Require().Error(err)
	s.Equal(model.UserModel{}, result)
}

func (s *UserMapperTestSuite) TestToCreateParams_ValidUserModel_ReturnsParams() {
	// Arrange
	now := time.Now().UTC()
	passwordHash := "hashed-password"
	role, _ := enum.NewRoleEnum("bidder")
	status, _ := enum.NewUserStatusEnum("active")
	user, err := model.RestoreUserModel(
		1,
		"John Doe",
		"john@example.com",
		&passwordHash,
		role,
		status,
		nil,
		nil,
		3,
		now,
		now,
	)
	s.Require().NoError(err)

	// Act
	result := s.sut.ToCreateParams(user)

	// Assert
	s.Equal(user.Name(), result.Name)
	s.Equal(user.Email(), result.Email)
	s.Equal(passwordHash, *result.PasswordHash)
	s.Equal(sqlcgen.UserRoleBidder, result.Role)
	s.Equal(sqlcgen.UserStatusActive, result.Status)
	s.Nil(result.OauthProvider)
	s.Nil(result.OauthProviderID)
	s.Equal(int64(user.Version()), result.Version)
	s.Equal(user.CreatedAt(), result.CreatedAt)
	s.Equal(user.UpdatedAt(), result.UpdatedAt)
}

func (s *UserMapperTestSuite) TestToUpdateParams_ValidUserModel_ReturnsParams() {
	// Arrange
	now := time.Now().UTC()
	passwordHash := "hashed-password"
	role, _ := enum.NewRoleEnum("admin")
	status, _ := enum.NewUserStatusEnum("blocked")
	user, err := model.RestoreUserModel(
		7,
		"Jane Doe",
		"jane@example.com",
		&passwordHash,
		role,
		status,
		nil,
		nil,
		5,
		now,
		now,
	)
	s.Require().NoError(err)

	// Act
	result := s.sut.ToUpdateParams(user)

	// Assert
	s.Equal(int64(user.ID()), result.ID)
	s.Equal(user.Name(), result.Name)
	s.Equal(user.Email(), result.Email)
	s.Equal(sqlcgen.UserRoleAdmin, result.Role)
	s.Equal(sqlcgen.UserStatusBlocked, result.Status)
	s.Equal(int64(user.Version()), result.Version)
	s.Equal(user.UpdatedAt(), result.UpdatedAt)
}
