package repository

import (
	"context"
	"errors"

	"github.com/Masterminds/squirrel"
	"github.com/gofrs/uuid/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Rizwank123/chatApp/internal/domain"
)

type pgxChatRepository struct {
	db *pgxpool.Pool
	sq squirrel.StatementBuilderType
}

func NewChatRepository(db *pgxpool.Pool) domain.ChatRepository {
	return &pgxChatRepository{
		db: db,
		sq: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}
}

// Friendship methods
func (r *pgxChatRepository) CreateFriendship(ctx context.Context, friendship *domain.Friendship) error {
	q := r.sq.Insert("friendships").
		Columns("personnel_id", "friend_id", "status").
		Values(friendship.PersonnelID, friendship.FriendID, friendship.Status).
		Suffix("RETURNING id, created_at, updated_at")
	sql, args, err := q.ToSql()
	if err != nil {
		return err
	}
	return r.db.QueryRow(ctx, sql, args...).Scan(&friendship.ID, &friendship.CreatedAt, &friendship.UpdatedAt)
}

func (r *pgxChatRepository) UpdateFriendshipStatus(ctx context.Context, id uuid.UUID, status domain.FriendshipStatus) error {
	q := r.sq.Update("friendships").
		Set("status", status).
		Set("updated_at", squirrel.Expr("now()")).
		Where(squirrel.Eq{"id": id})
	sql, args, err := q.ToSql()
	if err != nil {
		return err
	}
	_, err = r.db.Exec(ctx, sql, args...)
	return err
}

func (r *pgxChatRepository) FindFriendshipByID(ctx context.Context, id uuid.UUID) (result domain.Friendship, err error) {
	q := r.sq.Select("*").From("friendships").Where(squirrel.Eq{"id": id})
	sql, args, err := q.ToSql()
	if err != nil {
		return result, err
	}
	var row pgx.Rows
	row, err = r.db.Query(ctx, sql, args...)
	if err != nil {
		return result, err
	}

	result, err = pgx.CollectOneRow(row, pgx.RowToStructByNameLax[domain.Friendship])
	return result, err
}

func (r *pgxChatRepository) FindFriendships(ctx context.Context, userID uuid.UUID) ([]domain.Friendship, error) {
	q := r.sq.Select("*").From("friendships").
		Where(squirrel.Or{
			squirrel.Eq{"personnel_id": userID},
			squirrel.Eq{"friend_id": userID},
		})
	sql, args, err := q.ToSql()
	if err != nil {
		return nil, err
	}
	rows, err := r.db.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return pgx.CollectRows(rows, pgx.RowToStructByNameLax[domain.Friendship])
}

// Conversation methods
func (r *pgxChatRepository) CreateConversation(ctx context.Context, conversation *domain.Conversation) error {
	q := r.sq.Insert("conversations").
		Columns("name", "conversation_type").
		Values(conversation.Name, conversation.ConversationType).
		Suffix("RETURNING id, created_at, updated_at")
	sql, args, err := q.ToSql()
	if err != nil {
		return err
	}
	return r.db.QueryRow(ctx, sql, args...).Scan(&conversation.ID, &conversation.CreatedAt, &conversation.UpdatedAt)
}

func (r *pgxChatRepository) FindConversationByID(ctx context.Context, id uuid.UUID) (result domain.Conversation, err error) {
	q := r.sq.Select("*").From("conversations").Where(squirrel.Eq{"id": id})
	sql, args, err := q.ToSql()
	if err != nil {
		return domain.Conversation{}, err
	}
	row, err := r.db.Query(ctx, sql, args...)
	if err != nil {
		return domain.Conversation{}, err
	}
	result, err = pgx.CollectOneRow(row, pgx.RowToStructByNameLax[domain.Conversation])
	return result, err
}

func (r *pgxChatRepository) FindDirectConversation(ctx context.Context, userA, userB uuid.UUID) (*domain.Conversation, error) {
	q := `SELECT c.id, c.name, c.conversation_type, c.created_at, c.updated_at
	      FROM conversations c
	      JOIN conversation_participants cp1 ON c.id = cp1.conversation_id AND cp1.personnel_id = $1 AND cp1.left_at IS NULL
	      JOIN conversation_participants cp2 ON c.id = cp2.conversation_id AND cp2.personnel_id = $2 AND cp2.left_at IS NULL
	      WHERE c.conversation_type = 'DIRECT' AND c.deleted_at IS NULL
	      LIMIT 1`
	var conv domain.Conversation
	err := r.db.QueryRow(ctx, q, userA, userB).Scan(&conv.ID, &conv.Name, &conv.ConversationType, &conv.CreatedAt, &conv.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &conv, nil
}

func (r *pgxChatRepository) FindUserConversations(ctx context.Context, userID uuid.UUID) ([]domain.Conversation, error) {
	q := r.sq.Select("c.*").
		From("conversations c").
		Join("conversation_participants cp ON c.id = cp.conversation_id").
		Where(squirrel.Eq{"cp.personnel_id": userID, "cp.left_at": nil})
	sql, args, err := q.ToSql()
	if err != nil {
		return nil, err
	}
	rows, err := r.db.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return pgx.CollectRows(rows, pgx.RowToStructByNameLax[domain.Conversation])
}

func (r *pgxChatRepository) FindConversationParticipantIDs(ctx context.Context, conversationID uuid.UUID) ([]uuid.UUID, error) {
	q := r.sq.Select("personnel_id").
		From("conversation_participants").
		Where(squirrel.Eq{"conversation_id": conversationID, "left_at": nil, "deleted_at": nil})
	sql, args, err := q.ToSql()
	if err != nil {
		return nil, err
	}
	rows, err := r.db.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var participantIDs []uuid.UUID
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		participantIDs = append(participantIDs, id)
	}
	return participantIDs, rows.Err()
}

func (r *pgxChatRepository) AddParticipant(ctx context.Context, participant *domain.ConversationParticipant) error {
	q := r.sq.Insert("conversation_participants").
		Columns("conversation_id", "personnel_id", "joined_at").
		Values(participant.ConversationID, participant.PersonnelID, participant.JoinedAt).
		Suffix("RETURNING id, created_at, updated_at")
	sql, args, err := q.ToSql()
	if err != nil {
		return err
	}
	return r.db.QueryRow(ctx, sql, args...).Scan(&participant.ID, &participant.CreatedAt, &participant.UpdatedAt)
}

func (r *pgxChatRepository) RemoveParticipant(ctx context.Context, conversationID, userID uuid.UUID) error {
	q := r.sq.Update("conversation_participants").
		Set("left_at", squirrel.Expr("now()")).
		Where(squirrel.Eq{"conversation_id": conversationID, "personnel_id": userID, "left_at": nil})
	sql, args, err := q.ToSql()
	if err != nil {
		return err
	}
	_, err = r.db.Exec(ctx, sql, args...)
	return err
}

// Message methods
func (r *pgxChatRepository) CreateMessage(ctx context.Context, message *domain.Message) error {
	q := r.sq.Insert("messages").
		Columns("id", "conversation_id", "sender_id", "message_type", "content").
		Values(message.ID, message.ConversationID, message.SenderID, message.MessageType, message.Content).
		Suffix("RETURNING created_at, updated_at")
	sql, args, err := q.ToSql()
	if err != nil {
		return err
	}
	return r.db.QueryRow(ctx, sql, args...).Scan(&message.CreatedAt, &message.UpdatedAt)
}

func (r *pgxChatRepository) FindMessagesByConversation(ctx context.Context, conversationID uuid.UUID, options domain.QueryOptions) ([]domain.Message, int64, error) {
	countQuery := r.sq.Select("COUNT(*)").From("messages").Where(squirrel.Eq{"conversation_id": conversationID})
	sqlC, argsC, err := countQuery.ToSql()
	if err != nil {
		return nil, 0, err
	}
	var total int64
	err = r.db.QueryRow(ctx, sqlC, argsC...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	q := r.sq.Select("*").From("messages").
		Where(squirrel.Eq{"conversation_id": conversationID}).
		OrderBy("created_at DESC").
		Limit(uint64(options.Limit)).Offset(uint64(options.Offset))
	sql, args, err := q.ToSql()
	if err != nil {
		return nil, total, err
	}
	rows, err := r.db.Query(ctx, sql, args...)
	if err != nil {
		return nil, total, err
	}
	defer rows.Close()
	messages, err := pgx.CollectRows(rows, pgx.RowToStructByNameLax[domain.Message])
	return messages, total, err
}

func (r *pgxChatRepository) UpdateMessageStatus(ctx context.Context, status *domain.MessageStatusEntity) error {
	q := r.sq.Insert("message_stat").
		Columns("id", "message_id", "personnel_id", "status").
		Values(status.ID, status.MessageID, status.PersonnelID, status.Status).
		Suffix(`ON CONFLICT (message_id, personnel_id) DO UPDATE SET status = EXCLUDED.status, updated_at = now() RETURNING created_at, updated_at`)
	sql, args, err := q.ToSql()
	if err != nil {
		return err
	}
	return r.db.QueryRow(ctx, sql, args...).Scan(&status.CreatedAt, &status.UpdatedAt)
}

func (r *pgxChatRepository) FindMessageStatuses(ctx context.Context, messageID uuid.UUID) ([]domain.MessageStatusEntity, error) {
	q := r.sq.Select("*").From("message_stat").Where(squirrel.Eq{"message_id": messageID})
	sql, args, err := q.ToSql()
	if err != nil {
		return nil, err
	}
	rows, err := r.db.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return pgx.CollectRows(rows, pgx.RowToStructByNameLax[domain.MessageStatusEntity])
}
