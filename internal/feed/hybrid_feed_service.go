package feed

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Rizwank123/chatApp/internal/cache"
	"github.com/Rizwank123/chatApp/internal/domain"
)

// HybridFeedConfig specifies threshold and capacity parameters.
type HybridFeedConfig struct {
	CelebrityThreshold int64 // Default 25,000 followers
	MaxInboxCapacity   int64 // Max posts retained in Redis ZSET (e.g. 800)
	DefaultPageLimit   int64 // Default posts per page
}

// FeedPostItem represents an item returned in the home feed.
type FeedPostItem struct {
	Post        domain.Post       `json:"post"`
	Author      *domain.Personnel `json:"author,omitempty"`
	IsCelebrity bool              `json:"is_celebrity"`
	Score       float64           `json:"score"`
}

// FeedPage represents a cursor-paginated feed response.
type FeedPage struct {
	Items      []FeedPostItem `json:"items"`
	NextCursor string         `json:"next_cursor,omitempty"`
	HasMore    bool           `json:"has_more"`
}

// HybridFeedEngine orchestrates fanout-on-write and fanout-on-read feed assembly.
type HybridFeedEngine struct {
	db         *pgxpool.Pool
	redis      *cache.RedisClient
	socialRepo domain.SocialGraphRepository
	postRepo   domain.PostRepository
	config     HybridFeedConfig
}

// NewHybridFeedEngine instantiates the hybrid feed engine.
func NewHybridFeedEngine(
	db *pgxpool.Pool,
	rdb *cache.RedisClient,
	socialRepo domain.SocialGraphRepository,
	postRepo domain.PostRepository,
	cfg *HybridFeedConfig,
) *HybridFeedEngine {
	if cfg == nil {
		cfg = &HybridFeedConfig{
			CelebrityThreshold: 25000,
			MaxInboxCapacity:   800,
			DefaultPageLimit:   20,
		}
	}
	return &HybridFeedEngine{
		db:         db,
		redis:      rdb,
		socialRepo: socialRepo,
		postRepo:   postRepo,
		config:     *cfg,
	}
}

// FanoutPost distributes a new post to eligible followers and the author's own timeline.
func (e *HybridFeedEngine) FanoutPost(ctx context.Context, post *domain.Post) error {
	score := float64(post.CreatedAt.UnixMilli())

	// 1. Always add to the author's own timeline in Redis
	if e.redis != nil {
		_ = e.redis.AddPostToFeed(ctx, post.PersonnelID.String(), post.ID.String(), score, e.config.MaxInboxCapacity)
	}

	counters, err := e.socialRepo.GetUserCounters(ctx, post.PersonnelID)
	if err != nil {
		return fmt.Errorf("failed to fetch author counters: %w", err)
	}

	// Fanout-on-Read: Skip writing to millions of follower timelines if celebrity
	if counters.FollowerCount >= e.config.CelebrityThreshold {
		return nil
	}

	// Fanout-on-Write: Batch fetch followers and update their Redis ZSET timelines
	batchSize := 1000
	offset := 0

	for {
		followerIDs, err := e.socialRepo.GetFollowerIDs(ctx, post.PersonnelID, batchSize, offset)
		if err != nil {
			return fmt.Errorf("failed to fetch followers for fanout: %w", err)
		}
		if len(followerIDs) == 0 {
			break
		}

		for _, fid := range followerIDs {
			if e.redis != nil {
				_ = e.redis.AddPostToFeed(ctx, fid.String(), post.ID.String(), score, e.config.MaxInboxCapacity)
			}
		}

		offset += len(followerIDs)
		if len(followerIDs) < batchSize {
			break
		}
	}

	return nil
}

// GetUserFeed retrieves the hybrid home feed combining precomputed inbox + celebrity fanout-on-read.
func (e *HybridFeedEngine) GetUserFeed(ctx context.Context, userID uuid.UUID, cursor string, limit int64) (*FeedPage, error) {
	if e == nil {
		return &FeedPage{Items: []FeedPostItem{}, HasMore: false}, nil
	}

	if limit <= 0 || limit > 50 {
		limit = e.config.DefaultPageLimit
	}

	var maxScore float64 = 0
	if cursor != "" {
		if s, err := strconv.ParseFloat(cursor, 64); err == nil {
			maxScore = s
		}
	}

	postScores := make(map[uuid.UUID]float64)

	// 1. Fetch precomputed candidate post IDs from user's Redis inbox (Fanout-on-Write)
	if e.redis != nil {
		cachedIDs, err := e.redis.GetFeedPostIDs(ctx, userID.String(), maxScore, limit*2)
		if err == nil {
			for _, idStr := range cachedIDs {
				if pid, err := uuid.FromString(idStr); err == nil {
					postScores[pid] = 0 // Will populate from post.CreatedAt
				}
			}
		}
	}

	// 2. Fetch recent posts from followed celebrities (Fanout-on-Read)
	if e.socialRepo != nil {
		celebrityIDs, err := e.socialRepo.GetFollowingCelebrityIDs(ctx, userID, e.config.CelebrityThreshold)
		if err == nil && len(celebrityIDs) > 0 {
			celebPosts, err := e.getRecentCelebrityPosts(ctx, celebrityIDs, maxScore, limit)
			if err == nil {
				for _, cp := range celebPosts {
					postScores[cp.ID] = float64(cp.CreatedAt.UnixMilli())
				}
			}
		}
	}

	// 3. If no cached posts or celebrities found, fall back to relational post query
	if len(postScores) == 0 {
		if e.postRepo != nil {
			posts, _, err := e.postRepo.FindFeedPosts(ctx, userID, domain.QueryOptions{
				Limit:  limit,
				Offset: 0,
			})
			if err != nil {
				return nil, err
			}
			if len(posts) == 0 {
				posts, _, _ = e.postRepo.FindPublicPosts(ctx, domain.QueryOptions{
					Limit:  limit,
					Offset: 0,
				}, userID)
			}
			items := make([]FeedPostItem, len(posts))
			for i, p := range posts {
				items[i] = FeedPostItem{
					Post:  p,
					Score: float64(p.CreatedAt.UnixMilli()),
				}
			}
			return &FeedPage{
				Items:   items,
				HasMore: int64(len(posts)) == limit,
			}, nil
		}
		return &FeedPage{Items: []FeedPostItem{}, HasMore: false}, nil
	}

	candidateIDs := make([]uuid.UUID, 0, len(postScores))
	for id := range postScores {
		candidateIDs = append(candidateIDs, id)
	}

	posts, err := e.hydratePosts(ctx, candidateIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to hydrate posts: %w", err)
	}

	// 4. Sort posts by created_at descending
	sort.Slice(posts, func(i, j int) bool {
		return posts[i].CreatedAt.After(posts[j].CreatedAt)
	})

	// 5. Apply Cuckoo Filter to suppress already seen posts or record seen posts
	cuckooKey := fmt.Sprintf("seen:%s", userID.String())
	var feedItems []FeedPostItem
	for _, p := range posts {
		if e.redis != nil {
			// Check Cuckoo Filter
			seen, _ := e.redis.CuckooExists(ctx, cuckooKey, p.ID.String())
			if !seen {
				// Mark as seen for future fetches
				_ = e.redis.CuckooAdd(ctx, cuckooKey, p.ID.String())
			}
		}

		feedItems = append(feedItems, FeedPostItem{
			Post:  p,
			Score: float64(p.CreatedAt.UnixMilli()),
		})

		if int64(len(feedItems)) >= limit {
			break
		}
	}

	var nextCursor string
	hasMore := false
	if int64(len(posts)) > limit {
		hasMore = true
		lastItem := feedItems[len(feedItems)-1]
		nextCursor = fmt.Sprintf("%.0f", lastItem.Score)
	}

	return &FeedPage{
		Items:      feedItems,
		NextCursor: nextCursor,
		HasMore:    hasMore,
	}, nil
}

func (e *HybridFeedEngine) getRecentCelebrityPosts(ctx context.Context, celebrityIDs []uuid.UUID, maxScore float64, limit int64) ([]domain.Post, error) {
	q := `SELECT p.id, p.personnel_id, p.caption, p.media_url, p.location, p.visibility,
	             COALESCE(p.status, 'PUBLISHED') as status,
	             COALESCE(p.is_archived, false) as is_archived,
	             COALESCE(p.view_count, 0) as view_count,
	             COALESCE(p.share_count, 0) as share_count,
	             u.user_name as creator_username,
	             TRIM(per.first_name || ' ' || per.last_name) as creator_name,
	             COALESCE(per.avatar, '') as creator_avatar,
	             p.created_at, p.updated_at
	      FROM posts p
	      JOIN personnel per ON p.personnel_id = per.id
	      JOIN users u ON per.user_id = u.id
	      WHERE p.personnel_id = ANY($1) AND p.deleted_at IS NULL`
	args := []any{celebrityIDs}

	if maxScore > 0 {
		cutoff := time.UnixMilli(int64(maxScore))
		q += ` AND p.created_at < $2 ORDER BY p.created_at DESC LIMIT $3`
		args = append(args, cutoff, limit)
	} else {
		q += ` ORDER BY p.created_at DESC LIMIT $2`
		args = append(args, limit)
	}

	rows, err := e.db.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []domain.Post
	for rows.Next() {
		var p domain.Post
		if err := rows.Scan(
			&p.ID, &p.PersonnelID, &p.Caption, &p.MediaURL, &p.Location, &p.Visibility,
			&p.Status, &p.IsArchived, &p.ViewCount, &p.ShareCount,
			&p.CreatorUsername, &p.CreatorName, &p.CreatorAvatar,
			&p.CreatedAt, &p.UpdatedAt,
		); err != nil {
			return nil, err
		}
		posts = append(posts, p)
	}
	return posts, nil
}

func (e *HybridFeedEngine) hydratePosts(ctx context.Context, ids []uuid.UUID) ([]domain.Post, error) {
	q := `SELECT p.id, p.personnel_id, p.caption, p.media_url, p.location, p.visibility,
	             COALESCE(p.status, 'PUBLISHED') as status,
	             COALESCE(p.is_archived, false) as is_archived,
	             COALESCE(p.view_count, 0) as view_count,
	             COALESCE(p.share_count, 0) as share_count,
	             COALESCE((SELECT COUNT(*) FROM likes l WHERE l.post_id = p.id AND l.deleted_at IS NULL), 0) as likes,
	             COALESCE((SELECT COUNT(*) FROM comments c WHERE c.post_id = p.id AND c.deleted_at IS NULL), 0) as comments,
	             u.user_name as creator_username,
	             TRIM(per.first_name || ' ' || per.last_name) as creator_name,
	             COALESCE(per.avatar, '') as creator_avatar,
	             p.created_at, p.updated_at
	      FROM posts p
	      JOIN personnel per ON p.personnel_id = per.id
	      JOIN users u ON per.user_id = u.id
	      WHERE p.id = ANY($1) AND p.deleted_at IS NULL`

	rows, err := e.db.Query(ctx, q, ids)
	if err != nil {
		return nil, err
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
			&p.CreatedAt, &p.UpdatedAt,
		); err != nil {
			return nil, err
		}
		posts = append(posts, p)
	}

	// Fetch media for posts
	if len(posts) > 0 && e.db != nil {
		pids := make([]uuid.UUID, len(posts))
		pMap := make(map[uuid.UUID]*domain.Post, len(posts))
		for i := range posts {
			pids[i] = posts[i].ID
			pMap[posts[i].ID] = &posts[i]
		}

		mediaRows, err := e.db.Query(ctx, `
			SELECT id, post_id, owner_id, bucket, object_key, original_filename, mime_type,
			       file_size, media_type, width, height, duration, thumbnail_key, status, display_order,
			       created_at, updated_at
			FROM post_media
			WHERE post_id = ANY($1)
			ORDER BY display_order ASC
		`, pids)
		if err == nil {
			defer mediaRows.Close()
			for mediaRows.Next() {
				var m domain.PostMedia
				var thumbKey, origName *string
				if err := mediaRows.Scan(
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
					if targetPost, ok := pMap[m.PostID]; ok {
						targetPost.Media = append(targetPost.Media, m)
						if len(targetPost.MediaURL) == 0 {
							targetPost.MediaURL = append(targetPost.MediaURL, m.ObjectKey)
						}
					}
				}
			}
		}
	}

	return posts, nil
}
