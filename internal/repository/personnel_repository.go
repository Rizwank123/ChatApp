package repository

import (
	"context"

	"github.com/gofrs/uuid/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/chatApp/internal/domain"
)

type pgxPersonnelRepository struct {
	db *pgxpool.Pool
}

func NewPersonnelRepository(db *pgxpool.Pool) domain.PersonnelRepository {
	return &pgxPersonnelRepository{db: db}
}

// Filter implements domain.PersonnelRepository.
func (r *pgxPersonnelRepository) Filter(ctx context.Context, in domain.FilterInput, options domain.QueryOptions) (result []domain.Personnel, total int64, err error) {
	if ctx == nil {
		ctx = context.Background()
	}
	txVal := ctx.Value(TxKey)

	// Retrieve the record count
	cq := `SELECT COUNT(*) FROM personnel WHERE deleted_at IS NULL`
	cq, cargs := buildQueryForFilter(in, cq)
	if txVal != nil {
		tx := txVal.(pgx.Tx)
		err = tx.QueryRow(ctx, cq, cargs...).Scan(&total)
	} else {
		err = r.db.QueryRow(ctx, cq, cargs...).Scan(&total)
	}
	if err != nil {
		return result, total, err
	}

	// Retrieve the data
	q := `SELECT * FROM personnel WHERE deleted_at IS NULL`
	q, args := buildQueryForFilter(in, q)
	q = buildSortKeysForFilter(in, q)
	q = applyLimitAndOffset(q, options)
	q = buildSelectorForQuery(q, options)
	var rows pgx.Rows
	if txVal != nil {
		tx := txVal.(pgx.Tx)
		rows, err = tx.Query(ctx, q, args...)
	} else {
		rows, err = r.db.Query(ctx, q, args...)
	}
	defer rows.Close()

	if err != nil {
		return result, total, err
	}
	if rows == nil {
		return result, total, nil
	}

	// Collect the results from the rows
	result, err = pgx.CollectRows(rows, pgx.RowToStructByNameLax[domain.Personnel])
	return result, total, err
}

// FindByID implements domain.PersonnelRepository.
func (r *pgxPersonnelRepository) FindByID(ctx context.Context, id uuid.UUID) (result domain.Personnel, err error) {
	if ctx == nil {
		ctx = context.Background()
	}
	txVal := ctx.Value(TxKey)
	q := `SELECT * FROM  personnel WHERE id = $1 AND deleted_at IS NULL`
	args := []interface{}{id}
	var rows pgx.Rows
	if txVal != nil {
		tx := txVal.(pgx.Tx)
		rows, err = tx.Query(ctx, q, args...)
	} else {
		rows, err = r.db.Query(ctx, q, args...)
	}
	defer rows.Close()
	if err != nil {
		return result, err
	}
	if rows == nil {
		return result, nil
	}
	// Collect the results from the rows
	result, err = pgx.CollectOneRow(rows, pgx.RowToStructByNameLax[domain.Personnel])
	return result, err

}

// FindByUserID implements domain.PersonnelRepository.
func (r *pgxPersonnelRepository) FindByUserID(ctx context.Context, id uuid.UUID) (result domain.Personnel, err error) {
	if ctx == nil {
		ctx = context.Background()
	}
	txVal := ctx.Value(TxKey)
	q := `SELECT * FROM  personnel WHERE user_id = $1 AND deleted_at IS NULL`
	args := []interface{}{id}
	var rows pgx.Rows
	if txVal != nil {
		tx := txVal.(pgx.Tx)
		rows, err = tx.Query(ctx, q, args...)
	} else {
		rows, err = r.db.Query(ctx, q, args...)
	}
	defer rows.Close()
	if err != nil {
		return result, err
	}
	if rows == nil {
		return result, nil
	}

	// Collect the results from the rows
	result, err = pgx.CollectOneRow(rows, pgx.RowToStructByNameLax[domain.Personnel])
	return result, err

}

// Create implements domain.PersonnelRepository.
func (r *pgxPersonnelRepository) Create(ctx context.Context, entity *domain.Personnel) (err error) {
	// check if context has transaction
	if ctx == nil {
		ctx = context.Background()
	}
	txVal := ctx.Value(TxKey)

	q := `INSERT INTO personnel(first_name, last_name, gender, email, mobile, address, role, user_id, activation_status, avatar) values($1, $2, $3, $4, $5, $6, $7, $8,$9, $10) RETURNING id, created_at, updated_at`
	args := []interface{}{entity.FirstName, entity.LastName, entity.Gender, entity.Email, entity.Mobile, entity.Address, entity.Role, entity.UserID, entity.ActivationStatus, entity.Avatar}
	if txVal != nil {
		tx := txVal.(pgx.Tx)
		err = tx.QueryRow(ctx, q, args...).Scan(&entity.ID, &entity.CreatedAt, &entity.UpdatedAt)
	} else {
		err = r.db.QueryRow(ctx, q, args...).Scan(&entity.ID, &entity.CreatedAt, &entity.UpdatedAt)
	}
	return err
}

// Delete implements domain.PersonnelRepository.
func (r *pgxPersonnelRepository) Delete(ctx context.Context, id uuid.UUID) (err error) {
	if ctx == nil {
		ctx = context.Background()
	}
	txVal := ctx.Value(TxKey)
	q := `UPDATE  personnel SET deleted_at = now() WHERE id = $1`
	args := []interface{}{id}
	if txVal != nil {
		tx := txVal.(pgx.Tx)
		_, err = tx.Exec(ctx, q, args...)
	} else {
		_, err = r.db.Exec(ctx, q, args...)
	}
	return err
}

// Update implements domain.PersonnelRepository.
func (r *pgxPersonnelRepository) Update(ctx context.Context, entity *domain.Personnel) (err error) {
	if ctx == nil {
		ctx = context.Background()
	}
	txVal := ctx.Value(TxKey)

	q := `UPDATE personnel SET first_name = $1, last_name = $2, gender = $3, email = $4, mobile = $5, address = $6, role = $7, user_id = $8, activation_status = $9, avatar = $10 WHERE id = $11 RETURNING updated_at`
	args := []interface{}{entity.FirstName, entity.LastName, entity.Gender, entity.Email, entity.Mobile, entity.Address, entity.Role, entity.UserID, entity.ActivationStatus, entity.Avatar, entity.ID}
	if txVal != nil {
		tx := txVal.(pgx.Tx)
		err = tx.QueryRow(ctx, q, args...).Scan(&entity.UpdatedAt)
	} else {
		err = r.db.QueryRow(ctx, q, args...).Scan(&entity.UpdatedAt)
	}
	return err
}
