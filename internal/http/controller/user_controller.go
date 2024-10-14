package controller

import (
	"net/http"
	"net/url"

	"github.com/gofrs/uuid/v5"
	"github.com/labstack/echo/v4"

	"github.com/chatApp/internal/domain"
	"github.com/chatApp/internal/http/transport"
)

type UserController struct {
	ur domain.UserService
}

func NewUserController(ur domain.UserService) UserController {
	return UserController{ur: ur}
}

// FindByID finds a user by ID.
//
//	@Summary		Find a user by ID
//	@Description	Find a user based on the provided ID
//	@Tags			User
//	@ID				findUserByID
//	@Accept			json
//	@Produce		json
//	@Security		JWT
//	@Param			Authorization	header		string	true	"Bearer "
//	@Param			id				path		string	true	"User ID"
//	@Success		200				{object}	domain.BaseResponse{data=domain.User}
//	@Failure		400				{object}	domain.InvalidRequestError
//	@Failure		401				{object}	domain.UnauthorizedError
//	@Failure		403				{object}	domain.ForbiddenAccessError
//	@Failure		500				{object}	domain.SystemError
//	@Router			/users/{id} [get]
func (c UserController) FindByID(ctx echo.Context) error {
	// Parse the path param
	id, err := uuid.FromString(ctx.Param("id"))
	if err != nil {
		return err

	}
	// Call the service to find the user by id
	result, err := c.ur.FindByID(id)
	if err != nil {
		return err
	}
	// Return the result
	return transport.SendResponse(ctx, http.StatusOK, result)
}

// FindByUserName Find a user by username
//
//	@Summary		Find a user by username
//	@Description	Get user information by their username
//	@Tags			User
//	@Accept			json
//	@Produce		json
//	@Security		JWT
//	@Param			Authorization	header		string	true	"Bearer "
//	@Param			username		path		string	true	"username"
//	@Success		200				{object}	domain.BaseResponse{data=domain.User}"
//	@Failure		400				{object}	domain.InvalidRequestError
//	@Failure		401				{object}	domain.UnauthorizedError
//	@Failure		403				{object}	domain.ForbiddenAccessError
//	@Failure		500				{object}	domain.SystemError
//	@Router			/users/{username} [get]
func (c UserController) FindByUserName(ctx echo.Context) error {
	// get the username from the path param
	encodedUserName := ctx.Param("username")

	// Decode the URL-encoded username (e.g., decode %2B to +)
	userName, err := url.QueryUnescape(encodedUserName)
	if err != nil {
		return err
	}
	// Call the service method to find user by username
	result, err := c.ur.FindByUserName(userName)
	if err != nil {
		return err
	}
	// Return the result
	return transport.SendResponse(ctx, http.StatusOK, result)
}

// Login authenticates a user based on login credentials.
//
//	@Summary		User login
//	@Description	Authenticate a user using provided credentials
//	@Tags			Auth
//	@ID				userLogin
//	@Accept			json
//	@Produce		json
//	@Param			body	body		domain.LoginInput	true	"Login input"
//	@Success		200		{object}	domain.BaseResponse{data=domain.LoginOutput}
//	@Failure		400		{object}	domain.InvalidRequestError
//	@Failure		401		{object}	domain.UnauthorizedError
//	@Failure		500		{object}	domain.SystemError
//	@Router			/users/login [post]
func (c UserController) Login(ctx echo.Context) error {
	// Decode the request body
	var in domain.LoginInput
	err := transport.DecodeAndValidateRequestBody(ctx, &in)
	if err != nil {
		return err
	}
	// Call the service to login
	result, err := c.ur.Login(in)
	if err != nil {
		return err
	}
	// Return the result
	return transport.SendResponse(ctx, http.StatusOK, result)

}

// RegisterUser  Register a new user
//
//	@Summary		Register a new user
//	@Description	Create a new user with the provided details
//	@Tags			Auth
//	@Accept			json
//	@Produce		json
//	@Param			user	body		domain.RegisterUserInput	true	"User registration details"
//	@Success		200		{object}	domain.BaseResponse{data=domain.User}"
//	@Failure		400		{object}	domain.InvalidRequestError
//	@Failure		401		{object}	domain.UnauthorizedError
//	@Failure		403		{object}	domain.ForbiddenAccessError
//	@Failure		500		{object}	domain.SystemError
//	@Router			/users [post]
func (c UserController) RegisterUser(ctx echo.Context) error {
	// Decode the request body
	var in domain.RegisterUserInput
	transport.DecodeAndValidateRequestBody(ctx, &in)

	// Call service method to create  a new user
	result, err := c.ur.CreateUser(in)
	if err != nil {
		return err
	}
	// Send the response
	return transport.SendResponse(ctx, http.StatusCreated, result)
}
