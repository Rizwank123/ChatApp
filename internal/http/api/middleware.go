package api

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/go-playground/validator"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/labstack/echo/v4"
	echomiddleware "github.com/labstack/echo/v4/middleware"

	"github.com/Rizwank123/chatApp/internal/domain"
	"github.com/Rizwank123/chatApp/internal/http/swagger"
	"github.com/Rizwank123/chatApp/internal/http/transport"
)

// SetupMiddleware sets up middleware for the echo server
func (b ChatApi) SetupMiddleware(e *echo.Echo) {
	// Set up the validator middleware
	vv10 := validator.New()
	vv10.RegisterValidation("trim", func(fl validator.FieldLevel) bool {
		return len(strings.TrimSpace(fl.Field().String())) != 0
	})
	e.Validator = &transport.CustomValidator{Validator: vv10}
	// Set up the error handler middleware
	e.HTTPErrorHandler = errorMiddleware
	// Set the request body limit to 10M
	e.Use(echomiddleware.BodyLimit("10M"))
	// Recovery middleware recovers from panics anywhere in the chain,
	e.Use(echomiddleware.Recover())
	// Add request ID middleware
	e.Use(echomiddleware.RequestID())
	// Add CORS middleware
	origins := make([]string, 0)
	for _, origin := range strings.Split(b.cfg.AllowedOrigins, ",") {
		if origin = strings.TrimSpace(origin); origin != "" {
			origins = append(origins, origin)
		}
	}
	e.Use(echomiddleware.CORSWithConfig(echomiddleware.CORSConfig{
		AllowOrigins: origins,
		AllowMethods: []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, echo.HeaderAuthorization},
	}))
	// Add the swagger redirect middleware
	e.Use(swagger.RedirectSwagger)
	// Add the logger middleware
	e.Use(
		echomiddleware.LoggerWithConfig(
			echomiddleware.LoggerConfig{
				Format: "${time_rfc3339} ${id} ${remote_ip} ${method} ${path} ${latency_human} ${status} ${error}\n",
			},
		),
	)
}

// errorMiddleware absorbs and processes all errors
func errorMiddleware(err error, c echo.Context) {
	switch typedErr := err.(type) {
	case *echo.HTTPError:
		message := "request failed"
		if typedErr.Code < http.StatusInternalServerError {
			if text, ok := typedErr.Message.(string); ok && text != "" {
				message = text
			}
		} else {
			c.Logger().Error(err)
			message = "internal server error"
		}
		switch typedErr.Code {
		case http.StatusUnauthorized:
			_ = c.JSON(typedErr.Code, domain.UnauthorizedError{
				Code:    domain.ErrorCodeUNAUTHORIZED,
				Message: message,
			})
		case http.StatusForbidden:
			_ = c.JSON(typedErr.Code, domain.ForbiddenAccessError{
				Code:    domain.ErrorCodeFORBIDDEN_ACCESS,
				Message: message,
			})
		case http.StatusNotFound:
			_ = c.JSON(typedErr.Code, domain.NotFoundError{})
		case http.StatusBadRequest:
			_ = c.JSON(typedErr.Code, domain.InvalidRequestError{Message: message})
		default:
			_ = c.JSON(typedErr.Code, domain.SystemError{Code: domain.ErrorCodeINTERNAL_SERVER_ERROR, Message: message})
		}

	case validator.ValidationErrors:
		var ve error
		fields := make([]string, 0)
		errs := err.(validator.ValidationErrors)

		for _, e := range errs {
			if e.Tag() == "required" {

				fields = append(fields, fmt.Sprintf("%s is required", e.Field()))
				continue
			}

			if e.Tag() == "e164" {
				fields = append(fields, fmt.Sprintf("%s is an invalid mobile number", e.Field()))
				continue
			}

			if e.Tag() == "email" {
				fields = append(fields, fmt.Sprintf("%s is an invalid email address", e.Field()))
				continue
			}

			if e.Tag() == "oneof" {
				fields = append(fields, fmt.Sprintf("%s must be one of %s", e.Field(), e.Param()))
				continue
			}

			if e.Tag() == "max" {
				fields = append(fields, fmt.Sprintf("%s must not exceed %s characters", e.Field(), e.Param()))
				continue
			}

		}

		ve = domain.ValidationError{
			Code:    domain.ErrorCodeVALIDATION_ERROR,
			Message: domain.MessageVALIDATIONFAILED,
			Fields:  fields,
		}
		_ = c.JSON(http.StatusBadRequest, ve)

	case *pgconn.PgError:
		c.Logger().Error(err)
		res := domain.SystemError{
			Code:    domain.ErrorCodeINTERNAL_SERVER_ERROR,
			Message: "internal server error",
		}
		_ = c.JSON(http.StatusInternalServerError, res)

	case domain.DataNotFoundError:
		res := domain.UserError{
			Code:    domain.ErrorCodeINVALID_REQUEST,
			Message: "requested record was not found",
		}
		_ = c.JSON(http.StatusBadRequest, res)

	case domain.UserError:
		usrErr := err.(domain.UserError)
		res := domain.UserError{
			Code:    usrErr.Code,
			Message: usrErr.Message,
		}
		_ = c.JSON(http.StatusBadRequest, res)

	case domain.UnauthorizedError:
		res := domain.UnauthorizedError{
			Code:    domain.ErrorCodeUNAUTHORIZED,
			Message: domain.MessageUNAUTHORIZEDACCESS,
		}
		_ = c.JSON(http.StatusUnauthorized, res)

	case domain.ForbiddenAccessError:
		res := domain.ForbiddenAccessError{
			Code:    domain.ErrorCodeFORBIDDEN_ACCESS,
			Message: domain.MessageFORBIDDENACCESS,
		}
		_ = c.JSON(http.StatusForbidden, res)

	default:
		c.Logger().Error(err)
		res := domain.SystemError{
			Code:    domain.ErrorCodeINTERNAL_SERVER_ERROR,
			Message: "internal server error",
		}
		_ = c.JSON(http.StatusInternalServerError, res)
	}
}
