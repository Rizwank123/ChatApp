package controller

import (
	"net/http"

	"github.com/gofrs/uuid/v5"
	"github.com/labstack/echo/v4"

	"github.com/Rizwank123/chatApp/internal/domain"
	"github.com/Rizwank123/chatApp/internal/http/transport"
	"github.com/Rizwank123/chatApp/internal/pkg/security"
)

type PersonnelController struct {
	ps domain.PersonnelService
}

func NewPersonnelController(ps domain.PersonnelService) PersonnelController {
	return PersonnelController{ps: ps}
}

func canManagePersonnel(ctx echo.Context, targetID uuid.UUID) bool {
	claims := security.GetClaimsForContext(ctx)
	if claims == nil {
		return false
	}
	if role, _ := claims["role"].(string); role == "ADMIN" {
		return true
	}
	identity, _ := claims["personnel_id"].(string)
	callerID, err := uuid.FromString(identity)
	return err == nil && callerID == targetID
}

func isAdmin(ctx echo.Context) bool {
	claims := security.GetClaimsForContext(ctx)
	role, _ := claims["role"].(string)
	return role == "ADMIN"
}

func redactPersonnelPrivateFields(ctx echo.Context, personnel *domain.Personnel) {
	if personnel == nil || canManagePersonnel(ctx, personnel.ID) {
		return
	}
	personnel.Email = nil
	personnel.Mobile = ""
	personnel.Address = domain.Address{}
	personnel.Role = ""
	personnel.ActivationStatus = ""
}

// FindPersonnelByID finds personnel by ID.
//
//	@Summary		Find personnel by ID
//	@Description	Find personnel based on the provided ID
//	@Tags			Personnel
//	@ID				findPersonnelByID
//	@Accept			json
//	@Produce		json
//	@Security		JWT
//	@Param			Authorization	header		string	true	"Bearer "
//	@Param			id				path		string	true	"Personnel ID"
//	@Success		200				{object}	domain.BaseResponse{data=domain.Personnel}
//	@Failure		400				{object}	domain.InvalidRequestError
//	@Failure		401				{object}	domain.UnauthorizedError
//	@Failure		403				{object}	domain.ForbiddenAccessError
//	@Failure		500				{object}	domain.SystemError
//	@Router			/personnel/{id} [get]
func (c PersonnelController) FindPersonnelByID(ctx echo.Context) error {
	//  get id from path
	id, err := uuid.FromString(ctx.Param("id"))
	if err != nil {
		return err
	}
	//  call service
	result, err := c.ps.FindByID(id)
	if err != nil {
		return err
	}
	redactPersonnelPrivateFields(ctx, &result)
	//  return result
	return transport.SendResponse(ctx, http.StatusOK, result)

}

// Filter filters personnel based on input criteria.
//
//	@Summary		Filter personnel
//	@Description	Filter personnel using provided criteria
//	@Tags			Personnel
//	@ID				filterPersonnel
//	@Accept			json
//	@Produce		json
//	@Security		JWT
//	@Param			Authorization	header		string				true	"Bearer "
//	@Param			body			body		domain.FilterInput	true	"Filter input"
//	@Success		200				{object}	domain.BaseResponse{data=[]domain.Personnel}
//	@Failure		400				{object}	domain.InvalidRequestError
//	@Failure		401				{object}	domain.UnauthorizedError
//	@Failure		500				{object}	domain.SystemError
//	@Router			/personnel/filter [post]
func (c PersonnelController) Filter(ctx echo.Context) error {
	if !isAdmin(ctx) {
		return echo.NewHTTPError(http.StatusForbidden, "administrator access required")
	}
	//  get filter from request body
	var in domain.FilterInput
	if err := transport.DecodeAndValidateRequestBody(ctx, &in); err != nil {
		return err
	}
	// Load query options from request param
	opt := transport.DecodeQueryOptions(ctx)
	//  call service
	result, total, err := c.ps.Filter(in, opt)
	if err != nil {
		return err
	}
	//  return result
	return transport.SendPaginationResponse(ctx, http.StatusOK, result, total)
}

// FindPersonnelByUserID finds personnel by user id.
//
//	@Summary		Find personnel by user id
//	@Description	Find personnel based on the provided user id
//	@Tags			Personnel
//	@ID				findPersonnelByUserID
//	@Accept			json
//	@Produce		json
//	@Security		JWT
//	@Param			Authorization	header		string	true	"Bearer "
//	@Param			id				path		string	true	"User ID"
//	@Success		200				{object}	domain.BaseResponse{data=domain.Personnel}
//	@Failure		400				{object}	domain.InvalidRequestError
//	@Failure		401				{object}	domain.UnauthorizedError
//	@Failure		500				{object}	domain.SystemError
//	@Router			/personnel/user/id [get]
func (c PersonnelController) FindPersonnelByUserID(ctx echo.Context) error {
	//  get user id from path
	userID, err := uuid.FromString(ctx.Param("id"))
	if err != nil {
		return err
	}
	result, err := c.ps.FindByUserID(userID)
	if err != nil {
		return err
	}
	redactPersonnelPrivateFields(ctx, &result)
	//  return result
	return transport.SendResponse(ctx, http.StatusOK, result)
}

// CreatePersonnel creates a new personnel entry.
//
//	@Summary		Create new personnel
//	@Description	Create a new personnel record
//	@Tags			Personnel
//	@ID				createPersonnel
//	@Accept			json
//	@Produce		json
//	@Security		JWT
//	@Param			Authorization	header		string						true	"Bearer "
//	@Param			body			body		domain.CreatePersonnelInput	true	"Personnel creation input"
//	@Success		201				{object}	domain.BaseResponse{data=domain.Personnel}
//	@Failure		400				{object}	domain.InvalidRequestError
//	@Failure		401				{object}	domain.UnauthorizedError
//	@Failure		500				{object}	domain.SystemError
//	@Router			/personnel [post]
func (c PersonnelController) CreatePersonnel(ctx echo.Context) error {
	if !isAdmin(ctx) {
		return echo.NewHTTPError(http.StatusForbidden, "administrator access required")
	}
	//  get input from request body
	var in domain.CreatePersonnelInput
	if err := transport.DecodeAndValidateRequestBody(ctx, &in); err != nil {
		return err
	}
	//  call service
	result, err := c.ps.Create(in)

	if err != nil {
		return err
	}
	//  return result
	return transport.SendResponse(ctx, http.StatusCreated, result)
}

// UpdatePersonnel updates an existing personnel record.
//
//	@Summary		Update personnel
//	@Description	Update personnel based on the provided ID and input
//	@Tags			Personnel
//	@ID				updatePersonnel
//	@Accept			json
//	@Produce		json
//	@Security		JWT
//	@Param			Authorization	header		string						true	"Bearer "
//	@Param			id				path		string						true	"Personnel ID"
//	@Param			body			body		domain.UpdatePersonnelInput	true	"Personnel update input"
//	@Success		200				{object}	domain.BaseResponse{data=domain.Personnel}
//	@Failure		400				{object}	domain.InvalidRequestError
//	@Failure		401				{object}	domain.UnauthorizedError
//	@Failure		403				{object}	domain.ForbiddenAccessError
//	@Failure		500				{object}	domain.SystemError
//	@Router			/personnel/{id} [put]
func (c PersonnelController) UpdatePersonnel(ctx echo.Context) error {
	//  get id from path
	id, err := uuid.FromString(ctx.Param("id"))
	if err != nil {
		return err
	}
	if !canManagePersonnel(ctx, id) {
		return echo.NewHTTPError(http.StatusForbidden, "you may only update your own profile")
	}
	//  get input from request body
	var in domain.UpdatePersonnelInput
	if err := transport.DecodeAndValidateRequestBody(ctx, &in); err != nil {
		return err
	}
	//  call service
	result, err := c.ps.Update(id, in)
	if err != nil {
		return err
	}
	//  return result
	return transport.SendResponse(ctx, http.StatusOK, result)
}

// DeletePersonnel deletes personnel based on the provided ID.
//
//	@Summary		Delete personnel
//	@Description	Delete personnel using the provided ID
//	@Tags			Personnel
//	@ID				deletePersonnel
//	@Accept			json
//	@Produce		json
//	@Security		JWT
//	@Param			Authorization	header		string	true	"Bearer "
//	@Param			id				path		string	true	"Personnel ID"
//	@Success		204				{object}	nil
//	@Failure		400				{object}	domain.InvalidRequestError
//	@Failure		401				{object}	domain.UnauthorizedError
//	@Failure		500				{object}	domain.SystemError
//	@Router			/personnel/{id} [delete]
func (c PersonnelController) DeletePersonnel(ctx echo.Context) error {
	//  get id from path
	id, err := uuid.FromString(ctx.Param("id"))
	if err != nil {
		return err
	}
	if !canManagePersonnel(ctx, id) {
		return echo.NewHTTPError(http.StatusForbidden, "you may only delete your own profile")
	}
	//  call service
	err = c.ps.Delete(id)
	if err != nil {
		return err
	}
	//  return result
	return transport.SendResponse(ctx, http.StatusNoContent, nil)
}
