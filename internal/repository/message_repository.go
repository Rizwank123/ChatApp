package repository

import (
	"context"
	"log"

	"github.com/gofrs/uuid/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/chatApp/internal/domain"
)

type pgxMessageRepository struct {
	db *pgxpool.Pool
}

func NewPgxMessageRepository(db *pgxpool.Pool) domain.MessageRepository {
	return &pgxMessageRepository{db: db}
}

// FindAll implements domain.MessageRepository.
func (r *pgxMessageRepository) FindAll(ctx context.Context) (result []domain.Message, err error) {
	if ctx == nil {
		ctx = context.Background()
	}
	txVal := ctx.Value(TxKey)
	q := `SELECT *  FROM messages WHERE  deleted_at IS NULL`
	var rows pgx.Rows
	if txVal != nil {
		tx := txVal.(pgx.Tx)
		rows, err = tx.Query(ctx, q)
	} else {
		rows, err = r.db.Query(ctx, q)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result, err = pgx.CollectRows(rows, pgx.RowToStructByNameLax[domain.Message])
	if err != nil {
		return result, err
	}
	return result, nil

}

// FindByID implements domain.MessageRepository.
func (r *pgxMessageRepository) FindByID(ctx context.Context, id uuid.UUID) (result domain.Message, err error) {
	if ctx == nil {
		ctx = context.Background()
	}
	txVal := ctx.Value(TxKey)
	q := `SELECT * FROM messages WHERE id = $1 AND deleted_at IS NULL`
	args := []interface{}{id}
	var row pgx.Rows
	if txVal != nil {
		tx := txVal.(pgx.Tx)
		row, err = tx.Query(ctx, q, args...)
	} else {
		row, err = r.db.Query(ctx, q, args...)
	}
	if err != nil {
		return result, err
	}
	defer row.Close()
	result, err = pgx.CollectOneRow(row, pgx.RowToStructByNameLax[domain.Message])
	if err != nil {
		return result, err
	}
	return result, nil

}

// FindByReceiverID implements domain.MessageRepository.
func (r *pgxMessageRepository) FindByReceiverID(ctx context.Context, receiver_id uuid.UUID) (result []domain.Message, err error) {
	if ctx == nil {
		ctx = context.Background()
	}
	txVal := ctx.Value(TxKey)
	q := `SELECT * FROM messages WHERE receiver_id = $1 AND deleted_at IS NULL`
	args := []interface{}{receiver_id}
	var rows pgx.Rows
	if txVal != nil {
		tx := txVal.(pgx.Tx)
		rows, err = tx.Query(ctx, q, args...)
	} else {
		rows, err = r.db.Query(ctx, q, args...)
	}
	if err != nil {
		return result, err
	}
	defer rows.Close()
	result, err = pgx.CollectRows(rows, pgx.RowToStructByNameLax[domain.Message])
	if err != nil {
		return result, err
	}
	return result, nil
}

// FindBySenderID implements domain.MessageRepository.
func (r *pgxMessageRepository) FindBySenderID(ctx context.Context, sender_id uuid.UUID) (result []domain.Message, err error) {
	if ctx == nil {
		ctx = context.Background()
	}
	txVal := ctx.Value(TxKey)
	q := `SELECT * FROM messages WHERE receiver_id = $1 AND deleted_at IS NULL`
	args := []interface{}{sender_id}
	var rows pgx.Rows
	if txVal != nil {
		tx := txVal.(pgx.Tx)
		rows, err = tx.Query(ctx, q, args...)
	} else {
		rows, err = r.db.Query(ctx, q, args...)
	}
	if err != nil {
		return result, err
	}
	defer rows.Close()
	result, err = pgx.CollectRows(rows, pgx.RowToStructByNameLax[domain.Message])
	if err != nil {
		return result, err
	}
	return result, nil
}

// Create implements domain.MessageRepository.
func (r *pgxMessageRepository) Create(ctx context.Context, entity *domain.Message) (err error) {
	if ctx == nil {
		ctx = context.Background()
	}
	txVal := ctx.Value(TxKey)
	q := `INSERT INTO messages (sender_id, receiver_id, message) VALUES ($1, $2, $3) RETURNING id, created_at, updated_at`
	args := []interface{}{entity.SenderID, entity.ReceiverID, entity.Content}

	if txVal != nil {
		tx := txVal.(pgx.Tx)
		err = tx.QueryRow(ctx, q, args...).Scan(&entity.ID, &entity.CreatedAt, &entity.UpdatedAt)
	} else {
		err = r.db.QueryRow(ctx, q, args...).Scan(&entity.ID, &entity.CreatedAt, &entity.UpdatedAt)
	}
	return err

}

// CreateMessageStatus implements domain.MessageRepository.
func (r *pgxMessageRepository) CreateMessageStatus(ctx context.Context, entity *domain.MessageStatus) (err error) {
	if ctx == nil {
		ctx = context.Background()
	}
	txVal := ctx.Value(TxKey)
	q := `INSERT INTO message_statuses (message_id, status) VALUES ($1, $2) RETURNING UpdatedAt`
	args := []interface{}{entity.MessageID, entity.Mstatus}
	if txVal != nil {
		tx := txVal.(pgx.Tx)
		err = tx.QueryRow(ctx, q, args...).Scan(&entity.UpdatedAt)
	} else {
		err = r.db.QueryRow(ctx, q, args...).Scan(&entity.UpdatedAt)
	}
	return err

}

// CreateMultiple implements domain.MessageRepository.
func (r *pgxMessageRepository) CreateMultiple(ctx context.Context, entities []*domain.Message) (err error) {
	if ctx == nil {
		ctx = context.Background()
	}
	txVal := ctx.Value(TxKey)
	b := &pgx.Batch{}
	for _, entity := range entities {
		q := `INSERT INTO messages (sender_id, receiver_id, message) VALUES ($1,, $2, $3) RETURNING  id, created_at, updated_at`
		args := []interface{}{entity.SenderID, entity.ReceiverID, entity.Content}
		b.Queue(q, args...)
	}
	var br pgx.BatchResults
	if txVal != nil {
		tx := txVal.(pgx.Tx)
		br = tx.SendBatch(ctx, b)
	} else {
		br = r.db.SendBatch(ctx, b)
	}
	defer func(br pgx.BatchResults) {
		err := br.Close()
		if err != nil {
			log.Println("Error closing batch results", err)
		}
	}(br)
	for idx := range entities {
		err = br.QueryRow().Scan(&entities[idx].ID, &entities[idx].CreatedAt, &entities[idx].UpdatedAt)
		if err != nil {
			return err
		}
	}

	return err
}

// Update implements domain.MessageRepository.
func (r *pgxMessageRepository) Update(ctx context.Context, entity *domain.Message) (err error) {
	if ctx == nil {
		ctx = context.Background()
	}
	txVal := ctx.Value(ctx)
	q := `UPDATE messages SET sender_id = $1, receiver_id = $2, message = $3 WHERE id = $4  RETURNING updated_at`
	args := []interface{}{entity.SenderID, entity.ReceiverID, entity.Content, entity.ID}
	if txVal != nil {
		tx := txVal.(pgx.Tx)
		err = tx.QueryRow(ctx, q, args...).Scan(&entity.UpdatedAt)
	} else {
		err = r.db.QueryRow(ctx, q, args...).Scan(&entity.UpdatedAt)
	}
	return err
}

// UpdateMessageStatus implements domain.MessageRepository.
func (r *pgxMessageRepository) UpdateMessageStatus(ctx context.Context, entity *domain.MessageStatus) (err error) {
	if ctx == nil {
		ctx = context.Background()
	}
	txVal := ctx.Value(TxKey)
	q := `UPDATE message_statuses SET status = $1 WHERE message_id = $2 RETURNING  updated_at`
	args := []interface{}{entity.Mstatus, entity.MessageID}
	if txVal != nil {
		tx := txVal.(pgx.Tx)
		err = tx.QueryRow(ctx, q, args...).Scan(&entity.UpdatedAt)
	} else {
		err = r.db.QueryRow(ctx, q, args...).Scan(&entity.UpdatedAt)
	}
	return err
}

// UpdateMultiple implements domain.MessageRepository.
func (r *pgxMessageRepository) UpdateMultiple(ctx context.Context, entities []*domain.Message) (err error) {
	if ctx == nil {
		ctx = context.Background()
	}
	txVal := ctx.Value(TxKey)

	b := &pgx.Batch{}
	for _, entity := range entities {
		q := `UPDATE messages SET sender_id = $1, receiver_id = $2, message = $3  WHERE id = $4 `
		args := []interface{}{entity.SenderID, entity.ReceiverID, entity.Content, entity.ID}
		b.Queue(q, args...)
	}
	// Send the batch
	if txVal != nil {
		tx := txVal.(pgx.Tx)
		err = tx.SendBatch(ctx, b).Close()
	} else {
		err = r.db.SendBatch(ctx, b).Close()
	}

	return err
}

// Delete implements domain.MessageRepository.
func (r *pgxMessageRepository) Delete(ctx context.Context, id uuid.UUID) (err error) {
	if ctx == nil {
		ctx = context.Background()
	}
	txVal := ctx.Value(TxKey)

	q := `UPDATE messages set deleted_at = NOW() Where id = $1`
	args := []interface{}{id}
	if txVal != nil {
		tx := txVal.(pgx.Tx)
		_, err = tx.Exec(ctx, q, args...)
	} else {
		_, err = r.db.Exec(ctx, q, args...)
	}
	return err
}
