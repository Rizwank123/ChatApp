package domain

import (
	"context"
	"database/sql/driver"
	"encoding/json"
	"time"

	"github.com/gofrs/uuid/v5"
)

type (
	// JSONB represents a JSONB type
	JSONB map[string]interface{} // @name JSONB
)

type (
	// Base defines the mode for Base
	Base struct {
		ID uuid.UUID `db:"id" json:"id,omitempty" example:""`
	}

	// BaseAudit defines the model for base audit
	BaseAudit struct {
		CratedAt  time.Time `db:"created_at" json:"created_at,omitempty"`
		UpdatedAt time.Time `db:"updated_at" json:"updated_at,omitempty"`
		DeletedAt time.Time `db:"deleted_at" json:"deleted_at,omitempty"`
	}
)

type (
	// BaseResponse defines base response fields.
	BaseResponse struct {
		Data interface{} `json:"data"`
	} // @name BaseResponse

	// PaginationResponse defines pagination response fields.
	PaginationResponse struct {
		Data  interface{} `json:"data"`
		Total int64       `json:"total" example:"1000"`
		Size  int64       `json:"size" example:"10"`
		Page  int64       `json:"page" example:"1"`
	} // @name PaginationResponse
)

type (
	FilterOp string // @name FilterOp

	// QueryAssociation defines the associations of an entity
	QueryAssociation struct {
		// Name represents the name of the association
		Name         string `json:"name" example:"organization"`
		SelectFields string `json:"selectFields" example:"id,name"`
	} // @name QueryAssociation

	// QueryOptions defines the options for a query
	QueryOptions struct {
		Limit        int64              `json:"limit" example:"10"`
		Offset       int64              `json:"offset" example:"0"`
		SelectFields string             `json:"selectFields" example:"id,name"`
		Associations []QueryAssociation `json:"associations"`
	} // @name QueryOptions

	// FilterFieldPredicate defines the predicate for filters
	FilterFieldPredicate struct {
		// Field represents a column for the entity you are filtering
		Field string `json:"field" example:"name"`
		// Operator represents the filter operation you'd like to perform on the field
		Operator FilterOp `json:"operator" enums:"eq,neq,gt,gte,lt,lte,in,nin,like,nlike,ilike,nilike,isnull,notnull,between" example:"eq"`
		// Value represents the value you'd like to filter by
		Value interface{} `json:"value"`
	} // @name FilterFieldPredicate

	// SortDirection defines the sort direction
	SortDirection string // @name SortDirection

	// SortKey defines the sort key for sorting
	SortKey struct {
		// Field represents a column for the entity you are sorting
		Field string `json:"field" example:"name"`
		// Direction represents the direction of the sort
		Direction string `json:"direction" example:"asc"`
	} // @name SortKey

	// FilterInput defines the input for filtering
	FilterInput struct {
		OrganizationID uuid.UUID `json:"-"`
		// Fields represents the fields you want to filter
		Fields []FilterFieldPredicate `json:"fields"`
		// SortKeys represents the sort keys you want to sort by
		SortKeys []SortKey `json:"sort_keys"`
	} // @name FilterInput
)

type (
	// Transactioner defines the methods that any transactioner should implement.
	Transactioner interface {
		Begin(ctx context.Context) (result context.Context, err error)
		Commit(ctx context.Context) (err error)
		Rollback(ctx context.Context, err error)
	}
)

// Value implements the driver.Valuer interface,
func (j *JSONB) Value() (driver.Value, error) {
	valueString, err := json.Marshal(j)
	return string(valueString), err
}

// Scan implements the sql.Scanner interface,
func (j *JSONB) Scan(value interface{}) error {
	if err := json.Unmarshal([]byte(value.(string)), &j); err != nil {
		return err
	}
	return nil
}

const (
	FilterOpEq      FilterOp = "eq"
	FilterOpNeq     FilterOp = "neq"
	FilterOpGt      FilterOp = "gt"
	FilterOpGte     FilterOp = "gte"
	FilterOpLt      FilterOp = "lt"
	FilterOpLte     FilterOp = "lte"
	FilterOpIn      FilterOp = "in"
	FilterOpNin     FilterOp = "nin"
	FilterOpLike    FilterOp = "like"
	FilterOpNlike   FilterOp = "nlike"
	FilterOpIlike   FilterOp = "ilike"
	FilterOpNilike  FilterOp = "nilike"
	FilterOpIsnull  FilterOp = "isnull"
	FilterOpNotnull FilterOp = "notnull"
	FilterOpBetween FilterOp = "between"
)

// GetTableNameForEntity returns the table name for the entity
func GetTableNameForEntity(ent interface{}) (result string) {
	switch ent.(type) {
	case User:
		return "users"
	default:
		return "Invalid  entity"
	}
}
