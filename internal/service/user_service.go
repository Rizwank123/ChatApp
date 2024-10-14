package service

import (
	"context"
	"errors"

	"github.com/gofrs/uuid/v5"
	"github.com/jackc/pgx/v5"

	"github.com/chatApp/internal/domain"
	"github.com/chatApp/internal/pkg/config"
	"github.com/chatApp/internal/pkg/security"
	"github.com/chatApp/internal/pkg/util"
)

type UserServiceImpl struct {
	apu util.AppUtil
	cfg config.ChatApiConfig
	pr  domain.PersonnelRepository
	tr  domain.Transactioner
	scm security.Manager
	ur  domain.UserRepository
}

func NewUserService(apu util.AppUtil, cfg config.ChatApiConfig, pr domain.PersonnelRepository, smc security.Manager, tr domain.Transactioner, ur domain.UserRepository) domain.UserService {
	return &UserServiceImpl{
		cfg: cfg,
		apu: apu,
		pr:  pr,
		ur:  ur,
		scm: smc,
		tr:  tr,
	}
}

// CreateUser implements domain.UserService.
func (s *UserServiceImpl) CreateUser(in domain.RegisterUserInput) (result domain.User, err error) {
	ctx := context.Background()
	ctx, err = s.tr.Begin(ctx)
	if err != nil {
		return result, err
	}

	defer func() {
		s.tr.Rollback(ctx, err)
	}()

	result, err = s.ur.FindByUserName(ctx, in.UserName)
	if err != nil {
		if !errors.Is(err, pgx.ErrNoRows) {
			return result, err
		}
	}
	if result.UserName == in.UserName {
		return result, errors.New("user with this username already exists: " + result.UserName + "please login or use another user name ")
	}
	pass, err := s.apu.EncryptPassword(in.Password)
	if err != nil {
		return result, err
	}

	result = domain.User{
		UserName: in.UserName,
		Password: &pass,
		Role:     string(in.Role),
	}

	err = s.ur.CreateUser(ctx, &result)
	if err != nil {
		return result, err
	}
	personnel, err := s.pr.FindByUserID(ctx, result.ID)
	if err != nil {
		if !errors.Is(err, pgx.ErrNoRows) {
			err = domain.NotFoundError{}
			return result, err
		}
	}
	if personnel.FirstName != "" && personnel.LastName != "" && personnel.ID.IsNil() {
		return result, errors.New("user already exists")
	}
	personnel.FirstName = in.FirstName
	personnel.LastName = in.LastName
	personnel.UserID = result.ID
	personnel.Mobile = result.UserName
	personnel.ActivationStatus = domain.ActivationStatusACTIVE
	err = s.pr.Create(ctx, &personnel)
	if err != nil {
		return result, err
	}
	// commit the transaction
	err = s.tr.Commit(ctx)
	if err != nil {
		return result, err
	}
	return result, nil

}
func (s *UserServiceImpl) Login(in domain.LoginInput) (result domain.LoginOutput, err error) {
	// check if user exists
	usr, err := s.ur.FindByUserName(context.Background(), in.UserName)
	if err != nil {
		return result, err
	}
	if usr.UserName != "" && usr.ID.IsNil() {
		return result, errors.New("user with this username:" + in.UserName + " does not exist")
	}
	match, err := s.apu.PasswordCheck(*usr.Password, in.Password)
	if err != nil || !match {
		return result, errors.New("wrong password")
	}
	// generate token
	ti := security.TokenMetadata{
		UserID: usr.ID.String(),
		Role:   usr.Role,
	}
	token, err := s.scm.GenerateAuthToken(ti)
	if err != nil {
		return result, err
	}
	result = domain.LoginOutput{
		Token:     token,
		ExpiresIn: int64(s.cfg.AuthExpiryPeriod),
	}
	return result, err
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
		Password: &in.Password,
		Role:     string(in.Role),
	}
	return result, s.ur.UpdateUser(context.Background(), &result)
}
