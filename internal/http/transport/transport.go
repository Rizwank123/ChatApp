package transport

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/go-playground/validator"
	"github.com/labstack/echo/v4"
	"github.com/leebenson/conform"

	"github.com/chatApp/internal/domain"
)

const PageMax = 500

// DecodeQueryOptions decodes the query options
func DecodeQueryOptions(ctx echo.Context) domain.QueryOptions {
	limit, offset := GetLimitAndOffset(ctx)
	selectFields := ctx.QueryParam("fields")
	associationsString := ctx.QueryParam("associations")
	var associations []domain.QueryAssociation
	if associationsString != "" {
		entities := strings.Split(associationsString, "|")
		for _, entity := range entities {
			fields := strings.Split(entity, ":")
			if len(fields) < 2 {
				fields = append(fields, "*")
			}
			associations = append(associations, domain.QueryAssociation{Name: fields[0], SelectFields: fields[1]})
		}
	}
	q := domain.QueryOptions{}
	if limit != 0 {
		q.Limit = limit
	}
	if offset != 0 {
		q.Offset = offset
	}

	if selectFields != "" {
		q.SelectFields = selectFields
	}

	if associations != nil {
		q.Associations = associations
	}

	return q
}

// DecodeAndValidateRequestBody decodes and validates the request body
func DecodeAndValidateRequestBody(ctx echo.Context, t interface{}) error {
	err := ctx.Bind(t)
	if err != nil {
		return err
	}
	err = conform.Strings(t)
	if err != nil {
		return err
	}
	err = ctx.Validate(t)
	if err != nil {
		return err
	}
	return nil
}

// GetLimitAndOffset gets the limit and offset from the query params
func GetLimitAndOffset(ctx echo.Context) (int64, int64) {
	p := ctx.QueryParam("page")
	s := ctx.QueryParam("size")
	page, _ := strconv.Atoi(p)
	if page < 0 {
		page = 0
	}

	size, _ := strconv.Atoi(s)
	switch {
	case size > PageMax:
		size = PageMax
	case size <= 0:
		size = PageMax
	}

	return int64(size), int64(page * size)
}

// SendResponse sends a response
func SendResponse(ctx echo.Context, status int, data interface{}) error {
	var finalResult domain.BaseResponse
	if data != nil {
		finalResult = domain.BaseResponse{
			Data: data,
		}
	}
	if status == http.StatusNoContent {
		return ctx.NoContent(status)
	}
	return ctx.JSON(status, finalResult)
}

// SendPaginationResponse sends a paginated response
func SendPaginationResponse(ctx echo.Context, status int, data interface{}, total int64) error {
	p, _ := strconv.Atoi(ctx.QueryParam("page"))
	s, _ := strconv.Atoi(ctx.QueryParam("size"))

	page := int64(p)
	if page <= 0 {
		page = 0
	}

	size := int64(s)
	if size <= 0 {
		size = PageMax
	}

	finalResult := domain.PaginationResponse{
		Data:  data,
		Page:  page,
		Size:  size,
		Total: total,
	}
	if status == http.StatusNoContent {
		return ctx.NoContent(status)
	}
	return ctx.JSON(status, finalResult)
}

// CustomValidator custom validator for echo
type CustomValidator struct {
	Validator *validator.Validate
}

// Validate validates the request body
func (cv *CustomValidator) Validate(i interface{}) error {
	if err := cv.Validator.Struct(i); err != nil {
		return err
	}
	return nil
}
