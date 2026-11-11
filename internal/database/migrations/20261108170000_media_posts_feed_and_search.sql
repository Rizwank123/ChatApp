-- +goose Up
-- +goose StatementBegin

-- 1. Enable extensions for fuzzy/trigram search
CREATE EXTENSION IF NOT EXISTS pg_trgm;

-- 2. Add post status, archive flag, and search_vector to posts
ALTER TABLE "public"."posts" ADD COLUMN IF NOT EXISTS "status" VARCHAR(20) NOT NULL DEFAULT 'PUBLISHED';
ALTER TABLE "public"."posts" ADD COLUMN IF NOT EXISTS "is_archived" BOOLEAN NOT NULL DEFAULT FALSE;
ALTER TABLE "public"."posts" ADD COLUMN IF NOT EXISTS "search_vector" tsvector;

-- Backfill search_vector
UPDATE "public"."posts" SET "search_vector" = to_tsvector('english', coalesce("caption", '') || ' ' || coalesce("location", ''))
WHERE "search_vector" IS NULL;

-- Search vector trigger function
CREATE OR REPLACE FUNCTION posts_search_vector_update() RETURNS trigger AS $$
BEGIN
    NEW.search_vector := to_tsvector('english', coalesce(NEW.caption, '') || ' ' || coalesce(NEW.location, ''));
    RETURN NEW;
END
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trg_posts_search_vector ON "public"."posts";
CREATE TRIGGER trg_posts_search_vector
BEFORE INSERT OR UPDATE OF caption, location ON "public"."posts"
FOR EACH ROW EXECUTE FUNCTION posts_search_vector_update();

-- 3. Create media_uploads table for presigned upload tracking
CREATE TABLE IF NOT EXISTS "public"."media_uploads" (
    "id" UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    "owner_id" UUID NOT NULL REFERENCES "public"."personnel"("id") ON DELETE CASCADE,
    "bucket" VARCHAR(100) NOT NULL,
    "object_key" VARCHAR(500) NOT NULL,
    "original_filename" VARCHAR(255) NOT NULL,
    "mime_type" VARCHAR(100) NOT NULL,
    "file_size" BIGINT NOT NULL DEFAULT 0,
    "media_type" VARCHAR(20) NOT NULL,
    "status" VARCHAR(20) NOT NULL DEFAULT 'PENDING',
    "expires_at" TIMESTAMPTZ NOT NULL,
    "created_at" TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    "updated_at" TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_media_uploads_owner ON "public"."media_uploads"("owner_id", "status");
CREATE INDEX IF NOT EXISTS idx_media_uploads_expires ON "public"."media_uploads"("expires_at") WHERE "status" = 'PENDING';

-- 4. Create post_media table
CREATE TABLE IF NOT EXISTS "public"."post_media" (
    "id" UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    "post_id" UUID NOT NULL REFERENCES "public"."posts"("id") ON DELETE CASCADE,
    "owner_id" UUID NOT NULL REFERENCES "public"."personnel"("id") ON DELETE CASCADE,
    "bucket" VARCHAR(100) NOT NULL,
    "object_key" VARCHAR(500) NOT NULL,
    "original_filename" VARCHAR(255),
    "mime_type" VARCHAR(100) NOT NULL,
    "file_size" BIGINT NOT NULL DEFAULT 0,
    "media_type" VARCHAR(20) NOT NULL DEFAULT 'IMAGE',
    "width" INT DEFAULT 0,
    "height" INT DEFAULT 0,
    "duration" INT DEFAULT 0,
    "thumbnail_key" VARCHAR(500),
    "status" VARCHAR(20) NOT NULL DEFAULT 'READY',
    "display_order" INT DEFAULT 0,
    "created_at" TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    "updated_at" TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_post_media_post ON "public"."post_media"("post_id", "display_order" ASC);
CREATE INDEX IF NOT EXISTS idx_post_media_owner ON "public"."post_media"("owner_id");

-- 5. Create user_mentions table
CREATE TABLE IF NOT EXISTS "public"."user_mentions" (
    "id" UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    "post_id" UUID NOT NULL REFERENCES "public"."posts"("id") ON DELETE CASCADE,
    "mentioned_personnel_id" UUID NOT NULL REFERENCES "public"."personnel"("id") ON DELETE CASCADE,
    "created_at" TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    UNIQUE("post_id", "mentioned_personnel_id")
);
CREATE INDEX IF NOT EXISTS idx_user_mentions_mentioned ON "public"."user_mentions"("mentioned_personnel_id");

-- 6. Create media_processing_jobs table
CREATE TABLE IF NOT EXISTS "public"."media_processing_jobs" (
    "id" UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    "media_id" UUID REFERENCES "public"."post_media"("id") ON DELETE CASCADE,
    "upload_id" UUID REFERENCES "public"."media_uploads"("id") ON DELETE SET NULL,
    "job_type" VARCHAR(50) NOT NULL,
    "status" VARCHAR(20) NOT NULL DEFAULT 'PENDING',
    "retry_count" INT DEFAULT 0,
    "max_retries" INT DEFAULT 3,
    "error_message" TEXT,
    "created_at" TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    "updated_at" TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_media_jobs_status ON "public"."media_processing_jobs"("status", "created_at" ASC);

-- 7. High performance indexes for search and feeds
CREATE INDEX IF NOT EXISTS idx_posts_search_vector ON "public"."posts" USING gin("search_vector");
CREATE INDEX IF NOT EXISTS idx_posts_caption_trgm ON "public"."posts" USING gin("caption" gin_trgm_ops);
CREATE INDEX IF NOT EXISTS idx_hashtags_name_trgm ON "public"."hashtags" USING gin("name" gin_trgm_ops);
CREATE INDEX IF NOT EXISTS idx_users_username_trgm ON "public"."users" USING gin("user_name" gin_trgm_ops);
CREATE INDEX IF NOT EXISTS idx_personnel_display_name_trgm ON "public"."personnel" USING gin("display_name" gin_trgm_ops);
CREATE INDEX IF NOT EXISTS idx_posts_public_feed ON "public"."posts"("created_at" DESC) WHERE "deleted_at" IS NULL AND "visibility" = 'PUBLIC' AND "status" = 'PUBLISHED' AND "is_archived" = FALSE;

-- +goose StatementEnd
-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_posts_public_feed;
DROP INDEX IF EXISTS idx_personnel_display_name_trgm;
DROP INDEX IF EXISTS idx_users_username_trgm;
DROP INDEX IF EXISTS idx_hashtags_name_trgm;
DROP INDEX IF EXISTS idx_posts_caption_trgm;
DROP INDEX IF EXISTS idx_posts_search_vector;
DROP TABLE IF EXISTS "public"."media_processing_jobs";
DROP TABLE IF EXISTS "public"."user_mentions";
DROP TABLE IF EXISTS "public"."post_media";
DROP TABLE IF EXISTS "public"."media_uploads";
DROP TRIGGER IF EXISTS trg_posts_search_vector ON "public"."posts";
DROP FUNCTION IF EXISTS posts_search_vector_update;
ALTER TABLE "public"."posts" DROP COLUMN IF EXISTS "search_vector";
ALTER TABLE "public"."posts" DROP COLUMN IF EXISTS "is_archived";
ALTER TABLE "public"."posts" DROP COLUMN IF EXISTS "status";
-- +goose StatementEnd
