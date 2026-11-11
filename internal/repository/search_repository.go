package repository

import (
	"context"
	"fmt"
	"strings"

	"github.com/gofrs/uuid/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Rizwank123/chatApp/internal/domain"
)

type pgxSearchRepository struct {
	db       *pgxpool.Pool
	postRepo domain.PostRepository
}

// NewSearchRepository instantiates the PostgreSQL pg_trgm & tsvector search repository.
func NewSearchRepository(db *pgxpool.Pool, postRepo domain.PostRepository) domain.SearchRepository {
	return &pgxSearchRepository{
		db:       db,
		postRepo: postRepo,
	}
}

func (r *pgxSearchRepository) SearchUsers(ctx context.Context, viewerID uuid.UUID, query string, limit, offset int) ([]domain.UserSearchResult, int64, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	cleanQuery := strings.TrimSpace(query)
	if cleanQuery == "" {
		return []domain.UserSearchResult{}, 0, nil
	}
	if limit <= 0 {
		limit = 20
	}

	searchPattern := "%" + cleanQuery + "%"
	prefixPattern := cleanQuery + "%"

	sqlQuery := `
		SELECT per.id as personnel_id,
		       per.user_id,
		       u.user_name,
		       TRIM(per.first_name || ' ' || per.last_name) as display_name,
		       COALESCE(per.avatar, '') as avatar,
		       COALESCE(per.address->>'bio', '') as bio,
		       COALESCE(pr.is_private, false) as is_private,
		       EXISTS(SELECT 1 FROM follows f WHERE f.follower_id = $1 AND f.following_id = per.id AND f.deleted_at IS NULL) as is_following,
		       EXISTS(SELECT 1 FROM follows f2 WHERE f2.follower_id = per.id AND f2.following_id = $1 AND f2.deleted_at IS NULL) as follows_you,
		       (
		           CASE 
		               WHEN LOWER(u.user_name) = LOWER($2) THEN 1.0
		               WHEN LOWER(u.user_name) LIKE LOWER($3) THEN 0.85
		               WHEN LOWER(u.user_name) LIKE LOWER($4) THEN 0.7
		               ELSE similarity(u.user_name, $2)
		           END
		       ) as match_score
		FROM personnel per
		JOIN users u ON per.user_id = u.id
		LEFT JOIN privacy_settings pr ON per.id = pr.personnel_id
		WHERE per.deleted_at IS NULL
		  AND (
		      u.user_name ILIKE $4
		      OR per.first_name ILIKE $4
		      OR per.last_name ILIKE $4
		      OR similarity(u.user_name, $2) > 0.15
		  )
		ORDER BY match_score DESC, u.user_name ASC
		LIMIT $5 OFFSET $6
	`

	rows, err := r.db.Query(ctx, sqlQuery, viewerID, cleanQuery, prefixPattern, searchPattern, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var results []domain.UserSearchResult
	for rows.Next() {
		var u domain.UserSearchResult
		if err := rows.Scan(
			&u.PersonnelID,
			&u.UserID,
			&u.UserName,
			&u.DisplayName,
			&u.Avatar,
			&u.Bio,
			&u.IsPrivate,
			&u.IsFollowing,
			&u.FollowsYou,
			&u.MatchScore,
		); err != nil {
			return nil, 0, err
		}
		u.IsMutual = u.IsFollowing && u.FollowsYou
		results = append(results, u)
	}

	countQuery := `
		SELECT COUNT(*)
		FROM personnel per
		JOIN users u ON per.user_id = u.id
		WHERE per.deleted_at IS NULL
		  AND (
		      u.user_name ILIKE $1
		      OR per.first_name ILIKE $1
		      OR per.last_name ILIKE $1
		      OR similarity(u.user_name, $2) > 0.15
		  )
	`
	var total int64
	_ = r.db.QueryRow(ctx, countQuery, searchPattern, cleanQuery).Scan(&total)

	return results, total, nil
}

func (r *pgxSearchRepository) SearchPosts(ctx context.Context, viewerID uuid.UUID, query string, limit, offset int) ([]domain.Post, int64, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	cleanQuery := strings.TrimSpace(query)
	if cleanQuery == "" {
		return []domain.Post{}, 0, nil
	}
	if limit <= 0 {
		limit = 20
	}

	searchPattern := "%" + cleanQuery + "%"

	sqlQuery := `
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
		       EXISTS(SELECT 1 FROM likes l2 WHERE l2.post_id = p.id AND l2.personnel_id = $1 AND l2.deleted_at IS NULL) as is_liked,
		       EXISTS(SELECT 1 FROM saved_posts sp WHERE sp.post_id = p.id AND sp.personnel_id = $1) as is_saved,
		       p.created_at, p.updated_at,
		       (
		           CASE
		               WHEN p.caption ILIKE $2 THEN 1.0
		               WHEN p.search_vector @@ plainto_tsquery('english', $3) THEN ts_rank(p.search_vector, plainto_tsquery('english', $3))
		               ELSE similarity(p.caption, $3)
		           END
		       ) as rank
		FROM posts p
		JOIN personnel per ON p.personnel_id = per.id
		JOIN users u ON per.user_id = u.id
		WHERE p.deleted_at IS NULL
		  AND (p.visibility = 'PUBLIC' OR p.personnel_id = $1)
		  AND (p.status = 'PUBLISHED' OR p.status IS NULL)
		  AND (p.is_archived = false OR p.is_archived IS NULL)
		  AND (
		      p.caption ILIKE $2
		      OR p.search_vector @@ plainto_tsquery('english', $3)
		      OR similarity(p.caption, $3) > 0.1
		  )
		ORDER BY rank DESC, p.created_at DESC
		LIMIT $4 OFFSET $5
	`

	rows, err := r.db.Query(ctx, sqlQuery, viewerID, searchPattern, cleanQuery, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var posts []domain.Post
	for rows.Next() {
		var p domain.Post
		var rank float64
		if err := rows.Scan(
			&p.ID, &p.PersonnelID, &p.Caption, &p.MediaURL, &p.Location, &p.Visibility,
			&p.Status, &p.IsArchived, &p.ViewCount, &p.ShareCount,
			&p.Likes, &p.Comments,
			&p.CreatorUsername, &p.CreatorName, &p.CreatorAvatar,
			&p.IsLiked, &p.IsSaved,
			&p.CreatedAt, &p.UpdatedAt,
			&rank,
		); err != nil {
			return nil, 0, err
		}
		_ = r.enrichPostMedia(ctx, &p)
		posts = append(posts, p)
	}

	countQuery := `
		SELECT COUNT(*)
		FROM posts p
		WHERE p.deleted_at IS NULL
		  AND (p.visibility = 'PUBLIC' OR p.personnel_id = $1)
		  AND (p.status = 'PUBLISHED' OR p.status IS NULL)
		  AND (p.is_archived = false OR p.is_archived IS NULL)
		  AND (
		      p.caption ILIKE $2
		      OR p.search_vector @@ plainto_tsquery('english', $3)
		      OR similarity(p.caption, $3) > 0.1
		  )
	`
	var total int64
	_ = r.db.QueryRow(ctx, countQuery, viewerID, searchPattern, cleanQuery).Scan(&total)

	return posts, total, nil
}

func (r *pgxSearchRepository) SearchHashtags(ctx context.Context, query string, limit, offset int) ([]domain.HashtagSearchResult, int64, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	cleanQuery := strings.TrimPrefix(strings.TrimSpace(query), "#")
	if cleanQuery == "" {
		return []domain.HashtagSearchResult{}, 0, nil
	}
	if limit <= 0 {
		limit = 20
	}

	searchPattern := "%" + cleanQuery + "%"
	prefixPattern := cleanQuery + "%"

	sqlQuery := `
		SELECT h.id,
		       h.name,
		       COUNT(ph.post_id) as post_count,
		       (
		           CASE
		               WHEN LOWER(h.name) = LOWER($1) THEN 1.0
		               WHEN LOWER(h.name) LIKE LOWER($2) THEN 0.8
		               ELSE similarity(h.name, $1)
		           END
		       ) as match_score
		FROM hashtags h
		LEFT JOIN post_hashtags ph ON h.id = ph.hashtag_id
		WHERE h.deleted_at IS NULL
		  AND (
		      h.name ILIKE $3
		      OR similarity(h.name, $1) > 0.15
		  )
		GROUP BY h.id, h.name
		ORDER BY match_score DESC, post_count DESC
		LIMIT $4 OFFSET $5
	`

	rows, err := r.db.Query(ctx, sqlQuery, cleanQuery, prefixPattern, searchPattern, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var results []domain.HashtagSearchResult
	for rows.Next() {
		var h domain.HashtagSearchResult
		if err := rows.Scan(&h.ID, &h.Name, &h.PostCount, &h.MatchScore); err != nil {
			return nil, 0, err
		}
		results = append(results, h)
	}

	countQuery := `
		SELECT COUNT(DISTINCT h.id)
		FROM hashtags h
		WHERE h.deleted_at IS NULL
		  AND (h.name ILIKE $1 OR similarity(h.name, $2) > 0.15)
	`
	var total int64
	_ = r.db.QueryRow(ctx, countQuery, searchPattern, cleanQuery).Scan(&total)

	return results, total, nil
}

func (r *pgxSearchRepository) GetPostsByHashtag(ctx context.Context, viewerID uuid.UUID, tag string, limit, offset int) ([]domain.Post, int64, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	cleanTag := strings.TrimPrefix(strings.ToLower(strings.TrimSpace(tag)), "#")
	if cleanTag == "" {
		return []domain.Post{}, 0, nil
	}
	if limit <= 0 {
		limit = 20
	}

	sqlQuery := `
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
		       EXISTS(SELECT 1 FROM likes l2 WHERE l2.post_id = p.id AND l2.personnel_id = $1 AND l2.deleted_at IS NULL) as is_liked,
		       EXISTS(SELECT 1 FROM saved_posts sp WHERE sp.post_id = p.id AND sp.personnel_id = $1) as is_saved,
		       p.created_at, p.updated_at
		FROM posts p
		JOIN post_hashtags ph ON p.id = ph.post_id
		JOIN hashtags h ON ph.hashtag_id = h.id
		JOIN personnel per ON p.personnel_id = per.id
		JOIN users u ON per.user_id = u.id
		WHERE LOWER(h.name) = $2
		  AND p.deleted_at IS NULL
		  AND (p.visibility = 'PUBLIC' OR p.personnel_id = $1)
		  AND (p.status = 'PUBLISHED' OR p.status IS NULL)
		  AND (p.is_archived = false OR p.is_archived IS NULL)
		ORDER BY p.created_at DESC
		LIMIT $3 OFFSET $4
	`

	rows, err := r.db.Query(ctx, sqlQuery, viewerID, cleanTag, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var posts []domain.Post
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
		_ = r.enrichPostMedia(ctx, &p)
		posts = append(posts, p)
	}

	countQuery := `
		SELECT COUNT(*)
		FROM posts p
		JOIN post_hashtags ph ON p.id = ph.post_id
		JOIN hashtags h ON ph.hashtag_id = h.id
		WHERE LOWER(h.name) = $1
		  AND p.deleted_at IS NULL
		  AND (p.visibility = 'PUBLIC' OR p.personnel_id = $2)
		  AND (p.status = 'PUBLISHED' OR p.status IS NULL)
		  AND (p.is_archived = false OR p.is_archived IS NULL)
	`
	var total int64
	_ = r.db.QueryRow(ctx, countQuery, cleanTag, viewerID).Scan(&total)

	return posts, total, nil
}

func (r *pgxSearchRepository) GetHashtagDetails(ctx context.Context, name string) (*domain.HashtagSearchResult, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	cleanTag := strings.TrimPrefix(strings.ToLower(strings.TrimSpace(name)), "#")

	query := `
		SELECT h.id, h.name, COUNT(ph.post_id) as post_count
		FROM hashtags h
		LEFT JOIN post_hashtags ph ON h.id = ph.hashtag_id
		WHERE LOWER(h.name) = $1 AND h.deleted_at IS NULL
		GROUP BY h.id, h.name
	`
	var res domain.HashtagSearchResult
	err := r.db.QueryRow(ctx, query, cleanTag).Scan(&res.ID, &res.Name, &res.PostCount)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("hashtag not found")
		}
		return nil, err
	}
	return &res, nil
}

func (r *pgxSearchRepository) enrichPostMedia(ctx context.Context, post *domain.Post) error {
	if post == nil {
		return nil
	}
	query := `
		SELECT id, post_id, owner_id, bucket, object_key, original_filename, mime_type,
		       file_size, media_type, width, height, duration, thumbnail_key, status, display_order,
		       created_at, updated_at
		FROM post_media
		WHERE post_id = $1
		ORDER BY display_order ASC
	`
	rows, err := r.db.Query(ctx, query, post.ID)
	if err != nil {
		return nil
	}
	defer rows.Close()

	for rows.Next() {
		var m domain.PostMedia
		var thumbKey, origName *string
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
			post.Media = append(post.Media, m)
		}
	}
	return nil
}
