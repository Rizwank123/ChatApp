package repository

import (
	"context"
	"strings"

	"github.com/gofrs/uuid/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Rizwank123/chatApp/internal/domain"
)

type pgxSocialGraphRepository struct {
	db *pgxpool.Pool
}

// NewSocialGraphRepository creates a new SocialGraphRepository instance.
func NewSocialGraphRepository(db *pgxpool.Pool) domain.SocialGraphRepository {
	return &pgxSocialGraphRepository{db: db}
}

func (r *pgxSocialGraphRepository) GetUserCounters(ctx context.Context, personnelID uuid.UUID) (*domain.UserCounters, error) {
	q := `SELECT personnel_id, follower_count, following_count, post_count, updated_at
	      FROM user_counters WHERE personnel_id = $1`
	var c domain.UserCounters
	err := r.db.QueryRow(ctx, q, personnelID).Scan(
		&c.PersonnelID, &c.FollowerCount, &c.FollowingCount, &c.PostCount, &c.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		// Initialize empty counters if missing
		initQ := `INSERT INTO user_counters (personnel_id, follower_count, following_count, post_count)
		          VALUES ($1, 0, 0, 0)
		          ON CONFLICT (personnel_id) DO NOTHING
		          RETURNING personnel_id, follower_count, following_count, post_count, updated_at`
		errInit := r.db.QueryRow(ctx, initQ, personnelID).Scan(
			&c.PersonnelID, &c.FollowerCount, &c.FollowingCount, &c.PostCount, &c.UpdatedAt,
		)
		if errInit != nil && errInit != pgx.ErrNoRows {
			return nil, errInit
		}
		c.PersonnelID = personnelID
		return &c, nil
	}
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *pgxSocialGraphRepository) IncrementFollowerCount(ctx context.Context, personnelID uuid.UUID, delta int64) error {
	q := `INSERT INTO user_counters (personnel_id, follower_count, following_count, post_count, updated_at)
	      VALUES ($1, GREATEST($2, 0), 0, 0, NOW())
	      ON CONFLICT (personnel_id) DO UPDATE
	      SET follower_count = GREATEST(user_counters.follower_count + $2, 0), updated_at = NOW()`
	_, err := r.db.Exec(ctx, q, personnelID, delta)
	return err
}

func (r *pgxSocialGraphRepository) IncrementFollowingCount(ctx context.Context, personnelID uuid.UUID, delta int64) error {
	q := `INSERT INTO user_counters (personnel_id, follower_count, following_count, post_count, updated_at)
	      VALUES ($1, 0, GREATEST($2, 0), 0, NOW())
	      ON CONFLICT (personnel_id) DO UPDATE
	      SET following_count = GREATEST(user_counters.following_count + $2, 0), updated_at = NOW()`
	_, err := r.db.Exec(ctx, q, personnelID, delta)
	return err
}

func (r *pgxSocialGraphRepository) IncrementPostCount(ctx context.Context, personnelID uuid.UUID, delta int64) error {
	q := `INSERT INTO user_counters (personnel_id, follower_count, following_count, post_count, updated_at)
	      VALUES ($1, 0, 0, GREATEST($2, 0), NOW())
	      ON CONFLICT (personnel_id) DO UPDATE
	      SET post_count = GREATEST(user_counters.post_count + $2, 0), updated_at = NOW()`
	_, err := r.db.Exec(ctx, q, personnelID, delta)
	return err
}

func (r *pgxSocialGraphRepository) BlockUser(ctx context.Context, blockerID, blockedID uuid.UUID, reason string) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// Insert block record
	blockQ := `INSERT INTO blocks (blocker_id, blocked_id, reason) VALUES ($1, $2, $3)
	           ON CONFLICT (blocker_id, blocked_id) DO NOTHING`
	if _, err := tx.Exec(ctx, blockQ, blockerID, blockedID, reason); err != nil {
		return err
	}

	// Sever existing follow relationships both ways
	delFollowQ := `DELETE FROM follows WHERE (follower_id = $1 AND following_id = $2)
	                                    OR (follower_id = $2 AND following_id = $1)`
	if _, err := tx.Exec(ctx, delFollowQ, blockerID, blockedID); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (r *pgxSocialGraphRepository) UnblockUser(ctx context.Context, blockerID, blockedID uuid.UUID) error {
	q := `DELETE FROM blocks WHERE blocker_id = $1 AND blocked_id = $2`
	_, err := r.db.Exec(ctx, q, blockerID, blockedID)
	return err
}

func (r *pgxSocialGraphRepository) IsBlocked(ctx context.Context, userA, userB uuid.UUID) (bool, error) {
	q := `SELECT EXISTS (
		SELECT 1 FROM blocks WHERE (blocker_id = $1 AND blocked_id = $2)
		                        OR (blocker_id = $2 AND blocked_id = $1)
	)`
	var blocked bool
	err := r.db.QueryRow(ctx, q, userA, userB).Scan(&blocked)
	return blocked, err
}

func (r *pgxSocialGraphRepository) GetBlockedUsers(ctx context.Context, blockerID uuid.UUID) ([]uuid.UUID, error) {
	q := `SELECT blocked_id FROM blocks WHERE blocker_id = $1`
	rows, err := r.db.Query(ctx, q, blockerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var blockedIDs []uuid.UUID
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		blockedIDs = append(blockedIDs, id)
	}
	return blockedIDs, nil
}

func (r *pgxSocialGraphRepository) CreateFollowRequest(ctx context.Context, req *domain.FollowRequest) error {
	q := `INSERT INTO follow_requests (requester_id, requested_id, status)
	      VALUES ($1, $2, $3)
	      ON CONFLICT (requester_id, requested_id) DO UPDATE
	      SET status = $3, updated_at = NOW()
	      RETURNING id, created_at, updated_at`
	return r.db.QueryRow(ctx, q, req.RequesterID, req.RequestedID, req.Status).Scan(
		&req.ID, &req.CreatedAt, &req.UpdatedAt,
	)
}

func (r *pgxSocialGraphRepository) UpdateFollowRequestStatus(ctx context.Context, reqID uuid.UUID, status domain.FollowRequestStatus) error {
	q := `UPDATE follow_requests SET status = $1, updated_at = NOW() WHERE id = $2`
	_, err := r.db.Exec(ctx, q, status, reqID)
	return err
}

func (r *pgxSocialGraphRepository) GetPendingFollowRequests(ctx context.Context, requestedID uuid.UUID) ([]domain.FollowRequest, error) {
	q := `SELECT id, requester_id, requested_id, status, created_at, updated_at
	      FROM follow_requests
	      WHERE requested_id = $1 AND status = 'PENDING'
	      ORDER BY created_at DESC`
	rows, err := r.db.Query(ctx, q, requestedID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []domain.FollowRequest
	for rows.Next() {
		var item domain.FollowRequest
		if err := rows.Scan(&item.ID, &item.RequesterID, &item.RequestedID, &item.Status, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return nil, err
		}
		list = append(list, item)
	}
	return list, nil
}

func (r *pgxSocialGraphRepository) SavePost(ctx context.Context, personnelID, postID uuid.UUID) error {
	q := `INSERT INTO saved_posts (personnel_id, post_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`
	_, err := r.db.Exec(ctx, q, personnelID, postID)
	return err
}

func (r *pgxSocialGraphRepository) UnsavePost(ctx context.Context, personnelID, postID uuid.UUID) error {
	q := `DELETE FROM saved_posts WHERE personnel_id = $1 AND post_id = $2`
	_, err := r.db.Exec(ctx, q, personnelID, postID)
	return err
}

func (r *pgxSocialGraphRepository) GetSavedPosts(ctx context.Context, personnelID uuid.UUID, limit, offset int) ([]domain.Post, int64, error) {
	countQ := `SELECT COUNT(*) FROM saved_posts WHERE personnel_id = $1`
	var total int64
	if err := r.db.QueryRow(ctx, countQ, personnelID).Scan(&total); err != nil {
		return nil, 0, err
	}

	q := `SELECT p.id, p.personnel_id, p.caption, p.media_url, p.location, p.visibility, p.created_at, p.updated_at
	      FROM saved_posts s
	      JOIN posts p ON s.post_id = p.id
	      WHERE s.personnel_id = $1 AND p.deleted_at IS NULL
	      ORDER BY s.created_at DESC
	      LIMIT $2 OFFSET $3`
	rows, err := r.db.Query(ctx, q, personnelID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var posts []domain.Post
	for rows.Next() {
		var p domain.Post
		if err := rows.Scan(&p.ID, &p.PersonnelID, &p.Caption, &p.MediaURL, &p.Location, &p.Visibility, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, 0, err
		}
		posts = append(posts, p)
	}
	return posts, total, nil
}

func (r *pgxSocialGraphRepository) GetFollowerIDs(ctx context.Context, personnelID uuid.UUID, limit, offset int) ([]uuid.UUID, error) {
	q := `SELECT follower_id FROM follows WHERE following_id = $1 AND deleted_at IS NULL
	      ORDER BY created_at DESC LIMIT $2 OFFSET $3`
	rows, err := r.db.Query(ctx, q, personnelID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var followers []uuid.UUID
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		followers = append(followers, id)
	}
	return followers, nil
}

func (r *pgxSocialGraphRepository) Follow(ctx context.Context, followerID, followingID uuid.UUID) error {
	q := `INSERT INTO follows (follower_id, following_id, created_at, updated_at, deleted_at)
	      VALUES ($1, $2, NOW(), NOW(), NULL)
	      ON CONFLICT (follower_id, following_id)
	      DO UPDATE SET deleted_at = NULL, updated_at = NOW()`
	_, err := r.db.Exec(ctx, q, followerID, followingID)
	return err
}

func (r *pgxSocialGraphRepository) Unfollow(ctx context.Context, followerID, followingID uuid.UUID) error {
	q := `UPDATE follows SET deleted_at = NOW() WHERE follower_id = $1 AND following_id = $2 AND deleted_at IS NULL`
	_, err := r.db.Exec(ctx, q, followerID, followingID)
	return err
}

func (r *pgxSocialGraphRepository) IsFollowing(ctx context.Context, followerID, followingID uuid.UUID) (bool, error) {
	q := `SELECT EXISTS(SELECT 1 FROM follows WHERE follower_id = $1 AND following_id = $2 AND deleted_at IS NULL)`
	var exists bool
	err := r.db.QueryRow(ctx, q, followerID, followingID).Scan(&exists)
	return exists, err
}

func (r *pgxSocialGraphRepository) IsMutualFollow(ctx context.Context, userA, userB uuid.UUID) (bool, error) {
	q := `SELECT (
		EXISTS(SELECT 1 FROM follows WHERE follower_id = $1 AND following_id = $2 AND deleted_at IS NULL)
		AND
		EXISTS(SELECT 1 FROM follows WHERE follower_id = $2 AND following_id = $1 AND deleted_at IS NULL)
	)`
	var isMutual bool
	err := r.db.QueryRow(ctx, q, userA, userB).Scan(&isMutual)
	return isMutual, err
}

func (r *pgxSocialGraphRepository) CanChat(ctx context.Context, userA, userB uuid.UUID) (bool, error) {
	if userA == userB {
		return false, nil
	}

	q := `SELECT
		(SELECT COUNT(*) FROM personnel p
		 JOIN users u ON p.user_id = u.id
		 WHERE p.id IN ($1, $2)
		   AND p.deleted_at IS NULL
		   AND u.deleted_at IS NULL
		   AND p.activation_status = 'ACTIVE'
		) = 2 AS active_accounts,
		EXISTS(
			SELECT 1 FROM blocks
			WHERE (blocker_id = $1 AND blocked_id = $2)
			   OR (blocker_id = $2 AND blocked_id = $1)
		) AS is_blocked,
		(
			EXISTS(SELECT 1 FROM follows WHERE follower_id = $1 AND following_id = $2 AND deleted_at IS NULL)
			OR
			EXISTS(SELECT 1 FROM follows WHERE follower_id = $2 AND following_id = $1 AND deleted_at IS NULL)
		) AS has_follow_relationship`

	var activeAccounts, isBlocked, hasFollowRelationship bool
	err := r.db.QueryRow(ctx, q, userA, userB).Scan(&activeAccounts, &isBlocked, &hasFollowRelationship)
	if err != nil {
		return false, err
	}

	return activeAccounts && !isBlocked && hasFollowRelationship, nil
}

func (r *pgxSocialGraphRepository) GetRelationship(ctx context.Context, currentPersonnelID, targetPersonnelID uuid.UUID) (*domain.UserRelationship, error) {
	rel := &domain.UserRelationship{
		TargetPersonnelID: targetPersonnelID,
	}
	if currentPersonnelID == targetPersonnelID {
		return rel, nil
	}

	q := `SELECT
		EXISTS(SELECT 1 FROM follows WHERE follower_id = $1 AND following_id = $2 AND deleted_at IS NULL) AS is_following,
		EXISTS(SELECT 1 FROM follows WHERE follower_id = $2 AND following_id = $1 AND deleted_at IS NULL) AS follows_you,
		EXISTS(SELECT 1 FROM blocks WHERE (blocker_id = $1 AND blocked_id = $2) OR (blocker_id = $2 AND blocked_id = $1)) AS is_blocked,
		(SELECT COUNT(*) FROM personnel p
		 JOIN users u ON p.user_id = u.id
		 WHERE p.id IN ($1, $2) AND p.deleted_at IS NULL AND u.deleted_at IS NULL AND p.activation_status = 'ACTIVE'
		) = 2 AS active_accounts`

	var isFollowing, followsYou, isBlocked, activeAccounts bool
	err := r.db.QueryRow(ctx, q, currentPersonnelID, targetPersonnelID).Scan(&isFollowing, &followsYou, &isBlocked, &activeAccounts)
	if err != nil {
		return nil, err
	}

	rel.IsFollowing = isFollowing
	rel.FollowsYou = followsYou
	rel.IsMutual = isFollowing && followsYou
	rel.CanChat = (isFollowing || followsYou) && !isBlocked && activeAccounts

	return rel, nil
}

func (r *pgxSocialGraphRepository) GetFollowersList(ctx context.Context, currentPersonnelID, targetPersonnelID uuid.UUID, query string, limit, offset int) ([]domain.FollowUserItem, int64, error) {
	if limit <= 0 {
		limit = 20
	}

	term := "%" + strings.ToLower(strings.TrimSpace(query)) + "%"
	var countQ string
	var countArgs []any
	if strings.TrimSpace(query) != "" {
		countQ = `SELECT COUNT(*) FROM follows f
		          JOIN personnel p ON f.follower_id = p.id
		          JOIN users u ON p.user_id = u.id
		          WHERE f.following_id = $1 AND f.deleted_at IS NULL AND p.deleted_at IS NULL AND u.deleted_at IS NULL
		            AND (LOWER(u.user_name) LIKE $2 OR LOWER(COALESCE(p.display_name, '')) LIKE $2 OR LOWER(p.first_name) LIKE $2 OR LOWER(p.last_name) LIKE $2)`
		countArgs = []any{targetPersonnelID, term}
	} else {
		countQ = `SELECT COUNT(*) FROM follows WHERE following_id = $1 AND deleted_at IS NULL`
		countArgs = []any{targetPersonnelID}
	}

	var total int64
	if err := r.db.QueryRow(ctx, countQ, countArgs...).Scan(&total); err != nil {
		return nil, 0, err
	}

	var q string
	var args []any
	if strings.TrimSpace(query) != "" {
		q = `SELECT
			p.id, p.user_id, u.user_name, COALESCE(p.display_name, TRIM(CONCAT(p.first_name, ' ', p.last_name))) AS display_name,
			p.avatar, p.bio,
			EXISTS(SELECT 1 FROM follows f1 WHERE f1.follower_id = $1 AND f1.following_id = p.id AND f1.deleted_at IS NULL) AS is_following,
			EXISTS(SELECT 1 FROM follows f2 WHERE f2.follower_id = p.id AND f2.following_id = $1 AND f2.deleted_at IS NULL) AS follows_you,
			f.created_at
		FROM follows f
		JOIN personnel p ON f.follower_id = p.id
		JOIN users u ON p.user_id = u.id
		WHERE f.following_id = $2 AND f.deleted_at IS NULL AND p.deleted_at IS NULL AND u.deleted_at IS NULL
		  AND (LOWER(u.user_name) LIKE $3 OR LOWER(COALESCE(p.display_name, '')) LIKE $3 OR LOWER(p.first_name) LIKE $3 OR LOWER(p.last_name) LIKE $3)
		ORDER BY f.created_at DESC
		LIMIT $4 OFFSET $5`
		args = []any{currentPersonnelID, targetPersonnelID, term, limit, offset}
	} else {
		q = `SELECT
			p.id, p.user_id, u.user_name, COALESCE(p.display_name, TRIM(CONCAT(p.first_name, ' ', p.last_name))) AS display_name,
			p.avatar, p.bio,
			EXISTS(SELECT 1 FROM follows f1 WHERE f1.follower_id = $1 AND f1.following_id = p.id AND f1.deleted_at IS NULL) AS is_following,
			EXISTS(SELECT 1 FROM follows f2 WHERE f2.follower_id = p.id AND f2.following_id = $1 AND f2.deleted_at IS NULL) AS follows_you,
			f.created_at
		FROM follows f
		JOIN personnel p ON f.follower_id = p.id
		JOIN users u ON p.user_id = u.id
		WHERE f.following_id = $2 AND f.deleted_at IS NULL AND p.deleted_at IS NULL AND u.deleted_at IS NULL
		ORDER BY f.created_at DESC
		LIMIT $3 OFFSET $4`
		args = []any{currentPersonnelID, targetPersonnelID, limit, offset}
	}

	rows, err := r.db.Query(ctx, q, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var list []domain.FollowUserItem
	for rows.Next() {
		var item domain.FollowUserItem
		if err := rows.Scan(
			&item.PersonnelID, &item.UserID, &item.UserName, &item.DisplayName,
			&item.Avatar, &item.Bio, &item.IsFollowing, &item.FollowsYou, &item.CreatedAt,
		); err != nil {
			return nil, 0, err
		}
		item.IsMutual = item.IsFollowing && item.FollowsYou
		list = append(list, item)
	}

	return list, total, nil
}

func (r *pgxSocialGraphRepository) GetFollowingList(ctx context.Context, currentPersonnelID, targetPersonnelID uuid.UUID, query string, limit, offset int) ([]domain.FollowUserItem, int64, error) {
	if limit <= 0 {
		limit = 20
	}

	term := "%" + strings.ToLower(strings.TrimSpace(query)) + "%"
	var countQ string
	var countArgs []any
	if strings.TrimSpace(query) != "" {
		countQ = `SELECT COUNT(*) FROM follows f
		          JOIN personnel p ON f.following_id = p.id
		          JOIN users u ON p.user_id = u.id
		          WHERE f.follower_id = $1 AND f.deleted_at IS NULL AND p.deleted_at IS NULL AND u.deleted_at IS NULL
		            AND (LOWER(u.user_name) LIKE $2 OR LOWER(COALESCE(p.display_name, '')) LIKE $2 OR LOWER(p.first_name) LIKE $2 OR LOWER(p.last_name) LIKE $2)`
		countArgs = []any{targetPersonnelID, term}
	} else {
		countQ = `SELECT COUNT(*) FROM follows WHERE follower_id = $1 AND deleted_at IS NULL`
		countArgs = []any{targetPersonnelID}
	}

	var total int64
	if err := r.db.QueryRow(ctx, countQ, countArgs...).Scan(&total); err != nil {
		return nil, 0, err
	}

	var q string
	var args []any
	if strings.TrimSpace(query) != "" {
		q = `SELECT
			p.id, p.user_id, u.user_name, COALESCE(p.display_name, TRIM(CONCAT(p.first_name, ' ', p.last_name))) AS display_name,
			p.avatar, p.bio,
			EXISTS(SELECT 1 FROM follows f1 WHERE f1.follower_id = $1 AND f1.following_id = p.id AND f1.deleted_at IS NULL) AS is_following,
			EXISTS(SELECT 1 FROM follows f2 WHERE f2.follower_id = p.id AND f2.following_id = $1 AND f2.deleted_at IS NULL) AS follows_you,
			f.created_at
		FROM follows f
		JOIN personnel p ON f.following_id = p.id
		JOIN users u ON p.user_id = u.id
		WHERE f.follower_id = $2 AND f.deleted_at IS NULL AND p.deleted_at IS NULL AND u.deleted_at IS NULL
		  AND (LOWER(u.user_name) LIKE $3 OR LOWER(COALESCE(p.display_name, '')) LIKE $3 OR LOWER(p.first_name) LIKE $3 OR LOWER(p.last_name) LIKE $3)
		ORDER BY f.created_at DESC
		LIMIT $4 OFFSET $5`
		args = []any{currentPersonnelID, targetPersonnelID, term, limit, offset}
	} else {
		q = `SELECT
			p.id, p.user_id, u.user_name, COALESCE(p.display_name, TRIM(CONCAT(p.first_name, ' ', p.last_name))) AS display_name,
			p.avatar, p.bio,
			EXISTS(SELECT 1 FROM follows f1 WHERE f1.follower_id = $1 AND f1.following_id = p.id AND f1.deleted_at IS NULL) AS is_following,
			EXISTS(SELECT 1 FROM follows f2 WHERE f2.follower_id = p.id AND f2.following_id = $1 AND f2.deleted_at IS NULL) AS follows_you,
			f.created_at
		FROM follows f
		JOIN personnel p ON f.following_id = p.id
		JOIN users u ON p.user_id = u.id
		WHERE f.follower_id = $2 AND f.deleted_at IS NULL AND p.deleted_at IS NULL AND u.deleted_at IS NULL
		ORDER BY f.created_at DESC
		LIMIT $3 OFFSET $4`
		args = []any{currentPersonnelID, targetPersonnelID, limit, offset}
	}

	rows, err := r.db.Query(ctx, q, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var list []domain.FollowUserItem
	for rows.Next() {
		var item domain.FollowUserItem
		if err := rows.Scan(
			&item.PersonnelID, &item.UserID, &item.UserName, &item.DisplayName,
			&item.Avatar, &item.Bio, &item.IsFollowing, &item.FollowsYou, &item.CreatedAt,
		); err != nil {
			return nil, 0, err
		}
		item.IsMutual = item.IsFollowing && item.FollowsYou
		list = append(list, item)
	}

	return list, total, nil
}

func (r *pgxSocialGraphRepository) SearchUsers(ctx context.Context, currentPersonnelID uuid.UUID, query string, limit, offset int) ([]domain.UserProfileSearchResult, error) {
	if limit <= 0 {
		limit = 20
	}
	term := "%" + strings.ToLower(strings.TrimSpace(query)) + "%"

	q := `SELECT
		u.id AS user_id, p.id AS personnel_id, u.user_name,
		COALESCE(p.display_name, TRIM(CONCAT(p.first_name, ' ', p.last_name))) AS display_name,
		p.avatar, p.bio,
		EXISTS(SELECT 1 FROM follows f1 WHERE f1.follower_id = $1 AND f1.following_id = p.id AND f1.deleted_at IS NULL) AS is_following,
		EXISTS(SELECT 1 FROM follows f2 WHERE f2.follower_id = p.id AND f2.following_id = $1 AND f2.deleted_at IS NULL) AS follows_you
	FROM users u
	JOIN personnel p ON p.user_id = u.id
	WHERE u.deleted_at IS NULL AND p.deleted_at IS NULL
	  AND (LOWER(u.user_name) LIKE $2 OR LOWER(p.first_name) LIKE $2 OR LOWER(p.last_name) LIKE $2 OR LOWER(COALESCE(p.display_name, '')) LIKE $2)
	ORDER BY
		CASE WHEN LOWER(u.user_name) = LOWER(TRIM($3)) THEN 1
		     WHEN LOWER(u.user_name) LIKE LOWER(TRIM($3)) || '%' THEN 2
		     ELSE 3 END,
		u.created_at DESC
	LIMIT $4 OFFSET $5`

	rows, err := r.db.Query(ctx, q, currentPersonnelID, term, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []domain.UserProfileSearchResult
	for rows.Next() {
		var item domain.UserProfileSearchResult
		if err := rows.Scan(
			&item.UserID, &item.PersonnelID, &item.UserName, &item.DisplayName,
			&item.Avatar, &item.Bio, &item.IsFollowing, &item.FollowsYou,
		); err != nil {
			return nil, err
		}
		item.IsMutual = item.IsFollowing && item.FollowsYou
		results = append(results, item)
	}

	return results, nil
}

func (r *pgxSocialGraphRepository) GetFollowingCelebrityIDs(ctx context.Context, personnelID uuid.UUID, celebrityThreshold int64) ([]uuid.UUID, error) {
	q := `SELECT f.following_id
	      FROM follows f
	      JOIN user_counters uc ON f.following_id = uc.personnel_id
	      WHERE f.follower_id = $1 AND f.deleted_at IS NULL AND uc.follower_count >= $2`
	rows, err := r.db.Query(ctx, q, personnelID, celebrityThreshold)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var celebrities []uuid.UUID
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		celebrities = append(celebrities, id)
	}
	return celebrities, nil
}
