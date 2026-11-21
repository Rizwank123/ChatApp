package repository

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/gofrs/uuid/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Rizwank123/chatApp/internal/domain"
)

var safeSQLIdentifier = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)

func buildQueryForFilter(in domain.FilterInput, q string) (res string, args []interface{}, err error) {
	args = make([]interface{}, 0, len(in.Fields))
	for _, f := range in.Fields {
		if !safeSQLIdentifier.MatchString(f.Field) {
			return "", nil, errors.New("invalid filter field")
		}
		argVal := f.Value
		placeholder := func(value interface{}) string {
			args = append(args, value)
			return fmt.Sprintf("$%d", len(args))
		}
		switch f.Operator {
		case domain.FilterOpEq:
			q += fmt.Sprintf(" AND %s = %s", f.Field, placeholder(argVal))
		case domain.FilterOpNeq:
			q += fmt.Sprintf(" AND %s != %s", f.Field, placeholder(argVal))
		case domain.FilterOpGt:
			q += fmt.Sprintf(" AND %s > %s", f.Field, placeholder(argVal))
		case domain.FilterOpGte:
			q += fmt.Sprintf(" AND %s >= %s", f.Field, placeholder(argVal))
		case domain.FilterOpLt:
			q += fmt.Sprintf(" AND %s < %s", f.Field, placeholder(argVal))
		case domain.FilterOpLte:
			q += fmt.Sprintf(" AND %s <= %s", f.Field, placeholder(argVal))
		case domain.FilterOpLike:
			value, ok := argVal.(string)
			if !ok {
				return "", nil, errors.New("like filter requires a string value")
			}
			argVal = fmt.Sprintf("%%%s%%", strings.ToLower(value))
			q += fmt.Sprintf(" AND %s LIKE %s", f.Field, placeholder(argVal))
		case domain.FilterOpNlike:
			value, ok := argVal.(string)
			if !ok {
				return "", nil, errors.New("not-like filter requires a string value")
			}
			argVal = fmt.Sprintf("%%%s%%", strings.ToLower(value))
			q += fmt.Sprintf(" AND %s NOT LIKE %s", f.Field, placeholder(argVal))
		case domain.FilterOpIlike:
			value, ok := argVal.(string)
			if !ok {
				return "", nil, errors.New("ilike filter requires a string value")
			}
			argVal = fmt.Sprintf("%%%s%%", strings.ToLower(value))
			q += fmt.Sprintf(" AND %s ILIKE %s", f.Field, placeholder(argVal))
		case domain.FilterOpNilike:
			value, ok := argVal.(string)
			if !ok {
				return "", nil, errors.New("not-ilike filter requires a string value")
			}
			argVal = fmt.Sprintf("%%%s%%", strings.ToLower(value))
			q += fmt.Sprintf(" AND %s NOT ILIKE %s", f.Field, placeholder(argVal))
		case domain.FilterOpIn:
			q += fmt.Sprintf(" AND %s = ANY(%s)", f.Field, placeholder(argVal))
		case domain.FilterOpNin:
			q += fmt.Sprintf(" AND %s <> ALL(%s)", f.Field, placeholder(argVal))
		case domain.FilterOpIsnull:
			q += fmt.Sprintf(" AND %s IS NULL", f.Field)
		case domain.FilterOpNotnull:
			q += fmt.Sprintf(" AND %s IS NOT NULL", f.Field)
		case domain.FilterOpBetween:
			values, ok := argVal.([]interface{})
			if !ok || len(values) != 2 {
				return "", nil, errors.New("between filter requires exactly two values")
			}
			q += fmt.Sprintf(" AND %s BETWEEN %s AND %s", f.Field, placeholder(values[0]), placeholder(values[1]))
		default:
			return "", nil, errors.New("unsupported filter operator")
		}
	}
	return q, args, nil
}

func buildSortKeysForFilter(in domain.FilterInput, q string) (string, error) {
	parts := make([]string, 0, len(in.SortKeys))
	for _, s := range in.SortKeys {
		if !safeSQLIdentifier.MatchString(s.Field) {
			return "", errors.New("invalid sort field")
		}
		direction := strings.ToUpper(strings.TrimSpace(s.Direction))
		if direction != "ASC" && direction != "DESC" {
			return "", errors.New("invalid sort direction")
		}
		parts = append(parts, fmt.Sprintf("%s %s", s.Field, direction))
	}
	if len(parts) > 0 {
		q += " ORDER BY " + strings.Join(parts, ", ")
	}
	return q, nil
}

func applyLimitAndOffset(q string, options domain.QueryOptions) string {
	if options.Limit > 0 {
		q += fmt.Sprintf(" LIMIT %d", options.Limit)
	}
	if options.Offset > 0 {
		q += fmt.Sprintf(" OFFSET %d", options.Offset)
	}
	return q
}

func buildSelectorForQuery(q string, opts domain.QueryOptions) (result string, err error) {
	if opts.SelectFields == "" {
		opts.SelectFields = "*"
	}
	fields := strings.Split(opts.SelectFields, ",")
	for i := range fields {
		fields[i] = strings.TrimSpace(fields[i])
		if fields[i] == "*" && len(fields) == 1 {
			continue
		}
		if !safeSQLIdentifier.MatchString(fields[i]) {
			return "", errors.New("invalid selected field")
		}
	}
	result = strings.Replace(q, "SELECT * FROM", strings.Join([]string{"SELECT", strings.Join(fields, ", "), "FROM"}, " "), 1)
	return result, nil
}

func loadAssociations[T any](ctx context.Context, db *pgxpool.Pool, filterColumn string, fields string, ids []uuid.UUID, ent T, many bool) (result interface{}, err error) {
	txVal := ctx.Value(TxKey)
	if fields == "" {
		fields = "*"
	}
	tblName := domain.GetTableNameForEntity(ent)
	q := fmt.Sprintf("SELECT %s FROM %s WHERE %s = ANY($1) AND deleted_at IS NULL", fields, strings.ToLower(tblName), strings.ToLower(filterColumn))
	args := []interface{}{ids}
	var rows pgx.Rows
	if txVal != nil {
		tx := txVal.(pgx.Tx)
		rows, err = tx.Query(ctx, q, args...)
	} else {
		rows, err = db.Query(ctx, q, args...)
	}
	defer rows.Close()
	if err != nil {
		return result, err
	}
	if rows == nil {
		return result, err
	}

	// Collect the results
	if many {
		result, err = pgx.CollectRows(rows, pgx.RowToStructByNameLax[T])
	} else {
		result, err = pgx.CollectOneRow(rows, pgx.RowToStructByNameLax[T])
	}

	return result, err
}

const (
	TblUser      = "user"
	TblPersonnel = "personnel"
)
