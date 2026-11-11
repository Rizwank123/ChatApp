package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisClient wraps redis.Client and provides helper operations for caching,
// Redis Sets, Sorted Sets (for feeds), and RedisBloom (Bloom & Cuckoo Filters).
type RedisClient struct {
	client *redis.Client
}

// NewRedisClient creates and validates a new Redis connection.
func NewRedisClient(addr string, password string, db int) (*RedisClient, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:         addr,
		Password:     password,
		DB:           db,
		PoolSize:     50,
		MinIdleConns: 10,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to redis: %w", err)
	}

	return &RedisClient{client: rdb}, nil
}

// Client returns the underlying redis.Client.
func (r *RedisClient) Client() *redis.Client {
	return r.client
}

// Close gracefully closes the client connection.
func (r *RedisClient) Close() error {
	return r.client.Close()
}

// -------------------------------------------------------------
// Feed Timeline Caching (Redis Sorted Set - ZSET)
// -------------------------------------------------------------

// AddPostToFeed pushes a post ID into a user's feed timeline with created_at as score.
// Automatically caps the inbox at maxItems to keep memory bounded.
func (r *RedisClient) AddPostToFeed(ctx context.Context, userID string, postID string, score float64, maxItems int64) error {
	key := fmt.Sprintf("feed:%s", userID)
	pipe := r.client.Pipeline()
	pipe.ZAdd(ctx, key, redis.Z{Score: score, Member: postID})
	// Keep only the top maxItems posts
	if maxItems > 0 {
		pipe.ZRemRangeByRank(ctx, key, 0, -(maxItems + 1))
	}
	pipe.Expire(ctx, key, 14*24*time.Hour) // 14-day inactivity TTL
	_, err := pipe.Exec(ctx)
	return err
}

// GetFeedPostIDs retrieves cursor-paginated post IDs from the user's sorted set.
func (r *RedisClient) GetFeedPostIDs(ctx context.Context, userID string, maxScore float64, limit int64) ([]string, error) {
	key := fmt.Sprintf("feed:%s", userID)
	opt := &redis.ZRangeBy{
		Max:    fmt.Sprintf("(%f", maxScore), // exclusive if cursor provided
		Min:    "-inf",
		Offset: 0,
		Count:  limit,
	}
	if maxScore <= 0 {
		opt.Max = "+inf"
	}
	return r.client.ZRevRangeByScore(ctx, key, opt).Result()
}

// RemovePostFromFeed removes a post from a feed inbox.
func (r *RedisClient) RemovePostFromFeed(ctx context.Context, userID string, postID string) error {
	key := fmt.Sprintf("feed:%s", userID)
	return r.client.ZRem(ctx, key, postID).Err()
}

// -------------------------------------------------------------
// Cuckoo Filter (RedisBloom: CF.ADD, CF.EXISTS, CF.DEL)
// Used for dynamic feed seen/unseen and post hide/unhide with native deletion support.
// -------------------------------------------------------------

// CuckooAdd adds an item to a cuckoo filter.
func (r *RedisClient) CuckooAdd(ctx context.Context, filterKey string, item string) error {
	return r.client.Do(ctx, "CF.ADD", filterKey, item).Err()
}

// CuckooAddNX adds an item to a cuckoo filter only if it does not already exist.
func (r *RedisClient) CuckooAddNX(ctx context.Context, filterKey string, item string) (bool, error) {
	res, err := r.client.Do(ctx, "CF.ADDNX", filterKey, item).Int()
	return res == 1, err
}

// CuckooExists checks if an item might exist in the cuckoo filter.
func (r *RedisClient) CuckooExists(ctx context.Context, filterKey string, item string) (bool, error) {
	res, err := r.client.Do(ctx, "CF.EXISTS", filterKey, item).Int()
	return res == 1, err
}

// CuckooDel removes an item from a cuckoo filter (unlike standard Bloom Filter, Cuckoo supports deletions!).
func (r *RedisClient) CuckooDel(ctx context.Context, filterKey string, item string) (bool, error) {
	res, err := r.client.Do(ctx, "CF.DEL", filterKey, item).Int()
	return res == 1, err
}

// -------------------------------------------------------------
// Bloom Filter (RedisBloom: BF.ADD, BF.EXISTS)
// Used for append-only static checks, like username reservation.
// -------------------------------------------------------------

// BloomAdd adds an item to a standard Bloom filter.
func (r *RedisClient) BloomAdd(ctx context.Context, filterKey string, item string) error {
	return r.client.Do(ctx, "BF.ADD", filterKey, item).Err()
}

// BloomExists checks if an item exists in a standard Bloom filter.
func (r *RedisClient) BloomExists(ctx context.Context, filterKey string, item string) (bool, error) {
	res, err := r.client.Do(ctx, "BF.EXISTS", filterKey, item).Int()
	return res == 1, err
}
