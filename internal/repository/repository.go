package repository

import (
	"context"
	"fmt"
	"strings"

	"github.com/gofrs/uuid/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/chatApp/internal/domain"
)

func buildQueryForFilter(in domain.FilterInput, q string) (res string, args []interface{}) {
	args = make([]interface{}, 0)
	for idx, f := range in.Fields {
		argVal := f.Value
		switch f.Operator {
		case domain.FilterOpEq:
			q += fmt.Sprintf(" AND %s = $%d", f.Field, idx+1)
		case domain.FilterOpNeq:
			q += fmt.Sprintf(" AND %s != $%d", f.Field, idx+1)
		case domain.FilterOpGt:
			q += fmt.Sprintf(" AND %s > $%d", f.Field, idx+1)
		case domain.FilterOpGte:
			q += fmt.Sprintf(" AND %s >= $%d", f.Field, idx+1)
		case domain.FilterOpLt:
			q += fmt.Sprintf(" AND %s < $%d", f.Field, idx+1)
		case domain.FilterOpLte:
			q += fmt.Sprintf(" AND %s <= $%d", f.Field, idx+1)
		case domain.FilterOpLike:
			q += fmt.Sprintf(" AND %s LIKE $%d", f.Field, idx+1)
			argVal = fmt.Sprintf("%%%s%%", strings.ToLower(argVal.(string)))
		case domain.FilterOpNlike:
			q += fmt.Sprintf(" AND %s NOT LIKE $%d", f.Field, idx+1)
			argVal = fmt.Sprintf("%%%s%%", strings.ToLower(argVal.(string)))
		case domain.FilterOpIlike:
			q += fmt.Sprintf(" AND %s ILIKE $%d", f.Field, idx+1)
			argVal = fmt.Sprintf("%%%s%%", strings.ToLower(argVal.(string)))
		case domain.FilterOpNilike:
			q += fmt.Sprintf(" AND %s NOT ILIKE $%d", f.Field, idx+1)
			argVal = fmt.Sprintf("%%%s%%", strings.ToLower(argVal.(string)))
		case domain.FilterOpIn:
			q += fmt.Sprintf(" AND %s = ANY($%d)", f.Field, idx+1)
		case domain.FilterOpNin:
			q += fmt.Sprintf(" AND %s NOT IN ($%d)", f.Field, idx+1)
		case domain.FilterOpIsnull:
			q += fmt.Sprintf(" AND %s IS NULL", f.Field)
		case domain.FilterOpNotnull:
			q += fmt.Sprintf(" AND %s IS NOT NULL", f.Field)
		case domain.FilterOpBetween:
			q += fmt.Sprintf(" AND %s BETWEEN $%d AND $%d", f.Field, idx+1, idx+2)
		}
		args = append(args, argVal)
	}
	return q, args
}

func buildSortKeysForFilter(in domain.FilterInput, q string) string {
	for _, s := range in.SortKeys {
		q += fmt.Sprintf(" ORDER BY %s %s", s.Field, s.Direction)
	}
	return q
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

func buildSelectorForQuery(q string, opts domain.QueryOptions) (result string) {
	if opts.SelectFields == "" {
		opts.SelectFields = "*"
	}
	result = strings.Replace(q, "SELECT * FROM", strings.Join([]string{"SELECT", opts.SelectFields, "FROM"}, " "), 1)
	return result
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
	TblUser = "user"
	TblPersonnel = "personnel"
	
)