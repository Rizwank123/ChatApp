package repository

import (
	"context"

	"github.com/gofrs/uuid/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Rizwank123/chatApp/internal/domain"
)

type pgxNotificationRepository struct {
	db *pgxpool.Pool
}

// NewNotificationRepository returns a new NotificationRepository instance.
func NewNotificationRepository(db *pgxpool.Pool) domain.NotificationRepository {
	return &pgxNotificationRepository{db: db}
}

func (r *pgxNotificationRepository) Create(ctx context.Context, n *domain.Notification) error {
	q := `INSERT INTO notifications (recipient_id, actor_id, notification_type, entity_type, entity_id, title, message)
	      VALUES ($1, $2, $3, $4, $5, $6, $7)
	      RETURNING id, is_read, created_at`
	return r.db.QueryRow(ctx, q,
		n.RecipientID, n.ActorID, n.NotificationType, n.EntityType, n.EntityID, n.Title, n.Message,
	).Scan(&n.ID, &n.IsRead, &n.CreatedAt)
}

func (r *pgxNotificationRepository) FindUserNotifications(ctx context.Context, recipientID uuid.UUID, limit, offset int) ([]domain.Notification, int64, error) {
	countQ := `SELECT COUNT(*) FROM notifications WHERE recipient_id = $1 AND deleted_at IS NULL`
	var total int64
	if err := r.db.QueryRow(ctx, countQ, recipientID).Scan(&total); err != nil {
		return nil, 0, err
	}

	q := `SELECT id, recipient_id, actor_id, notification_type, entity_type, entity_id, title, message, is_read, read_at, created_at
	      FROM notifications
	      WHERE recipient_id = $1 AND deleted_at IS NULL
	      ORDER BY created_at DESC
	      LIMIT $2 OFFSET $3`
	rows, err := r.db.Query(ctx, q, recipientID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var list []domain.Notification
	for rows.Next() {
		var n domain.Notification
		if err := rows.Scan(
			&n.ID, &n.RecipientID, &n.ActorID, &n.NotificationType, &n.EntityType, &n.EntityID,
			&n.Title, &n.Message, &n.IsRead, &n.ReadAt, &n.CreatedAt,
		); err != nil {
			return nil, 0, err
		}
		list = append(list, n)
	}
	return list, total, nil
}

func (r *pgxNotificationRepository) MarkAsRead(ctx context.Context, id uuid.UUID, recipientID uuid.UUID) error {
	q := `UPDATE notifications SET is_read = TRUE, read_at = NOW() WHERE id = $1 AND recipient_id = $2`
	_, err := r.db.Exec(ctx, q, id, recipientID)
	return err
}

func (r *pgxNotificationRepository) MarkAllAsRead(ctx context.Context, recipientID uuid.UUID) error {
	q := `UPDATE notifications SET is_read = TRUE, read_at = NOW() WHERE recipient_id = $1 AND is_read = FALSE`
	_, err := r.db.Exec(ctx, q, recipientID)
	return err
}

func (r *pgxNotificationRepository) GetUnreadCount(ctx context.Context, recipientID uuid.UUID) (int64, error) {
	q := `SELECT COUNT(*) FROM notifications WHERE recipient_id = $1 AND is_read = FALSE AND deleted_at IS NULL`
	var count int64
	err := r.db.QueryRow(ctx, q, recipientID).Scan(&count)
	return count, err
}
