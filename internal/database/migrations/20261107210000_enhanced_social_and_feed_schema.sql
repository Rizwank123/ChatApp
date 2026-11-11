-- +goose Up
-- +goose StatementBegin

-- 1. Create missing indexes on existing social media tables
CREATE INDEX IF NOT EXISTS idx_posts_personnel_created ON "public"."posts"("personnel_id", "created_at" DESC) WHERE "deleted_at" IS NULL;
CREATE INDEX IF NOT EXISTS idx_posts_created ON "public"."posts"("created_at" DESC) WHERE "deleted_at" IS NULL;
CREATE INDEX IF NOT EXISTS idx_follows_follower ON "public"."follows"("follower_id") WHERE "deleted_at" IS NULL;
CREATE INDEX IF NOT EXISTS idx_follows_following ON "public"."follows"("following_id") WHERE "deleted_at" IS NULL;
CREATE INDEX IF NOT EXISTS idx_likes_post ON "public"."likes"("post_id") WHERE "deleted_at" IS NULL;
CREATE INDEX IF NOT EXISTS idx_likes_personnel ON "public"."likes"("personnel_id") WHERE "deleted_at" IS NULL;
CREATE INDEX IF NOT EXISTS idx_comments_post_created ON "public"."comments"("post_id", "created_at" DESC) WHERE "deleted_at" IS NULL;
CREATE INDEX IF NOT EXISTS idx_comments_parent ON "public"."comments"("parent_id") WHERE "deleted_at" IS NULL AND "parent_id" IS NOT NULL;

-- 2. Add new columns to existing tables
ALTER TABLE "public"."posts" ADD COLUMN IF NOT EXISTS "view_count" BIGINT DEFAULT 0;
ALTER TABLE "public"."posts" ADD COLUMN IF NOT EXISTS "share_count" BIGINT DEFAULT 0;
ALTER TABLE "public"."posts" ADD COLUMN IF NOT EXISTS "media_metadata" JSONB DEFAULT '{}'::jsonb;

ALTER TABLE "public"."personnel" ADD COLUMN IF NOT EXISTS "bio" TEXT DEFAULT '';
ALTER TABLE "public"."personnel" ADD COLUMN IF NOT EXISTS "is_private" BOOLEAN DEFAULT FALSE;
ALTER TABLE "public"."personnel" ADD COLUMN IF NOT EXISTS "website" VARCHAR(255) DEFAULT '';

-- 3. Create user_counters for O(1) profile counts
CREATE TABLE IF NOT EXISTS "public"."user_counters" (
    "personnel_id" UUID PRIMARY KEY REFERENCES "personnel"("id") ON DELETE CASCADE,
    "follower_count" BIGINT DEFAULT 0,
    "following_count" BIGINT DEFAULT 0,
    "post_count" BIGINT DEFAULT 0,
    "updated_at" TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

-- 4. Create blocks table
CREATE TABLE IF NOT EXISTS "public"."blocks" (
    "id" UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    "blocker_id" UUID NOT NULL REFERENCES "personnel"("id") ON DELETE CASCADE,
    "blocked_id" UUID NOT NULL REFERENCES "personnel"("id") ON DELETE CASCADE,
    "reason" TEXT,
    "created_at" TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    UNIQUE("blocker_id", "blocked_id")
);
CREATE INDEX IF NOT EXISTS idx_blocks_blocker ON "public"."blocks"("blocker_id");
CREATE INDEX IF NOT EXISTS idx_blocks_blocked ON "public"."blocks"("blocked_id");

-- 5. Create follow_requests table for private accounts
DO $$ BEGIN
    CREATE TYPE follow_request_status AS ENUM ('PENDING', 'ACCEPTED', 'REJECTED');
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

CREATE TABLE IF NOT EXISTS "public"."follow_requests" (
    "id" UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    "requester_id" UUID NOT NULL REFERENCES "personnel"("id") ON DELETE CASCADE,
    "requested_id" UUID NOT NULL REFERENCES "personnel"("id") ON DELETE CASCADE,
    "status" follow_request_status DEFAULT 'PENDING',
    "created_at" TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    "updated_at" TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    UNIQUE("requester_id", "requested_id")
);
CREATE INDEX IF NOT EXISTS idx_follow_requests_requested ON "public"."follow_requests"("requested_id", "status") WHERE "status" = 'PENDING';

-- 6. Create saved_posts (bookmarks) table
CREATE TABLE IF NOT EXISTS "public"."saved_posts" (
    "id" UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    "personnel_id" UUID NOT NULL REFERENCES "personnel"("id") ON DELETE CASCADE,
    "post_id" UUID NOT NULL REFERENCES "posts"("id") ON DELETE CASCADE,
    "created_at" TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    UNIQUE("personnel_id", "post_id")
);
CREATE INDEX IF NOT EXISTS idx_saved_posts_personnel ON "public"."saved_posts"("personnel_id", "created_at" DESC);

-- 7. Create notifications table
DO $$ BEGIN
    CREATE TYPE notification_type AS ENUM (
        'FOLLOW', 'FOLLOW_REQUEST', 'FOLLOW_ACCEPTED',
        'LIKE', 'COMMENT', 'MENTION', 'SYSTEM'
    );
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

CREATE TABLE IF NOT EXISTS "public"."notifications" (
    "id" UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    "recipient_id" UUID NOT NULL REFERENCES "personnel"("id") ON DELETE CASCADE,
    "actor_id" UUID REFERENCES "personnel"("id") ON DELETE SET NULL,
    "notification_type" notification_type NOT NULL,
    "entity_type" VARCHAR(50),
    "entity_id" UUID,
    "title" VARCHAR(255),
    "message" TEXT,
    "is_read" BOOLEAN DEFAULT FALSE,
    "read_at" TIMESTAMPTZ,
    "created_at" TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    "deleted_at" TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS idx_notifications_recipient ON "public"."notifications"("recipient_id", "created_at" DESC) WHERE "deleted_at" IS NULL;
CREATE INDEX IF NOT EXISTS idx_notifications_unread ON "public"."notifications"("recipient_id") WHERE "deleted_at" IS NULL AND "is_read" = FALSE;

-- 8. Create feed_cache table for fanout-on-write persistence fallback
CREATE TABLE IF NOT EXISTS "public"."feed_cache" (
    "id" UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    "personnel_id" UUID NOT NULL REFERENCES "personnel"("id") ON DELETE CASCADE,
    "post_id" UUID NOT NULL REFERENCES "posts"("id") ON DELETE CASCADE,
    "author_id" UUID NOT NULL REFERENCES "personnel"("id") ON DELETE CASCADE,
    "created_at" TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    UNIQUE("personnel_id", "post_id")
);
CREATE INDEX IF NOT EXISTS idx_feed_cache_user ON "public"."feed_cache"("personnel_id", "created_at" DESC);

-- 9. Create user_sessions table
CREATE TABLE IF NOT EXISTS "public"."user_sessions" (
    "id" UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    "user_id" UUID NOT NULL REFERENCES "users"("id") ON DELETE CASCADE,
    "token_jti" VARCHAR(255) UNIQUE NOT NULL,
    "refresh_token_hash" VARCHAR(255) NOT NULL,
    "ip_address" INET,
    "user_agent" TEXT,
    "expires_at" TIMESTAMPTZ NOT NULL,
    "revoked_at" TIMESTAMPTZ,
    "created_at" TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    "updated_at" TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_user_sessions_active ON "public"."user_sessions"("user_id", "expires_at") WHERE "revoked_at" IS NULL;

-- 10. Create outbox_events table for transactional outbox pattern
CREATE TABLE IF NOT EXISTS "public"."outbox_events" (
    "id" UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    "aggregate_type" VARCHAR(50) NOT NULL,
    "aggregate_id" UUID NOT NULL,
    "event_type" VARCHAR(100) NOT NULL,
    "payload" JSONB NOT NULL,
    "status" VARCHAR(20) NOT NULL DEFAULT 'PENDING',
    "retry_count" INT NOT NULL DEFAULT 0,
    "last_error" TEXT,
    "created_at" TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    "published_at" TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS idx_outbox_pending ON "public"."outbox_events"("created_at") WHERE "status" = 'PENDING';

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS "public"."outbox_events";
DROP TABLE IF EXISTS "public"."user_sessions";
DROP TABLE IF EXISTS "public"."feed_cache";
DROP TABLE IF EXISTS "public"."notifications";
DROP TYPE IF EXISTS notification_type;
DROP TABLE IF EXISTS "public"."saved_posts";
DROP TABLE IF EXISTS "public"."follow_requests";
DROP TYPE IF EXISTS follow_request_status;
DROP TABLE IF EXISTS "public"."blocks";
DROP TABLE IF EXISTS "public"."user_counters";

ALTER TABLE "public"."personnel" DROP COLUMN IF EXISTS "website";
ALTER TABLE "public"."personnel" DROP COLUMN IF EXISTS "is_private";
ALTER TABLE "public"."personnel" DROP COLUMN IF EXISTS "bio";

ALTER TABLE "public"."posts" DROP COLUMN IF EXISTS "media_metadata";
ALTER TABLE "public"."posts" DROP COLUMN IF EXISTS "share_count";
ALTER TABLE "public"."posts" DROP COLUMN IF EXISTS "view_count";

DROP INDEX IF EXISTS idx_comments_parent;
DROP INDEX IF EXISTS idx_comments_post_created;
DROP INDEX IF EXISTS idx_likes_personnel;
DROP INDEX IF EXISTS idx_likes_post;
DROP INDEX IF EXISTS idx_follows_following;
DROP INDEX IF EXISTS idx_follows_follower;
DROP INDEX IF EXISTS idx_posts_created;
DROP INDEX IF EXISTS idx_posts_personnel_created;
-- +goose StatementEnd
