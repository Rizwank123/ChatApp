package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/gofrs/uuid/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/Rizwank123/chatApp/internal/cache"
	"github.com/Rizwank123/chatApp/internal/domain"
	"github.com/Rizwank123/chatApp/internal/pkg/config"
	"github.com/Rizwank123/chatApp/internal/pkg/security"
	"github.com/Rizwank123/chatApp/internal/pkg/util"
)

type UserServiceImpl struct {
	apu     util.AppUtil
	cfg     config.ChatApiConfig
	pr      domain.PersonnelRepository
	tr      domain.Transactioner
	scm     security.Manager
	ur      domain.UserRepository
	checker *cache.UserExistenceChecker
}

func NewUserService(
	apu util.AppUtil,
	cfg config.ChatApiConfig,
	pr domain.PersonnelRepository,
	smc security.Manager,
	tr domain.Transactioner,
	ur domain.UserRepository,
	checker *cache.UserExistenceChecker,
) domain.UserService {
	if checker == nil {
		checker = cache.NewUserExistenceChecker(
			cache.NewMemoryBloomFilter(100000, 0.01),
			ur.ExistsByUserName,
		)
	}
	return &UserServiceImpl{
		cfg:     cfg,
		apu:     apu,
		pr:      pr,
		ur:      ur,
		scm:     smc,
		tr:      tr,
		checker: checker,
	}
}

// CheckUsernameAvailability validates username syntax, checks reserved list, and uses Bloom filter.
func (s *UserServiceImpl) CheckUsernameAvailability(rawUsername string) (domain.CheckUsernameAvailabilityOutput, error) {
	ctx := context.Background()
	normalized, err := util.ValidateUsername(rawUsername)
	if err != nil {
		return domain.CheckUsernameAvailabilityOutput{
			UserName:    rawUsername,
			IsAvailable: false,
			Reason:      err.Error(),
		}, nil
	}

	exists, err := s.checker.Exists(ctx, normalized)
	if err != nil {
		// Safe fallback: check DB directly
		exists, err = s.ur.ExistsByUserName(ctx, normalized)
		if err != nil {
			return domain.CheckUsernameAvailabilityOutput{}, fmt.Errorf("failed to check username availability: %w", err)
		}
	}

	if exists {
		return domain.CheckUsernameAvailabilityOutput{
			UserName:    normalized,
			IsAvailable: false,
			Reason:      "username is already taken",
		}, nil
	}

	return domain.CheckUsernameAvailabilityOutput{
		UserName:    normalized,
		IsAvailable: true,
	}, nil
}

// CreateUser implements domain.UserService.
func (s *UserServiceImpl) CreateUser(in domain.RegisterUserInput) (result domain.User, err error) {
	ctx := context.Background()

	// 1. Validate & normalize username
	normalizedUsername, valErr := util.ValidateUsername(in.UserName)
	if valErr != nil {
		return result, valErr
	}
	in.UserName = normalizedUsername

	// 2. Pre-filter with Bloom Filter (authoritative check)
	exists, chkErr := s.checker.Exists(ctx, normalizedUsername)
	if chkErr == nil && exists {
		return result, fmt.Errorf("user with username '%s' already exists, please choose another username", in.UserName)
	}

	// 3. Database transaction
	ctx, err = s.tr.Begin(ctx)
	if err != nil {
		return result, err
	}

	defer func() {
		s.tr.Rollback(ctx, err)
	}()

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
		// Handle database unique constraint conflict gracefully
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return result, fmt.Errorf("username '%s' is already taken", in.UserName)
		}
		if strings.Contains(strings.ToLower(err.Error()), "unique") || strings.Contains(strings.ToLower(err.Error()), "duplicate") {
			return result, fmt.Errorf("username '%s' is already taken", in.UserName)
		}
		return result, err
	}

	displayName := strings.TrimSpace(in.FirstName + " " + in.LastName)
	personnel := domain.Personnel{
		FirstName:        in.FirstName,
		LastName:         in.LastName,
		DisplayName:      &displayName,
		UserID:           result.ID,
		Mobile:           result.UserName,
		Role:             in.Role,
		ActivationStatus: domain.ActivationStatusACTIVE,
	}

	err = s.pr.Create(ctx, &personnel)
	if err != nil {
		return result, err
	}

	// commit the transaction
	err = s.tr.Commit(ctx)
	if err != nil {
		return result, err
	}

	// 4. Update Bloom filter post-commit
	_ = s.checker.RecordUser(ctx, in.UserName)

	return result, nil
}

func (s *UserServiceImpl) Login(in domain.LoginInput) (result domain.LoginOutput, err error) {
	normUsername := util.NormalizeUsername(in.UserName)

	// check if user exists
	usr, err := s.ur.FindByUserName(context.Background(), normUsername)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return result, errors.New("invalid username or password")
		}
		return result, err
	}
	if usr.UserName == "" || usr.ID.IsNil() {
		return result, errors.New("invalid username or password")
	}
	match, err := s.apu.PasswordCheck(*usr.Password, in.Password)
	if err != nil || !match {
		return result, errors.New("invalid username or password")
	}
	personnel, err := s.pr.FindByUserID(context.Background(), usr.ID)
	if err != nil {
		return result, err
	}
	// generate token
	ti := security.TokenMetadata{
		UserID:      usr.ID.String(),
		Role:        usr.Role,
		PersonnelID: personnel.ID.String(),
	}
	token, err := s.scm.GenerateAuthToken(ti)
	if err != nil {
		return result, err
	}
	result = domain.LoginOutput{
		Token:     token,
		ExpiresIn: int64(s.cfg.AuthExpiryPeriod),
	}
	return result, nil
}

func (s *UserServiceImpl) FindByID(id uuid.UUID) (result domain.User, err error) {
	return s.ur.FindByID(context.Background(), id)
}

func (s *UserServiceImpl) FindByUserName(username string) (result domain.User, err error) {
	norm := util.NormalizeUsername(username)
	return s.ur.FindByUserName(context.Background(), norm)
}

func (s *UserServiceImpl) UpdateUser(id uuid.UUID, in domain.UpdateUserInput) (result domain.User, err error) {
	norm, err := util.ValidateUsername(in.UserName)
	if err != nil {
		return result, err
	}
	in.UserName = norm

	entity := domain.User{
		Base:     domain.Base{ID: id},
		UserName: in.UserName,
		Role:     string(in.Role),
	}
	if in.Password != "" {
		pass, err := s.apu.EncryptPassword(in.Password)
		if err != nil {
			return result, err
		}
		entity.Password = &pass
	}
	err = s.ur.UpdateUser(context.Background(), &entity)
	if err != nil {
		return result, err
	}
	_ = s.checker.RecordUser(context.Background(), in.UserName)
	return entity, nil
}

func (s *UserServiceImpl) DeleteUser(id uuid.UUID) (err error) {
	return s.ur.DeleteUser(context.Background(), id)
}
