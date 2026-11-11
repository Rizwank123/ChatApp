package repository

import (
	"context"

	"github.com/gofrs/uuid/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Rizwank123/chatApp/internal/domain"
)

type pgxOutboxRepository struct {
	db *pgxpool.Pool
}

// NewOutboxRepository returns a new instance of OutboxRepository.
func NewOutboxRepository(db *pgxpool.Pool) domain.OutboxRepository {
	return &pgxOutboxRepository{db: db}
}

func (r *pgxOutboxRepository) CreateEvent(ctx context.Context, event *domain.OutboxEvent) error {
	q := `INSERT INTO outbox_events (aggregate_type, aggregate_id, event_type, payload, status)
	      VALUES ($1, $2, $3, $4, 'PENDING')
	      RETURNING id, created_at`

	// Support transaction propagation
	if txVal := ctx.Value(TxKey); txVal != nil {
		if tx, ok := txVal.(pgx.Tx); ok {
			return tx.QueryRow(ctx, q, event.AggregateType, event.AggregateID, event.EventType, event.Payload).Scan(&event.ID, &event.CreatedAt)
		}
	}

	return r.db.QueryRow(ctx, q, event.AggregateType, event.AggregateID, event.EventType, event.Payload).Scan(&event.ID, &event.CreatedAt)
}

func (r *pgxOutboxRepository) FetchPendingEvents(ctx context.Context, batchSize int) ([]domain.OutboxEvent, error) {
	q := `SELECT id, aggregate_type, aggregate_id, event_type, payload, status, retry_count, last_error, created_at, published_at
	      FROM outbox_events
	      WHERE status = 'PENDING'
	      ORDER BY created_at ASC
	      LIMIT $1`
	rows, err := r.db.Query(ctx, q, batchSize)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []domain.OutboxEvent
	for rows.Next() {
		var e domain.OutboxEvent
		if err := rows.Scan(
			&e.ID, &e.AggregateType, &e.AggregateID, &e.EventType, &e.Payload,
			&e.Status, &e.RetryCount, &e.LastError, &e.CreatedAt, &e.PublishedAt,
		); err != nil {
			return nil, err
		}
		events = append(events, e)
	}
	return events, nil
}

func (r *pgxOutboxRepository) MarkEventPublished(ctx context.Context, id uuid.UUID) error {
	q := `UPDATE outbox_events SET status = 'PUBLISHED', published_at = NOW() WHERE id = $1`
	_, err := r.db.Exec(ctx, q, id)
	return err
}

func (r *pgxOutboxRepository) MarkEventFailed(ctx context.Context, id uuid.UUID, errMessage string) error {
	q := `UPDATE outbox_events SET status = 'FAILED', retry_count = retry_count + 1, last_error = $2 WHERE id = $1`
	_, err := r.db.Exec(ctx, q, id, errMessage)
	return err
}
