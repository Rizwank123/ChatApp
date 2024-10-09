package service

import (
	"context"

	"github.com/gofrs/uuid/v5"

	"github.com/chatApp/internal/domain"
)

type UserServiceImpl struct {
	ur domain.UserRepository
}

// CreateUser implements domain.UserService.
func (s *UserServiceImpl) CreateUser(in domain.RegisterUserInput) (result domain.User, err error) {
	result = domain.User{
		UserName: in.UserName,
		Password: in.Password,
		Role:     string(in.Role),
	}
	return result, s.ur.CreateUser(context.Background(), &result)
}

// DeleteUser implements domain.UserService.
func (s *UserServiceImpl) DeleteUser(id uuid.UUID) (err error) {
	return s.ur.DeleteUser(context.Background(), id)
}

// FindByID implements domain.UserService.
func (s *UserServiceImpl) FindByID(id uuid.UUID) (result domain.User, err error) {
	return s.ur.FindByID(context.Background(), id)
}

// FindByUserName implements domain.UserService.
func (s *UserServiceImpl) FindByUserName(username string) (result domain.User, err error) {
	return s.ur.FindByUserName(context.Background(), username)
}

// UpdateUser implements domain.UserService.
func (s *UserServiceImpl) UpdateUser(id uuid.UUID, in domain.UpdateUserInput) (result domain.User, err error) {
	// find user by id
	result = domain.User{
		UserName: in.UserName,
		Password: in.Password,
		Role:     string(in.Role),
	}
	return result, s.ur.UpdateUser(context.Background(), &result)
}

func NewUserService(ur domain.UserRepository) domain.UserService {
	return &UserServiceImpl{ur: ur}
}
