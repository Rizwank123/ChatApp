package service

import (
	"context"

	"github.com/gofrs/uuid/v5"

	"github.com/chatApp/internal/domain"
)

type personnelServiceImpl struct {
	pr domain.PersonnelRepository
}

func NewPersonnelService(pr domain.PersonnelRepository) domain.PersonnelService {
	return &personnelServiceImpl{pr: pr}
}

// Filter implements domain.PersonnelService.
func (s *personnelServiceImpl) Filter(in domain.FilterInput, options domain.QueryOptions) (result []domain.Personnel, total int64, err error) {
	return s.pr.Filter(context.Background(), in, options)
}

// FindByID implements domain.PersonnelService.
func (s *personnelServiceImpl) FindByID(id uuid.UUID) (result domain.Personnel, err error) {
	return s.pr.FindByID(context.Background(), id)
}

// FindByUserID implements domain.PersonnelService.
func (s *personnelServiceImpl) FindByUserID(userID uuid.UUID) (result domain.Personnel, err error) {
	return s.pr.FindByUserID(context.Background(), userID)
}

// Create implements domain.PersonnelService.
func (s *personnelServiceImpl) Create(in domain.CreatePersonnelInput) (result domain.Personnel, err error) {
	result = domain.Personnel{
		FirstName:        in.FirstName,
		LastName:         in.LastName,
		Email:            &in.Email,
		Mobile:           in.Mobile,
		Role:             in.Role,
		Avatar:           &in.Avatar,
		Address:          in.Address,
		ActivationStatus: domain.ActivationStatusACTIVE,
	}
	err = s.pr.Create(context.Background(), &result)
	if err != nil {
		return result, err
	}
	return result, nil
}

// Update implements domain.PersonnelService.
func (s *personnelServiceImpl) Update(id uuid.UUID, in domain.UpdatePersonnelInput) (result domain.Personnel, err error) {
	result, err = s.pr.FindByID(context.Background(), id)
	if err != nil {
		return result, err
	}
	if in.FirstName != "" {
		result.FirstName = in.FirstName
	}
	if in.LastName != "" {
		result.LastName = in.LastName
	}
	if in.Email != "" {
		result.Email = &in.Email
	}
	if in.Gender != "" {
		result.Gender = &in.Gender

	}
	if in.Role != "" {
		result.Role = in.Role
	}
	if in.Avatar != "" {
		result.Avatar = &in.Avatar
	}
	if in.Address.City != "" {
		result.Address.City = in.Address.City
	}
	if in.Address.Country != "" {
		result.Address.Country = in.Address.Country
	}
	if in.Address.State != "" {
		result.Address.State = in.Address.State
	}
	if in.Address.Street != "" {
		result.Address.Street = in.Address.Street
	}
	if in.Address.Pincode != "" {
		result.Address.Pincode = in.Address.Pincode
	}
	if in.ActivationStatus != "" {
		result.ActivationStatus = in.ActivationStatus
	}
	err = s.pr.Update(context.Background(), &result)
	if err != nil {
		return result, err
	}
	return result, nil

}

// Delete implements domain.PersonnelService.
func (s *personnelServiceImpl) Delete(id uuid.UUID) (err error) {
	return s.pr.Delete(context.Background(), id)
}
