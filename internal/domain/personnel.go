package domain

import (
	"context"

	"github.com/gofrs/uuid/v5"
)

type (
	// Gender defines the model for personnel.gender
	Gender string // @name Gender
	// ActivationStatus defines the model for  personnel.activation_status
	ActivationStatus string // @name ActivationStatus
)
type (
	// Personnel define the model for Personnel
	Personnel struct {
		Base
		FirstName        string           `db:"first_name" json:"first_name,omitempty" example:"Mohammad"`
		LastName         string           `db:"last_name" json:"last_name,omitempty" example:"Rizwan"`
		Gender           *Gender          `db:"gender" json:"gender,omitempty" example:"MALE"`
		Email            *string          `db:"email" json:"email,omitempty" example:"expertkhan@gmail.com"`
		Mobile           string           `db:"mobile" json:"mobile,omitempty" example:"+919984778492"`
		Address          Address          `db:"address" sql:"jsonb" json:"address,omitempty"`
		Role             UserRole         `db:"role" json:"role,omitempty" example:"ADMIN"`
		Avatar           *string          `db:"avatar" json:"avatar,omitempty" example:"https://example.com/avatar.jpg"`
		UserID           uuid.UUID        `db:"user_id" json:"user_id"  example:"12345678-1234-1234-1234-123456789012"`
		ActivationStatus ActivationStatus `db:"activation_status" json:"activation_status,omitempty" example:"ACTIVATE"`
		BaseAudit
	} // @name  Personnel
)
type (
	// CreatePersonnelInput define the model for CreatePersonnelInput
	CreatePersonnelInput struct {
		FirstName        string           `json:"first_name" example:"Mohammad"`
		LastName         string           `json:"last_name,omitempty" example:"Rizwan"`
		Gender           Gender           `json:"gender,omitempty" example:"MALE"`
		Email            string           `json:"email,omitempty" example:"expertkhan@gmail.com"`
		Mobile           string           `json:"mobile,omitempty" example:"+919984778492"`
		Address          Address          `json:"address,omitempty"`
		Role             UserRole         `json:"role,omitempty" example:"ADMIN"`
		Avatar           string           `json:"avatar,omitempty" example:"https://example.com/avatar.jpg"`
		UserID           uuid.UUID        `json:"user_id" example:"12345678-1234-1234"`
		ActivationStatus ActivationStatus `json:"activation_status,omitempty" example:"ACTIVATE"`
	} // @name  CreatePersonnelInput

	// UpdatePersonnelInput define the model for  UpdatePersonnelInput
	UpdatePersonnelInput struct {
		FirstName        string           `json:"first_name,omitempty" example:"Mohammad"`
		LastName         string           `json:"last_name,omitempty" example:"Rizwan"`
		Gender           Gender           `json:"gender,omitempty" example:"MALE"`
		Address          Address          `json:"address,omitempty"`
		Email            string           `json:"email,omitempty" example:"expertkhan@gmail.com"`
		Role             UserRole         `json:"role,omitempty" example:"ADMIN"`
		Avatar           string           `json:"avatar,omitempty" example:"https://example.com/avatar.jpg"`
		ActivationStatus ActivationStatus `json:"activation_status,omitempty" example:"ACTIVATE"`
	} // @name  UpdatePersonnelInput

)

type (
	//  PersonnelRepository defines the methods  that  any personnel repository should implement
	PersonnelRepository interface {
		// FindByID  returns the personnel by id
		FindByID(ctx context.Context, id uuid.UUID) (result Personnel, err error)
		// FindByUserID  returns the personnel by user id
		FindByUserID(ctx context.Context, id uuid.UUID) (result Personnel, err error)
		// Filter filters personnel by criteria.
		// limit and offset are used for pagination.
		// total is the total number of entities.
		Filter(ctx context.Context, in FilterInput, options QueryOptions) (result []Personnel, total int64, err error)
		// Create creates a new personnel.
		Create(ctx context.Context, entity *Personnel) (err error)
		// Update  updates a personnel.
		Update(ctx context.Context, entity *Personnel) (err error)
		// Delete deletes a personnel.
		Delete(ctx context.Context, id uuid.UUID) (err error)
	}
	// PersonnelService defines the methods that personnel  service should implement
	PersonnelService interface {
		// FindByID  returns the personnel by id
		FindByID(id uuid.UUID) (result Personnel, err error)
		// FindByUserID  returns the personnel by user id
		FindByUserID(userID uuid.UUID) (result Personnel, err error)
		// Filter filters personnel by criteria.
		// limit and offset are used for pagination.
		// total is the total number of entities.
		Filter(in FilterInput, options QueryOptions) (result []Personnel, total int64, err error)
		// Create creates a new personnel.
		Create(in CreatePersonnelInput) (result Personnel, err error)
		// Update  updates a personnel.
		Update(id uuid.UUID, in UpdatePersonnelInput) (result Personnel, err error)
		// Delete deletes a personnel.
		Delete(id uuid.UUID) (err error)
	}
)

const (
	GenderMALE   = "MALE"   // @name GenderMALE
	GenderFEMALE = "FEMALE" // @name GenderFEMALE
	GenderOTHER  = "OTHER"  // @name GenderOTHER
)

const (
	ActivationStatusACTIVE   ActivationStatus = "ACTIVE"
	ActivationStatusDISABLED ActivationStatus = "DISABLED"
)
