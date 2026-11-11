# Database Design Document

**Project:** ChatApp - Instagram-Inspired Social Media Platform  
**Date:** November 7, 2026  
**Version:** 2.0 (Enhanced)

---

## Overview

This document describes the complete database schema for ChatApp.

---

## Current Schema Analysis

### Implemented Tables

#### 1. users
```sql
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_name VARCHAR NOT NULL UNIQUE,
    password VARCHAR,
    role USER_ROLE NOT NULL,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ
);
```

#### 2. personnel
```sql
CREATE TABLE personnel (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    first_name VARCHAR NOT NULL,
    last_name VARCHAR NOT NULL,
    gender VARCHAR,
    email VARCHAR,
    mobile VARCHAR NOT NULL,
    avatar VARCHAR,
    address JSONB DEFAULT '{}'::jsonb,
    user_id UUID NOT NULL UNIQUE,
    activation_status VARCHAR NOT NULL,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ
);
```

#### 3. posts (CRITICAL - NEEDS ENHANCEMENT)
```sql
CREATE TABLE posts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    personnel_id UUID NOT NULL,
    caption TEXT,
    media_url TEXT[] NOT NULL,
    location VARCHAR(255),
    visibility VARCHAR(20) NOT NULL DEFAULT 'PUBLIC',
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ
);

-- MISSING: view_count, share_count, media_metadata
```

#### 4. follows (NEEDS INDEXES)
```sql
CREATE TABLE follows (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    follower_id UUID NOT NULL,
    following_id UUID NOT NULL,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ
);

-- CRITICAL: Missing indexes on follower_id and following_id for query performance
```

#### 5. likes
```sql
CREATE TABLE likes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    personnel_id UUID NOT NULL,
    post_id UUID NOT NULL,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ,
    UNIQUE(personnel_id, post_id)
);
```

#### 6. comments (SUPPORTS NESTED REPLIES)
```sql
CREATE TABLE comments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    personnel_id UUID NOT NULL,
    post_id UUID NOT NULL,
    parent_id UUID REFERENCES comments(id),
    content TEXT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ
);
```

#### 7. hashtags & post_hashtags
```sql
CREATE TABLE hashtags (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL UNIQUE,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ
);

CREATE TABLE post_hashtags (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    post_id UUID NOT NULL,
    hashtag_id UUID NOT NULL,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ,
    UNIQUE(post_id, hashtag_id)
);
```

#### 8. friendships, conversations, conversation_participants, messages, message_stat
(Implemented with proper relationships and enums)

---

## Required New Tables (HIGH PRIORITY)

### 14. notifications - User Notifications
```sql
CREATE TYPE notification_type AS ENUM (
    'FOLLOW', 'FOLLOW_REQUEST', 'FOLLOW_ACCEPTED',
    'LIKE', 'COMMENT', 'MENTION', 'SYSTEM'
);

CREATE TABLE notifications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    recipient_id UUID NOT NULL,
    actor_id UUID,
    notification_type notification_type NOT NULL,
    entity_type VARCHAR(50),
    entity_id UUID,
    title VARCHAR(255),
    message TEXT,
    is_read BOOLEAN DEFAULT FALSE,
    read_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ
);

CREATE INDEX idx_notifications_recipient ON notifications(recipient_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_notifications_unread ON notifications(recipient_id) 
    WHERE deleted_at IS NULL AND is_read = FALSE;
```

### 15. saved_posts - Bookmarked Posts
```sql
CREATE TABLE saved_posts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    personnel_id UUID NOT NULL,
    post_id UUID NOT NULL,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(personnel_id, post_id)
);

CREATE INDEX idx_saved_posts_personnel ON saved_posts(personnel_id) WHERE deleted_at IS NULL;
```

### 16. blocks - User Blocking
```sql
CREATE TABLE blocks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    blocker_id UUID NOT NULL,
    blocked_id UUID NOT NULL,
    reason TEXT,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(blocker_id, blocked_id)
);

CREATE INDEX idx_blocks_user ON blocks(blocker_id);
```

### 17. follow_requests - Private Account Requests
```sql
CREATE TYPE follow_request_status AS ENUM ('PENDING', 'ACCEPTED', 'REJECTED');

CREATE TABLE follow_requests (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    requester_id UUID NOT NULL,
    requested_id UUID NOT NULL,
    status follow_request_status DEFAULT 'PENDING',
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(requester_id, requested_id)
);

CREATE INDEX idx_follow_requests_requested ON follow_requests(requested_id, status) 
    WHERE status = 'PENDING';
```

### 18. user_sessions - Session Management
```sql
CREATE TABLE user_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    token_jti VARCHAR(255) UNIQUE,
    refresh_token_hash VARCHAR(255),
    last_activity_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMPTZ NOT NULL,
    revoked_at TIMESTAMPTZ,
    ip_address INET
);

CREATE INDEX idx_sessions_active ON user_sessions(user_id, expires_at) 
    WHERE revoked_at IS NULL;
```

### 19. feed_cache - Precomputed Feeds (CRITICAL FOR SCALE)
```sql
CREATE TABLE feed_cache (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    personnel_id UUID NOT NULL,
    post_id UUID NOT NULL,
    author_id UUID NOT NULL,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(personnel_id, post_id)
);

CREATE INDEX idx_feed_cache_personnel ON feed_cache(personnel_id, created_at DESC);
```

### 20. user_counters - Denormalized Counts
```sql
CREATE TABLE user_counters (
    personnel_id UUID PRIMARY KEY,
    follower_count BIGINT DEFAULT 0,
    following_count BIGINT DEFAULT 0,
    post_count BIGINT DEFAULT 0,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);
```

---

## Required Indexes for Existing Tables (MUST ADD)

```sql
-- Posts indexes
CREATE INDEX idx_posts_personnel_created ON posts(personnel_id, created_at DESC) WHERE deleted_at IS NULL;
CREATE INDEX idx_posts_created ON posts(created_at DESC) WHERE deleted_at IS NULL;

-- Follows indexes (CRITICAL)
CREATE INDEX idx_follows_follower ON follows(follower_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_follows_following ON follows(following_id) WHERE deleted_at IS NULL;

-- Likes indexes
CREATE INDEX idx_likes_post ON likes(post_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_likes_personnel_post ON likes(personnel_id, post_id) WHERE deleted_at IS NULL;

-- Comments indexes
CREATE INDEX idx_comments_post_created ON comments(post_id, created_at DESC) WHERE deleted_at IS NULL;
CREATE INDEX idx_comments_parent ON comments(parent_id) WHERE deleted_at IS NULL AND parent_id IS NOT NULL;

-- Messages indexes
CREATE INDEX idx_messages_conversation_created ON messages(conversation_id, created_at DESC) WHERE deleted_at IS NULL;
```

---

## Missing Columns to Add (HIGH PRIORITY)

```sql
-- posts table
ALTER TABLE posts ADD COLUMN view_count BIGINT DEFAULT 0;
ALTER TABLE posts ADD COLUMN share_count BIGINT DEFAULT 0;
ALTER TABLE posts ADD COLUMN media_metadata JSONB;

-- personnel table
ALTER TABLE personnel ADD COLUMN bio TEXT;
ALTER TABLE personnel ADD COLUMN is_private BOOLEAN DEFAULT FALSE;
ALTER TABLE personnel ADD COLUMN website VARCHAR(255);

-- users table
ALTER TABLE users ADD COLUMN last_login_at TIMESTAMPTZ;
ALTER TABLE users ADD COLUMN failed_login_attempts INT DEFAULT 0;

-- comments table
ALTER TABLE comments ADD COLUMN like_count BIGINT DEFAULT 0;
```

---

## Query Performance Optimization Strategies

### Problem: Naive Feed Query (feed_repository.go)
```sql
-- Current - DOES NOT SCALE
SELECT p.*, (SELECT COUNT(*) FROM likes WHERE post_id = p.id) as likes
FROM posts p
JOIN follows f ON p.personnel_id = f.following_id
WHERE f.follower_id = $1
ORDER BY p.created_at DESC LIMIT 50;

-- Solution: Use feed_cache for fanout-on-write
SELECT f.post_id, p.*, a.first_name, a.last_name, a.avatar
FROM feed_cache f
JOIN posts p ON f.post_id = p.id
JOIN personnel a ON p.personnel_id = a.id
WHERE f.personnel_id = $1 AND p.deleted_at IS NULL
ORDER BY f.created_at DESC LIMIT 50;
```

### Problem: Counter Calculations
```sql
-- Current - SLOW at scale
(SELECT COUNT(*) FROM likes WHERE post_id = $1) as like_count

-- Solution: Denormalized counters
SELECT person_id, like_count FROM user_counters WHERE person_id = $1;
```

---

## Schema Versioning Migration Plan

**Phase 1:** Add missing indexes (30 minutes)  
**Phase 2:** Add new columns to existing tables (45 minutes)  
**Phase 3:** Create high-priority tables: notifications, saved_posts, blocks, feed_cache, user_counters (60 minutes)  
**Phase 4:** Backfill counters and cache for existing data (varies by data size)  
**Phase 5:** Create remaining tables (notifications preferences, follow_requests, sessions) (45 minutes)

---

## Data Retention Policies

```sql
-- Clean up old read notifications after 90 days
DELETE FROM notifications WHERE is_read = TRUE AND created_at < NOW() - INTERVAL '90 days';

-- Remove expired sessions weekly
DELETE FROM user_sessions WHERE revoked_at IS NOT NULL AND revoked_at < NOW() - INTERVAL '30 days';
```

---

## Monitoring Queries

```sql
-- Table sizes
SELECT schemaname, tablename, pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename)) 
FROM pg_tables WHERE schemaname = 'public' ORDER BY pg_total_relation_size(schemaname||'.'||tablename) DESC;

-- Unused indexes
SELECT schemaname, indexname, idx_scan FROM pg_stat_user_indexes WHERE idx_scan = 0;
```

---

## Feed Cache Implementation Details

### Purpose
Enable fast feed generation via precomputed entries instead of complex JOIN queries.

### Schema
```sql
CREATE TABLE feed_cache (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    personnel_id UUID NOT NULL,           -- Who's feed this belongs to
    post_id UUID NOT NULL,                -- Which post
    author_id UUID NOT NULL,              -- Who wrote the post (for caching)
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(personnel_id, post_id)         -- One entry per person-post combo
);

CREATE INDEX idx_feed_cache_personnel ON feed_cache(personnel_id, created_at DESC);
```

### Fanout Process (Post Created)
1. User posts → Insert into `posts` table
2. Get all eligible followers from `follows` table  
3. For each follower: Insert row into `feed_cache`
4. Async event system processes this

### Feed Query
```sql
SELECT f.post_id, p.caption, p.media_url, p.location, 
       p.visibility, a.first_name, a.last_name, a.avatar, a.user_id
FROM feed_cache f
JOIN posts p ON f.post_id = p.id
JOIN personnel a ON p.personnel_id = a.id
WHERE f.personnel_id = $1 AND p.deleted_at IS NULL
ORDER BY f.created_at DESC
LIMIT 50 OFFSET 0;
```

**Performance:** Single indexed query, ~5-20ms even with millions of entries.

### Cleanup on Delete
```sql
DELETE FROM feed_cache WHERE post_id = $1;
```

---

## User Counters Implementation

### Purpose
Avoid slow COUNT(*) queries for follower/following/post counts.

### Schema
```sql
CREATE TABLE user_counters (
    personnel_id UUID PRIMARY KEY,
    follower_count BIGINT DEFAULT 0,
    following_count BIGINT DEFAULT 0,
    post_count BIGINT DEFAULT 0,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);
```

### Update Strategies

**Option A: Triggers (Recommended for consistency)**
```sql
CREATE OR REPLACE FUNCTION update_follow_counts()
RETURNS TRIGGER AS $$
BEGIN
    IF TG_OP = 'INSERT' THEN
        INSERT INTO user_counters (personnel_id, follower_count)
        VALUES (NEW.following_id, 1)
        ON CONFLICT (personnel_id) 
        DO UPDATE SET follower_count = follower_count + 1, updated_at = NOW();
        
        INSERT INTO user_counters (personnel_id, following_count)
        VALUES (NEW.follower_id, 1)
        ON CONFLICT (personnel_id)
        DO UPDATE SET following_count = following_count + 1, updated_at = NOW();
    ELSIF TG_OP = 'DELETE' THEN
        UPDATE user_counters SET follower_count = follower_count - 1
        WHERE personnel_id = OLD.following_id AND follower_count > 0;
        
        UPDATE user_counters SET following_count = following_count - 1
        WHERE personnel_id = OLD.follower_id AND following_count > 0;
    END IF;
    RETURN COALESCE(NEW, OLD);
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER follows_counter_trigger 
AFTER INSERT OR DELETE ON follows
FOR EACH ROW EXECUTE FUNCTION update_follow_counts();
```

**Option B: Application-level atomic updates**
```sql
-- Increment (use in transaction)
UPDATE user_counters SET follower_count = follower_count + 1 WHERE personnel_id = $1;
INSERT INTO user_counters (personnel_id, follower_count) VALUES ($1, 1)
ON CONFLICT (personnel_id) DO UPDATE SET follower_count = follower_count + 1;

-- Decrement (use in transaction)
UPDATE user_counters SET follower_count = GREATEST(follower_count - 1, 0) WHERE personnel_id = $1;
```

---

## Full-Text Search Setup

### Install Extensions
```sql
CREATE EXTENSION IF NOT EXISTS pg_trgm;        -- Trigram similarity
CREATE EXTENSION IF NOT EXISTS pgcrypto;       -- Hash functions
```

### GIN Indexes for Search

**Personnel Name Search**
```sql
CREATE INDEX idx_personnel_name_search 
ON personnel USING gin(to_tsvector('english', first_name || ' ' || last_name))
WHERE deleted_at IS NULL;

CREATE INDEX idx_personnel_username_trgm 
ON personnel USING gin(mobile gin_trgm_ops)
WHERE deleted_at IS NULL;
```

**Hashtag Search**
```sql
CREATE INDEX idx_hashtags_search 
ON hashtags USING gin(name gin_trgm_ops)
WHERE deleted_at IS NULL;
```

**Post Caption Search**
```sql
CREATE INDEX idx_posts_caption_search 
ON posts USING gin(to_tsvector('english', caption))
WHERE deleted_at IS NULL;
```

---

## Privacy & Security Considerations

### Private Account Support
Add `is_private` flag to personnel table:
```sql
ALTER TABLE personnel ADD COLUMN is_private BOOLEAN DEFAULT FALSE;

-- Query modification for private account posts
SELECT p.* FROM posts p 
JOIN follows f ON p.personnel_id = f.following_id
WHERE f.follower_id = $1 AND p.deleted_at IS NULL
AND (p.visibility = 'PUBLIC' OR (p.visibility = 'PRIVATE' AND NOT p.is_private))
LIMIT 50;
```

### Blocked User Filtering
When generating feed, exclude posts from blocked users:
```sql
SELECT p.* FROM posts p
JOIN follows f ON p.personnel_id = f.following_id
LEFT JOIN blocks b ON p.personnel_id = b.blocked_id AND b.blocker_id = $1
WHERE f.follower_id = $1 AND p.deleted_at IS NULL 
  AND (b.id IS NULL OR b.blocker_id IS NULL)
LIMIT 50;
```

---

## Backup & Recovery Strategy

### Daily Full Backup (2 AM UTC)
```bash
pg_dump -h localhost -U postgres chat_app > backup_$(date +%Y%m%d).sql
pg_basebackup -h localhost -D /var/backups/base_backup -Ft -z -P -Xs
```

### Incremental Backups
- WAL archiving enabled
- Recovery Point Objective: 15 minutes
- Recovery Time Objective: 1 hour

### Table-Specific Retention

| Table | Retention | Archive Strategy |
|-------|-----------|------------------|
| notifications | 90 days auto-delete | Move to cold storage after 30 days |
| messages | 2 years in DB, archive older | Export to S3 annually |
| sessions | Delete expired immediately | Weekly cleanup job |

---

## Connection Pooling Configuration

```go
import (
    "github.com/jackc/pgx/v5/pgxpool"
)

// Recommended pool settings
config := &pgxpool.Config{
    MaxConns:         100,      // Total connections across all workers
    MinConns:         10,       // Minimum idle connections
    MaxConnLifetime: 1 * time.Hour,     // Close old connections
    MaxConnIdleTime: 30 * time.Minute,  // Close idle connections
    HealthCheckPeriod: 1 * time.Minute, // Check connection health
}

pool, err := pgxpool.NewWithConfig(ctx, config)
```

---

## Performance Targets

| Operation | Target Time | Acceptable |
|-----------|-------------|------------|
| Feed generation (50 posts) | < 50ms | < 100ms |
| User profile lookup | < 20ms | < 50ms |
| Post detail query | < 30ms | < 75ms |
| Search results (< 1K) | < 100ms | < 200ms |
| Notification list (100 items) | < 50ms | < 100ms |
| Follow/following query (1K users) | < 100ms | < 250ms |

---

## Conclusion

The enhanced schema supports:
- **Scalable feeds** via `feed_cache` for fanout-on-write architecture
- **Fast counters** via denormalized `user_counters` table
- **Complete notifications** system with preferences
- **Privacy controls** with private accounts and blocking
- **Full-text search** with PostgreSQL GIN indexes
- **Session management** for security and token revocation
- **Content moderation** with reporting system

All recommended indexes ensure optimal query performance as the dataset grows to millions of rows.
