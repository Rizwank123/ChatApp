package repository

import (
	"context"

	"github.com/gofrs/uuid/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Rizwank123/chatApp/internal/domain"
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
	cq, cargs, err := buildQueryForFilter(in, cq)
	if err != nil {
		return nil, 0, err
	}
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
	q, args, err := buildQueryForFilter(in, q)
	if err != nil {
		return nil, total, err
	}
	q, err = buildSortKeysForFilter(in, q)
	if err != nil {
		return nil, total, err
	}
	q = applyLimitAndOffset(q, options)
	q, err = buildSelectorForQuery(q, options)
	if err != nil {
		return nil, total, err
	}
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
func (r *pgxPersonnelRepository) FindByUserIDs(ctx context.Context, ids []uuid.UUID) (result []domain.Personnel, err error) {
	if ctx == nil {
		ctx = context.Background()
	}
	q := `SELECT * FROM  personnel WHERE user_id = ANY($1) AND deleted_at IS Null`
	args := []interface{}{ids}
	var rows pgx.Rows
	if txVal := ctx.Value(TxKey); txVal != nil {
		tx := txVal.(pgx.Tx)
		rows, err = tx.Query(ctx, q, args...)
	} else {
		rows, err = r.db.Query(ctx, q, args...)
	}
	defer rows.Close()
	if err != nil {
		return result, err
	}
	result, err = pgx.CollectRows(rows, pgx.RowToStructByNameLax[domain.Personnel])
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

	q := `UPDATE personnel SET first_name = $1, last_name = $2, display_name = $3, gender = $4, email = $5, mobile = $6, address = $7, avatar = $8, bio = $9, website = $10, is_private = $11, updated_at = NOW() WHERE id = $12 AND deleted_at IS NULL RETURNING updated_at`
	args := []interface{}{entity.FirstName, entity.LastName, entity.DisplayName, entity.Gender, entity.Email, entity.Mobile, entity.Address, entity.Avatar, entity.Bio, entity.Website, entity.IsPrivate, entity.ID}
	if txVal != nil {
		tx := txVal.(pgx.Tx)
		err = tx.QueryRow(ctx, q, args...).Scan(&entity.UpdatedAt)
	} else {
		err = r.db.QueryRow(ctx, q, args...).Scan(&entity.UpdatedAt)
	}
	return err
}
