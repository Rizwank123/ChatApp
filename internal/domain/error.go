package domain

import "fmt"

// NotFoundError defines model for not found error.
type NotFoundError struct{} // @name NotFoundError

func (e NotFoundError) Error() string {
	return "The resource you are looking for does not exist"
}

// InvalidRequestError defines model for invalid request error.
type InvalidRequestError struct {
	Message string `json:"message" example:"invalid request"`
} // @name InvalidRequestError

func (e InvalidRequestError) Error() string {
	return e.Message
}

// UnauthorizedError defines model for unauthorized error.
type UnauthorizedError struct {
	Code    string `json:"code" example:"UNAUTHORIZED"`
	Message string `json:"message" example:"You are not authorized to access this resource"`
} // @name UnauthorizedError

func (e UnauthorizedError) Error() string {
	return e.Message
}

// ForbiddenAccessError defines model for forbidden access error.
type ForbiddenAccessError struct {
	Code    string `json:"code" example:"FORBIDDEN_ACCESS"`
	Message string `json:"message" example:"You are forbidden from accessing this resource"`
} // @name ForbiddenAccessError

func (e ForbiddenAccessError) Error() string {
	return e.Message
}

// ValidationError defines model for validation error.
type ValidationError struct {
	Code    string   `json:"code" example:"VALIDATION_ERROR"`
	Message string   `json:"message" example:"Not a valid mobile number"`
	Fields  []string `json:"fields" example:"mobile_number is required"`
} // @name ValidationError

func (e ValidationError) Error() string {
	if len(e.Fields) > 0 {
		return fmt.Sprintf(e.Message, e.Fields)
	}
	return e.Message
}

// UserError defines model for user error.
type UserError struct {
	Code    string `json:"code" example:"INVALID_REQUEST"`
	Message string `json:"message" example:"Oops! Something went wrong. Please try again later"`
} // @name UserError

func (e UserError) Error() string {
	return e.Message
}

// DataNotFoundError defines model for data not found error.
type DataNotFoundError struct{} // @name DataNotFoundError

func (e DataNotFoundError) Error() string {
	return "The record you are looking for does not exist"
}

// SystemError defines model for system error.
type SystemError struct {
	Code    string `json:"code" example:"INTERNAL_SERVER_ERROR"`
	Message string `json:"message" example:"Oops! Something went wrong. Please try again later"`
} // @name SystemError

func (e SystemError) Error() string {
	return e.Message
}

const (
	ErrorCodeINVALID_REQUEST       = "INVALID_REQUEST"
	ErrorCodeVALIDATION_ERROR      = "VALIDATION_ERROR"
	ErrorCodeINTERNAL_SERVER_ERROR = "INTERNAL_SERVER_ERROR"
	ErrorCodeUNAUTHORIZED          = "UNAUTHORIZED"
	ErrorCodeFORBIDDEN_ACCESS      = "FORBIDDEN_ACCESS"

	ErrorCodeMOBILE_NUMBER_EXISTS   = "MOBILE_NUMBER_EXISTS"
	ErrorCodePERSONNEL_NAME_EXISTS  = "PERSONNEL_NAME_EXISTS"
	ErrorCodeCHECK_IN_OUT_OVERLAP   = "CHECK_IN_OUT_OVERLAP"
	ErrorCodeWORK_ITEM_ALREADY_PAID = "WORK_ITEM_ALREADY_PAID"
)

const (
	MessageVALIDATIONFAILED    = "Validation failed for some or all of the fields in the request"
	MessageMOBILENUMBEREXISTS  = "User with this mobile number already exists"
	MessagePERSONNELNAMEEXISTS = "User with this name is already registered in the system"

	MessageUNAUTHORIZEDACCESS = "You are not authorized to access this resource"
	MessageFORBIDDENACCESS    = "You are forbidden from accessing this resource"
)
