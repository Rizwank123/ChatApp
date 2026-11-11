package domain

import (
	"context"

	"github.com/gofrs/uuid/v5"
)

type (
	// UserRole represents the role a user in the system.
	UserRole string
)
type (
	// User defines the module for User
	User struct {
		Base
		UserName string  `db:"user_name" json:"user_name,omitempty" example:"+919984778491"`
		Password *string `db:"password" json:"-"`
		Role     string  `db:"role" json:"role,omitempty"  example:"ADMIN"`
		BaseAudit
	} // @name User
)

type (
	// CreateUserInput define the module for CreateUser
	RegisterUserInput struct {
		FirstName string   `json:"first_name" example:"Mohammad"`
		LastName  string   `json:"last_name" example:"Rizwan"`
		UserName  string   `json:"user_name" example:"+919984778491"`
		Role      UserRole `json:"role" example:"ADMIN"`
		Password  string   `json:"password" example:"password123"`
	} // @name CreateUserInput
	// UpdateUserInput define the module for the UpdateUserInput
	UpdateUserInput struct {
		RegisterUserInput
	} // @name UpdateUserInput
	// LoginInput  define the module for the LoginInput
	LoginInput struct {
		UserName string `json:"username" example:"+919876543210"`
		Password string `json:"password"`
	} // @name LoginInput
	// LoginOutput define the module for the LoginOutput
	LoginOutput struct {
		Token     string `json:"token"`
		ExpiresIn int64  `json:"expires_in"`
	} // @name LoginOutput

	// CheckUsernameAvailabilityInput input for checking username
	CheckUsernameAvailabilityInput struct {
		UserName string `json:"username" query:"username" validate:"required"`
	}

	// CheckUsernameAvailabilityOutput response for checking username
	CheckUsernameAvailabilityOutput struct {
		UserName    string `json:"username"`
		IsAvailable bool   `json:"is_available"`
		Reason      string `json:"reason,omitempty"`
	}

	// UserProfileSearchResult item for search results
	UserProfileSearchResult struct {
		UserID      uuid.UUID `json:"user_id"`
		PersonnelID uuid.UUID `json:"personnel_id"`
		UserName    string    `json:"user_name"`
		DisplayName string    `json:"display_name"`
		Avatar      *string   `json:"avatar,omitempty"`
		Bio         *string   `json:"bio,omitempty"`
		IsFollowing bool      `json:"is_following"`
		FollowsYou  bool      `json:"follows_you"`
		IsMutual    bool      `json:"is_mutual"`
	}
)

type (
	// UserRepository defines the methods that any use repository should implements
	UserRepository interface {
		// FindByID return the user by id
		FindByID(ctx context.Context, id uuid.UUID) (result User, err error)
		// FindByUserName return the user by username
		FindByUserName(ctx context.Context, username string) (result User, err error)
		// ExistsByUserName checks if a normalized username exists in PostgreSQL
		ExistsByUserName(ctx context.Context, username string) (bool, error)
		// GetAllUsernames fetches all usernames for Bloom filter backfill
		GetAllUsernames(ctx context.Context) ([]string, error)
		// Filter filters users by criteria.
		// limit and offset are used for pagination.
		// total is the total number of entities.
		Filter(ctx context.Context, in FilterInput, options QueryOptions) (result []User, total int64, err error)
		// CreateUser creates a new user
		CreateUser(ctx context.Context, entity *User) (err error)
		// UpdateUser updates the user
		UpdateUser(ctx context.Context, entity *User) (err error)
		// DeleteUser deletes the user
		DeleteUser(ctx context.Context, id uuid.UUID) (err error)
	} // @name UserRepository

	// UserService defines the methods that any use service should implements
	UserService interface {
		// CreateUser creates a new user
		CreateUser(in RegisterUserInput) (result User, err error)
		// FindByID  return the user by id
		FindByID(id uuid.UUID) (result User, err error)
		// FindByUserName return the user by username
		FindByUserName(username string) (result User, err error)
		// CheckUsernameAvailability validates and checks availability with Bloom filter
		CheckUsernameAvailability(rawUsername string) (CheckUsernameAvailabilityOutput, error)
		// Login return the user by username and password
		Login(in LoginInput) (result LoginOutput, err error)
		// UpdateUser updates the user
		UpdateUser(id uuid.UUID, in UpdateUserInput) (result User, err error)
		// DeleteUser deletes the user
		DeleteUser(id uuid.UUID) (err error)
	} // @name UserService

)

const (
	UserRoleAmin UserRole = "ADMIN"
	UserRoleUser UserRole = "USER"
)
