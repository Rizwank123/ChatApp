package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Rizwank123/chatApp/internal/domain"
)

type pgxPostRepository struct {
	db *pgxpool.Pool
}

func NewPostRepository(db *pgxpool.Pool) domain.PostRepository {
	return &pgxPostRepository{db: db}
}

// Add this constant at the top of the file after the imports

func (r *pgxPostRepository) CreatePost(ctx context.Context, entity *domain.Post) error {
	if ctx == nil {
		ctx = context.Background()
	}
	txVal := ctx.Value(TxKey)

	status := entity.Status
	if status == "" {
		status = "PUBLISHED"
	}
	visibility := entity.Visibility
	if visibility == "" {
		visibility = "PUBLIC"
	}

	query := `
		INSERT INTO posts (personnel_id, caption, media_url, location, visibility, status, is_archived)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, created_at, updated_at
	`
	args := []interface{}{entity.PersonnelID, entity.Caption, entity.MediaURL, entity.Location, visibility, status, entity.IsArchived}

	var err error
	if txVal != nil {
		tx := txVal.(pgx.Tx)
		err = tx.QueryRow(ctx, query, args...).Scan(&entity.ID, &entity.CreatedAt, &entity.UpdatedAt)
	} else {
		err = r.db.QueryRow(ctx, query, args...).Scan(&entity.ID, &entity.CreatedAt, &entity.UpdatedAt)
	}
	entity.Status = status
	entity.Visibility = visibility
	return err
}

func (r *pgxPostRepository) FindPostByID(ctx context.Context, id uuid.UUID, viewerID ...uuid.UUID) (result domain.Post, err error) {
	if ctx == nil {
		ctx = context.Background()
	}

	var vID uuid.UUID
	if len(viewerID) > 0 {
		vID = viewerID[0]
	}

	query := `
		SELECT p.id, p.personnel_id, p.caption, p.media_url, p.location, p.visibility,
		       COALESCE(p.status, 'PUBLISHED') as status,
		       COALESCE(p.is_archived, false) as is_archived,
		       COALESCE(p.view_count, 0) as view_count,
		       COALESCE(p.share_count, 0) as share_count,
		       (SELECT COUNT(*) FROM likes l WHERE l.post_id = p.id AND l.deleted_at IS NULL) as likes,
		       (SELECT COUNT(*) FROM comments c WHERE c.post_id = p.id AND c.deleted_at IS NULL) as comments,
		       u.user_name as creator_username,
		       TRIM(per.first_name || ' ' || per.last_name) as creator_name,
		       COALESCE(per.avatar, '') as creator_avatar,
		       EXISTS(SELECT 1 FROM likes l2 WHERE l2.post_id = p.id AND l2.personnel_id = $2 AND l2.deleted_at IS NULL) as is_liked,
		       EXISTS(SELECT 1 FROM saved_posts sp WHERE sp.post_id = p.id AND sp.personnel_id = $2) as is_saved,
		       p.created_at, p.updated_at
		FROM posts p
		JOIN personnel per ON p.personnel_id = per.id
		JOIN users u ON per.user_id = u.id
		WHERE p.id = $1 AND p.deleted_at IS NULL
	`

	err = r.db.QueryRow(ctx, query, id, vID).Scan(
		&result.ID, &result.PersonnelID, &result.Caption, &result.MediaURL, &result.Location, &result.Visibility,
		&result.Status, &result.IsArchived, &result.ViewCount, &result.ShareCount,
		&result.Likes, &result.Comments,
		&result.CreatorUsername, &result.CreatorName, &result.CreatorAvatar,
		&result.IsLiked, &result.IsSaved,
		&result.CreatedAt, &result.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return domain.Post{}, fmt.Errorf("post not found")
		}
		return domain.Post{}, err
	}

	_ = r.enrichPostData(ctx, &result)
	return result, nil
}

// Example of modifying a method that uses Exec
func (r *pgxPostRepository) DeletePost(ctx context.Context, id uuid.UUID) error {
	if ctx == nil {
		ctx = context.Background()
	}
	txVal := ctx.Value(TxKey)

	query := `UPDATE posts SET deleted_at = CURRENT_TIMESTAMP WHERE id = $1 AND deleted_at IS NULL`

	var err error
	if txVal != nil {
		tx := txVal.(pgx.Tx)
		_, err = tx.Exec(ctx, query, id)
	} else {
		_, err = r.db.Exec(ctx, query, id)
	}
	return err
}

func (r *pgxPostRepository) UpdatePost(ctx context.Context, post *domain.Post) error {
	if ctx == nil {
		ctx = context.Background()
	}
	txVal := ctx.Value(TxKey)

	status := post.Status
	if status == "" {
		status = "PUBLISHED"
	}
	visibility := post.Visibility
	if visibility == "" {
		visibility = "PUBLIC"
	}

	query := `
		UPDATE posts
		SET caption = $1, media_url = $2, location = $3, visibility = $4, status = $5, is_archived = $6, updated_at = CURRENT_TIMESTAMP
		WHERE id = $7 AND deleted_at IS NULL RETURNING updated_at
	`
	args := []interface{}{post.Caption, post.MediaURL, post.Location, visibility, status, post.IsArchived, post.ID}

	var err error
	if txVal != nil {
		tx := txVal.(pgx.Tx)
		err = tx.QueryRow(ctx, query, args...).Scan(&post.UpdatedAt)
	} else {
		err = r.db.QueryRow(ctx, query, args...).Scan(&post.UpdatedAt)
	}
	return err
}

func (r *pgxPostRepository) FindUserPosts(ctx context.Context, personnelID uuid.UUID, options domain.QueryOptions, viewerID ...uuid.UUID) (result []domain.Post, total int64, err error) {
	if ctx == nil {
		ctx = context.Background()
	}

	var vID uuid.UUID
	if len(viewerID) > 0 {
		vID = viewerID[0]
	}

	query := `
		SELECT p.id, p.personnel_id, p.caption, p.media_url, p.location, p.visibility,
		       COALESCE(p.status, 'PUBLISHED') as status,
		       COALESCE(p.is_archived, false) as is_archived,
		       COALESCE(p.view_count, 0) as view_count,
		       COALESCE(p.share_count, 0) as share_count,
		       (SELECT COUNT(*) FROM likes l WHERE l.post_id = p.id AND l.deleted_at IS NULL) as likes,
		       (SELECT COUNT(*) FROM comments c WHERE c.post_id = p.id AND c.deleted_at IS NULL) as comments,
		       u.user_name as creator_username,
		       TRIM(per.first_name || ' ' || per.last_name) as creator_name,
		       COALESCE(per.avatar, '') as creator_avatar,
		       EXISTS(SELECT 1 FROM likes l2 WHERE l2.post_id = p.id AND l2.personnel_id = $4 AND l2.deleted_at IS NULL) as is_liked,
		       EXISTS(SELECT 1 FROM saved_posts sp WHERE sp.post_id = p.id AND sp.personnel_id = $4) as is_saved,
		       p.created_at, p.updated_at
		FROM posts p
		JOIN personnel per ON p.personnel_id = per.id
		JOIN users u ON per.user_id = u.id
		WHERE p.personnel_id = $1 AND p.deleted_at IS NULL
		ORDER BY p.created_at DESC
		LIMIT $2 OFFSET $3
	`

	limit := options.Limit
	if limit <= 0 {
		limit = 20
	}
	rows, err := r.db.Query(ctx, query, personnelID, limit, options.Offset, vID)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	for rows.Next() {
		var p domain.Post
		if err := rows.Scan(
			&p.ID, &p.PersonnelID, &p.Caption, &p.MediaURL, &p.Location, &p.Visibility,
			&p.Status, &p.IsArchived, &p.ViewCount, &p.ShareCount,
			&p.Likes, &p.Comments,
			&p.CreatorUsername, &p.CreatorName, &p.CreatorAvatar,
			&p.IsLiked, &p.IsSaved,
			&p.CreatedAt, &p.UpdatedAt,
		); err != nil {
			return nil, 0, err
		}
		_ = r.enrichPostData(ctx, &p)
		result = append(result, p)
	}

	countQuery := `SELECT COUNT(*) FROM posts WHERE personnel_id = $1 AND deleted_at IS NULL`
	err = r.db.QueryRow(ctx, countQuery, personnelID).Scan(&total)
	return result, total, err
}

func (r *pgxPostRepository) FindFeedPosts(ctx context.Context, personnelID uuid.UUID, options domain.QueryOptions) (result []domain.Post, total int64, err error) {
	if ctx == nil {
		ctx = context.Background()
	}

	limit := options.Limit
	if limit <= 0 {
		limit = 20
	}

	// 1. Fetch posts from followed accounts + author's own posts
	query := `
		SELECT DISTINCT p.id, p.personnel_id, p.caption, p.media_url, p.location, p.visibility,
		       COALESCE(p.status, 'PUBLISHED') as status,
		       COALESCE(p.is_archived, false) as is_archived,
		       COALESCE(p.view_count, 0) as view_count,
		       COALESCE(p.share_count, 0) as share_count,
		       (SELECT COUNT(*) FROM likes l WHERE l.post_id = p.id AND l.deleted_at IS NULL) as likes,
		       (SELECT COUNT(*) FROM comments c WHERE c.post_id = p.id AND c.deleted_at IS NULL) as comments,
		       u.user_name as creator_username,
		       TRIM(per.first_name || ' ' || per.last_name) as creator_name,
		       COALESCE(per.avatar, '') as creator_avatar,
		       EXISTS(SELECT 1 FROM likes l2 WHERE l2.post_id = p.id AND l2.personnel_id = $1 AND l2.deleted_at IS NULL) as is_liked,
		       EXISTS(SELECT 1 FROM saved_posts sp WHERE sp.post_id = p.id AND sp.personnel_id = $1) as is_saved,
		       p.created_at, p.updated_at
		FROM posts p
		JOIN personnel per ON p.personnel_id = per.id
		JOIN users u ON per.user_id = u.id
		LEFT JOIN follows f ON p.personnel_id = f.following_id AND f.follower_id = $1 AND f.deleted_at IS NULL
		WHERE (f.follower_id IS NOT NULL OR p.personnel_id = $1)
		  AND p.deleted_at IS NULL
		  AND (p.status IS NULL OR p.status = 'PUBLISHED')
		  AND (p.is_archived IS NULL OR p.is_archived = false)
		ORDER BY p.created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.Query(ctx, query, personnelID, limit, options.Offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	for rows.Next() {
		var p domain.Post
		if err := rows.Scan(
			&p.ID, &p.PersonnelID, &p.Caption, &p.MediaURL, &p.Location, &p.Visibility,
			&p.Status, &p.IsArchived, &p.ViewCount, &p.ShareCount,
			&p.Likes, &p.Comments,
			&p.CreatorUsername, &p.CreatorName, &p.CreatorAvatar,
			&p.IsLiked, &p.IsSaved,
			&p.CreatedAt, &p.UpdatedAt,
		); err != nil {
			return nil, 0, err
		}
		_ = r.enrichPostData(ctx, &p)
		result = append(result, p)
	}

	countQuery := `
		SELECT COUNT(DISTINCT p.id)
		FROM posts p
		LEFT JOIN follows f ON p.personnel_id = f.following_id AND f.follower_id = $1 AND f.deleted_at IS NULL
		WHERE (f.follower_id IS NOT NULL OR p.personnel_id = $1)
		  AND p.deleted_at IS NULL
		  AND (p.status IS NULL OR p.status = 'PUBLISHED')
		  AND (p.is_archived IS NULL OR p.is_archived = false)
	`
	_ = r.db.QueryRow(ctx, countQuery, personnelID).Scan(&total)

	// If user feed is empty (e.g. brand new user following 0 people with 0 posts), gracefully fall back to latest public posts
	if len(result) == 0 && options.Offset == 0 {
		return r.FindPublicPosts(ctx, options, personnelID)
	}

	return result, total, nil
}

func (r *pgxPostRepository) FindPublicPosts(ctx context.Context, options domain.QueryOptions, viewerID ...uuid.UUID) (result []domain.Post, total int64, err error) {
	if ctx == nil {
		ctx = context.Background()
	}

	var vID uuid.UUID
	if len(viewerID) > 0 {
		vID = viewerID[0]
	}

	limit := options.Limit
	if limit <= 0 {
		limit = 20
	}

	query := `
		SELECT p.id, p.personnel_id, p.caption, p.media_url, p.location, p.visibility,
		       COALESCE(p.status, 'PUBLISHED') as status,
		       COALESCE(p.is_archived, false) as is_archived,
		       COALESCE(p.view_count, 0) as view_count,
		       COALESCE(p.share_count, 0) as share_count,
		       (SELECT COUNT(*) FROM likes l WHERE l.post_id = p.id AND l.deleted_at IS NULL) as likes,
		       (SELECT COUNT(*) FROM comments c WHERE c.post_id = p.id AND c.deleted_at IS NULL) as comments,
		       u.user_name as creator_username,
		       TRIM(per.first_name || ' ' || per.last_name) as creator_name,
		       COALESCE(per.avatar, '') as creator_avatar,
		       EXISTS(SELECT 1 FROM likes l2 WHERE l2.post_id = p.id AND l2.personnel_id = $3 AND l2.deleted_at IS NULL) as is_liked,
		       EXISTS(SELECT 1 FROM saved_posts sp WHERE sp.post_id = p.id AND sp.personnel_id = $3) as is_saved,
		       p.created_at, p.updated_at
		FROM posts p
		JOIN personnel per ON p.personnel_id = per.id
		JOIN users u ON per.user_id = u.id
		WHERE p.deleted_at IS NULL
		  AND (p.visibility = 'PUBLIC' OR p.visibility IS NULL)
		  AND (p.status = 'PUBLISHED' OR p.status IS NULL)
		  AND (p.is_archived = false OR p.is_archived IS NULL)
		ORDER BY p.created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.db.Query(ctx, query, limit, options.Offset, vID)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	for rows.Next() {
		var p domain.Post
		if err := rows.Scan(
			&p.ID, &p.PersonnelID, &p.Caption, &p.MediaURL, &p.Location, &p.Visibility,
			&p.Status, &p.IsArchived, &p.ViewCount, &p.ShareCount,
			&p.Likes, &p.Comments,
			&p.CreatorUsername, &p.CreatorName, &p.CreatorAvatar,
			&p.IsLiked, &p.IsSaved,
			&p.CreatedAt, &p.UpdatedAt,
		); err != nil {
			return nil, 0, err
		}
		_ = r.enrichPostData(ctx, &p)
		result = append(result, p)
	}

	countQuery := `
		SELECT COUNT(*)
		FROM posts p
		WHERE p.deleted_at IS NULL
		  AND (p.visibility = 'PUBLIC' OR p.visibility IS NULL)
		  AND (p.status = 'PUBLISHED' OR p.status IS NULL)
		  AND (p.is_archived = false OR p.is_archived IS NULL)
	`
	_ = r.db.QueryRow(ctx, countQuery).Scan(&total)
	return result, total, nil
}

func (r *pgxPostRepository) IncrementPostView(ctx context.Context, postID uuid.UUID) error {
	if ctx == nil {
		ctx = context.Background()
	}
	query := `UPDATE posts SET view_count = COALESCE(view_count, 0) + 1 WHERE id = $1 AND deleted_at IS NULL`
	_, err := r.db.Exec(ctx, query, postID)
	return err
}

func (r *pgxPostRepository) IncrementPostShare(ctx context.Context, postID uuid.UUID) error {
	if ctx == nil {
		ctx = context.Background()
	}
	query := `UPDATE posts SET share_count = COALESCE(share_count, 0) + 1 WHERE id = $1 AND deleted_at IS NULL`
	_, err := r.db.Exec(ctx, query, postID)
	return err
}

func (r *pgxPostRepository) CreateUserMention(ctx context.Context, postID, mentionedPersonnelID uuid.UUID) error {
	if ctx == nil {
		ctx = context.Background()
	}
	query := `
		INSERT INTO user_mentions (post_id, mentioned_personnel_id)
		VALUES ($1, $2)
		ON CONFLICT (post_id, mentioned_personnel_id) DO NOTHING
	`
	_, err := r.db.Exec(ctx, query, postID, mentionedPersonnelID)
	return err
}

func (r *pgxPostRepository) GetUserMentions(ctx context.Context, postID uuid.UUID) ([]uuid.UUID, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	query := `SELECT mentioned_personnel_id FROM user_mentions WHERE post_id = $1`
	rows, err := r.db.Query(ctx, query, postID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ids []uuid.UUID
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err == nil {
			ids = append(ids, id)
		}
	}
	return ids, nil
}

// Helper function to enrich post data with post_media attachments
func (r *pgxPostRepository) enrichPostData(ctx context.Context, post *domain.Post) error {
	if post == nil {
		return nil
	}

	query := `
		SELECT id, post_id, owner_id, bucket, object_key, original_filename, mime_type,
		       file_size, media_type, width, height, duration, thumbnail_key, status, display_order,
		       created_at, updated_at
		FROM post_media
		WHERE post_id = $1
		ORDER BY display_order ASC, created_at ASC
	`
	rows, err := r.db.Query(ctx, query, post.ID)
	if err != nil {
		return nil // Non-fatal, return post without media
	}
	defer rows.Close()

	var mediaList []domain.PostMedia
	for rows.Next() {
		var m domain.PostMedia
		var thumbKey *string
		var origName *string
		if err := rows.Scan(
			&m.ID, &m.PostID, &m.OwnerID, &m.Bucket, &m.ObjectKey, &origName,
			&m.MIMEType, &m.FileSize, &m.MediaType, &m.Width, &m.Height, &m.Duration,
			&thumbKey, &m.Status, &m.DisplayOrder, &m.CreatedAt, &m.UpdatedAt,
		); err == nil {
			if thumbKey != nil {
				m.ThumbnailKey = *thumbKey
			}
			if origName != nil {
				m.OriginalFilename = *origName
			}
			mediaList = append(mediaList, m)
		}
	}

	if len(mediaList) > 0 {
		post.Media = mediaList
		// Populate media_url if empty for backward compatibility
		if len(post.MediaURL) == 0 {
			for _, m := range mediaList {
				post.MediaURL = append(post.MediaURL, m.ObjectKey)
			}
		}
	}

	return nil
}

// FindOrCreateDirectConversation finds or creates a direct conversation between two users
func (r *pgxPostRepository) FindOrCreateDirectConversation(ctx context.Context, personnelID1, personnelID2 uuid.UUID) (domain.Conversation, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	txVal := ctx.Value(TxKey)

	// First check if a conversation already exists between these users
	query := `
		SELECT c.id, c.name, c.conversation_type, c.created_at, c.updated_at, c.deleted_at
		FROM conversations c
		JOIN conversation_participants p1 ON c.id = p1.conversation_id
		JOIN conversation_participants p2 ON c.id = p2.conversation_id
		WHERE c.conversation_type = $1
		AND p1.personnel_id = $2 AND p2.personnel_id = $3
		AND c.deleted_at IS NULL
		AND p1.deleted_at IS NULL AND p2.deleted_at IS NULL
		LIMIT 1
	`

	var conversation domain.Conversation
	var rows pgx.Rows
	var err error

	if txVal != nil {
		tx := txVal.(pgx.Tx)
		rows, err = tx.Query(ctx, query, domain.ConversationTypeDirect, personnelID1, personnelID2)
	} else {
		rows, err = r.db.Query(ctx, query, domain.ConversationTypeDirect, personnelID1, personnelID2)
	}

	if err != nil {
		return conversation, err
	}
	defer rows.Close()

	// If conversation exists, return it
	conversations, err := pgx.CollectRows(rows, pgx.RowToStructByNameLax[domain.Conversation])
	if err != nil {
		return conversation, err
	}

	if len(conversations) > 0 {
		return conversations[0], nil
	}

	// Create a new conversation if none exists
	conversation = domain.Conversation{
		Base: domain.Base{
			ID: uuid.Must(uuid.NewV4()),
		},
		ConversationType: domain.ConversationTypeDirect,
	}

	// Start a transaction if not already in one
	var tx pgx.Tx
	var needsCommit bool

	if txVal != nil {
		tx = txVal.(pgx.Tx)
	} else {
		tx, err = r.db.Begin(ctx)
		if err != nil {
			return conversation, err
		}
		needsCommit = true
		defer func() {
			if err != nil {
				tx.Rollback(ctx)
			} else if needsCommit {
				tx.Commit(ctx)
			}
		}()
	}

	// Insert the conversation
	insertQuery := `INSERT INTO conversations (conversation_type) VALUES ($1) RETURNING id, created_at, updated_at`
	err = tx.QueryRow(ctx, insertQuery, conversation.ConversationType).Scan(&conversation.ID, &conversation.CreatedAt, &conversation.UpdatedAt)
	if err != nil {
		return conversation, err
	}

	// Add both users as participants
	participant1 := domain.ConversationParticipant{
		ConversationID: conversation.ID,
		PersonnelID:    personnelID1,
		JoinedAt:       time.Now(),
	}

	participant2 := domain.ConversationParticipant{
		ConversationID: conversation.ID,
		PersonnelID:    personnelID2,
		JoinedAt:       time.Now(),
	}

	// Insert participants
	participantQuery := `INSERT INTO conversation_participants (conversation_id, personnel_id, joined_at) VALUES ($1, $2, $3)`
	_, err = tx.Exec(ctx, participantQuery, participant1.ConversationID, participant1.PersonnelID, participant1.JoinedAt)
	if err != nil {
		return conversation, err
	}

	_, err = tx.Exec(ctx, participantQuery, participant2.ConversationID, participant2.PersonnelID, participant2.JoinedAt)
	if err != nil {
		return conversation, err
	}

	return conversation, nil
}

// SaveDirectMessage saves a direct message to the database
func (r *pgxPostRepository) SaveDirectMessage(ctx context.Context, message *domain.Message) error {
	if ctx == nil {
		ctx = context.Background()
	}
	txVal := ctx.Value(TxKey)

	query := `
		INSERT INTO messages (id, conversation_id, sender_id, message_type, content)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING created_at, updated_at
	`
	args := []interface{}{
		message.ID, message.ConversationID, message.SenderID, message.MessageType, message.Content,
	}

	if txVal != nil {
		tx := txVal.(pgx.Tx)
		err := tx.QueryRow(ctx, query, args...).Scan(&message.CreatedAt, &message.UpdatedAt)
		return err
	}

	return r.db.QueryRow(ctx, query, args...).Scan(&message.CreatedAt, &message.UpdatedAt)
}

// FindDirectMessages finds direct messages between two users
func (r *pgxPostRepository) FindDirectMessages(ctx context.Context, personnelID1, personnelID2 uuid.UUID, options domain.QueryOptions) ([]domain.Message, int64, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	txVal := ctx.Value(TxKey)

	// First find the conversation between these users
	conversation, err := r.FindOrCreateDirectConversation(ctx, personnelID1, personnelID2)
	if err != nil {
		return nil, 0, err
	}

	// Query for messages in this conversation
	query := `
		SELECT id, conversation_id, sender_id, message_type, content, created_at, updated_at
		FROM messages
		WHERE conversation_id = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	var rows pgx.Rows
	if txVal != nil {
		tx := txVal.(pgx.Tx)
		rows, err = tx.Query(ctx, query, conversation.ID, options.Limit, options.Offset)
	} else {
		rows, err = r.db.Query(ctx, query, conversation.ID, options.Limit, options.Offset)
	}

	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	messages, err := pgx.CollectRows(rows, pgx.RowToStructByNameLax[domain.Message])
	if err != nil {
		return nil, 0, err
	}

	// Count total messages
	countQuery := `SELECT COUNT(*) FROM messages WHERE conversation_id = $1 AND deleted_at IS NULL`
	var total int64

	if txVal != nil {
		tx := txVal.(pgx.Tx)
		err = tx.QueryRow(ctx, countQuery, conversation.ID).Scan(&total)
	} else {
		err = r.db.QueryRow(ctx, countQuery, conversation.ID).Scan(&total)
	}

	return messages, total, err
}

func (r *pgxPostRepository) CreateFollow(ctx context.Context, follow *domain.Follow) error {
	if ctx == nil {
		ctx = context.Background()
	}
	txVal := ctx.Value(TxKey)

	query := `
		INSERT INTO follows (follower_id, following_id)
		VALUES ($1, $2)
		ON CONFLICT (follower_id, following_id)
		DO UPDATE SET deleted_at = NULL, updated_at = CURRENT_TIMESTAMP
		RETURNING id, created_at, updated_at
	`
	args := []interface{}{follow.FollowerID, follow.FollowingID}

	if txVal != nil {
		tx := txVal.(pgx.Tx)
		err := tx.QueryRow(ctx, query, args...).Scan(&follow.ID, &follow.CreatedAt, &follow.UpdatedAt)
		return err
	}
	return r.db.QueryRow(ctx, query, args...).Scan(&follow.ID, &follow.CreatedAt, &follow.UpdatedAt)
}

func (r *pgxPostRepository) DeleteFollow(ctx context.Context, followerID, followingID uuid.UUID) error {
	if ctx == nil {
		ctx = context.Background()
	}
	txVal := ctx.Value(TxKey)

	query := `UPDATE follows SET deleted_at = CURRENT_TIMESTAMP WHERE follower_id = $1 AND following_id = $2 AND deleted_at IS NULL`

	if txVal != nil {
		tx := txVal.(pgx.Tx)
		_, err := tx.Exec(ctx, query, followerID, followingID)
		return err
	}
	_, err := r.db.Exec(ctx, query, followerID, followingID)
	return err
}

func (r *pgxPostRepository) IsFollowing(ctx context.Context, followerID, followingID uuid.UUID) (bool, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	txVal := ctx.Value(TxKey)

	query := `SELECT EXISTS(SELECT 1 FROM follows WHERE follower_id = $1 AND following_id = $2 AND deleted_at IS NULL)`
	args := []interface{}{followerID, followingID}

	var exist bool
	if txVal != nil {
		tx := txVal.(pgx.Tx)
		err := tx.QueryRow(ctx, query, args...).Scan(&exist)
		return exist, err
	}
	err := r.db.QueryRow(ctx, query, args...).Scan(&exist)
	return exist, err
}

// CheckMutualFollow checks if two users are following each other
func (r *pgxPostRepository) CheckMutualFollow(ctx context.Context, personnelID1, personnelID2 uuid.UUID) (bool, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	txVal := ctx.Value(TxKey)

	// Check if user1 follows user2 AND user2 follows user1
	query := `
		SELECT EXISTS(
			SELECT 1 FROM follows f1 
			JOIN follows f2 ON f1.follower_id = f2.following_id AND f1.following_id = f2.follower_id
			WHERE f1.follower_id = $1 AND f1.following_id = $2 
			AND f1.deleted_at IS NULL AND f2.deleted_at IS NULL
		)
	`
	args := []interface{}{personnelID1, personnelID2}

	var mutualFollow bool
	if txVal != nil {
		tx := txVal.(pgx.Tx)
		err := tx.QueryRow(ctx, query, args...).Scan(&mutualFollow)
		return mutualFollow, err
	}
	err := r.db.QueryRow(ctx, query, args...).Scan(&mutualFollow)
	return mutualFollow, err
}

func (r *pgxPostRepository) CreateLike(ctx context.Context, like *domain.Like) error {
	if ctx == nil {
		ctx = context.Background()
	}
	txVal := ctx.Value(TxKey)

	query := `
		INSERT INTO likes (personnel_id, post_id)
		VALUES ($1, $2)
		ON CONFLICT (personnel_id, post_id)
		DO UPDATE SET deleted_at = NULL, updated_at = CURRENT_TIMESTAMP
		RETURNING id, created_at, updated_at
	`
	args := []interface{}{like.PersonnelID, like.PostID}

	if txVal != nil {
		tx := txVal.(pgx.Tx)
		return tx.QueryRow(ctx, query, args...).Scan(&like.ID, &like.CreatedAt, &like.UpdatedAt)
	}
	return r.db.QueryRow(ctx, query, args...).Scan(&like.ID, &like.CreatedAt, &like.UpdatedAt)
}

func (r *pgxPostRepository) DeleteLike(ctx context.Context, personnelID, postID uuid.UUID) error {
	if ctx == nil {
		ctx = context.Background()
	}
	txVal := ctx.Value(TxKey)

	query := `UPDATE likes SET deleted_at = CURRENT_TIMESTAMP WHERE personnel_id = $1 AND post_id = $2 AND deleted_at IS NULL`

	if txVal != nil {
		tx := txVal.(pgx.Tx)
		_, err := tx.Exec(ctx, query, personnelID, postID)
		return err
	}
	_, err := r.db.Exec(ctx, query, personnelID, postID)
	return err
}

func (r *pgxPostRepository) CreateComment(ctx context.Context, comment *domain.Comment) error {
	if ctx == nil {
		ctx = context.Background()
	}
	txVal := ctx.Value(TxKey)

	query := `INSERT INTO comments (personnel_id, post_id, parent_id, content) VALUES ($1, $2, $3, $4) RETURNING id, created_at, updated_at`
	args := []interface{}{comment.PersonnelID, comment.PostID, comment.ParentID, comment.Content}

	if txVal != nil {
		tx := txVal.(pgx.Tx)
		return tx.QueryRow(ctx, query, args...).Scan(&comment.ID, &comment.CreatedAt, &comment.UpdatedAt)
	}
	return r.db.QueryRow(ctx, query, args...).Scan(&comment.ID, &comment.CreatedAt, &comment.UpdatedAt)
}

func (r *pgxPostRepository) DeleteComment(ctx context.Context, id uuid.UUID, personnelID uuid.UUID) error {
	if ctx == nil {
		ctx = context.Background()
	}
	txVal := ctx.Value(TxKey)

	query := `UPDATE comments SET deleted_at = CURRENT_TIMESTAMP WHERE id = $1 AND personnel_id = $2 AND deleted_at IS NULL`

	if txVal != nil {
		tx := txVal.(pgx.Tx)
		_, err := tx.Exec(ctx, query, id, personnelID)
		return err
	}
	_, err := r.db.Exec(ctx, query, id, personnelID)
	return err
}

func (r *pgxPostRepository) FindPostComments(ctx context.Context, postID uuid.UUID, options domain.QueryOptions) (result []domain.Comment, total int64, err error) {
	if ctx == nil {
		ctx = context.Background()
	}
	txVal := ctx.Value(TxKey)

	query := `
		SELECT id, personnel_id, post_id, parent_id, content, created_at, updated_at, deleted_at
		FROM comments
		WHERE post_id = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	var rows pgx.Rows
	if txVal != nil {
		tx := txVal.(pgx.Tx)
		rows, err = tx.Query(ctx, query, postID, options.Limit, options.Offset)
	} else {
		rows, err = r.db.Query(ctx, query, postID, options.Limit, options.Offset)
	}
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	result, err = pgx.CollectRows(rows, pgx.RowToStructByNameLax[domain.Comment])
	if err != nil {
		return nil, 0, err
	}

	countQuery := `SELECT COUNT(*) FROM comments WHERE post_id = $1 AND deleted_at IS NULL`
	if txVal != nil {
		tx := txVal.(pgx.Tx)
		err = tx.QueryRow(ctx, countQuery, postID).Scan(&total)
	} else {
		err = r.db.QueryRow(ctx, countQuery, postID).Scan(&total)
	}
	return result, total, err
}

func (r *pgxPostRepository) CreateHashtag(ctx context.Context, hashtag *domain.Hashtag) error {
	if ctx == nil {
		ctx = context.Background()
	}
	txVal := ctx.Value(TxKey)

	query := `
		INSERT INTO hashtags (name)
		VALUES ($1)
		ON CONFLICT (name) DO UPDATE SET deleted_at = NULL, updated_at = CURRENT_TIMESTAMP
		RETURNING id, created_at, updated_at
	`
	args := []interface{}{hashtag.Name}

	if txVal != nil {
		tx := txVal.(pgx.Tx)
		return tx.QueryRow(ctx, query, args...).Scan(&hashtag.ID, &hashtag.CreatedAt, &hashtag.UpdatedAt)
	}
	return r.db.QueryRow(ctx, query, args...).Scan(&hashtag.ID, &hashtag.CreatedAt, &hashtag.UpdatedAt)
}

func (r *pgxPostRepository) FindHashtagByName(ctx context.Context, name string) (domain.Hashtag, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	txVal := ctx.Value(TxKey)

	query := `SELECT id, name FROM hashtags WHERE name = $1 AND deleted_at IS NULL`
	var hashtag domain.Hashtag

	if txVal != nil {
		tx := txVal.(pgx.Tx)
		err := tx.QueryRow(ctx, query, name).Scan(&hashtag.ID, &hashtag.Name)
		if err == pgx.ErrNoRows {
			return hashtag, fmt.Errorf("hashtag not found")
		}
		return hashtag, err
	}
	err := r.db.QueryRow(ctx, query, name).Scan(&hashtag.ID, &hashtag.Name)
	if err == pgx.ErrNoRows {
		return hashtag, fmt.Errorf("hashtag not found")
	}
	return hashtag, err
}

func (r *pgxPostRepository) CreatePostHashtag(ctx context.Context, postHashtag *domain.PostHashtag) error {
	if ctx == nil {
		ctx = context.Background()
	}
	txVal := ctx.Value(TxKey)

	query := `
		INSERT INTO post_hashtags (post_id, hashtag_id)
		VALUES ($1, $2)
		ON CONFLICT (post_id, hashtag_id) DO NOTHING
	`
	args := []interface{}{postHashtag.PostID, postHashtag.HashtagID}

	if txVal != nil {
		tx := txVal.(pgx.Tx)
		_, err := tx.Exec(ctx, query, args...)
		return err
	}
	_, err := r.db.Exec(ctx, query, args...)
	return err
}
