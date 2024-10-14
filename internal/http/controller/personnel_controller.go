package controller

import (
	"net/http"

	"github.com/gofrs/uuid/v5"
	"github.com/labstack/echo/v4"

	"github.com/chatApp/internal/domain"
	"github.com/chatApp/internal/http/transport"
)

type PersonnelController struct {
	ps domain.PersonnelService
}

func NewPersonnelController(ps domain.PersonnelService) PersonnelController {
	return PersonnelController{ps: ps}
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
	//  get filter from request body
	var in domain.FilterInput
	transport.DecodeAndValidateRequestBody(ctx, &in)
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
	//  get input from request body
	var in domain.CreatePersonnelInput
	transport.DecodeAndValidateRequestBody(ctx, &in)
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
	//  get input from request body
	var in domain.UpdatePersonnelInput
	transport.DecodeAndValidateRequestBody(ctx, &in)
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
	//  call service
	err = c.ps.Delete(id)
	if err != nil {
		return err
	}
	//  return result
	return transport.SendResponse(ctx, http.StatusNoContent, nil)
}
