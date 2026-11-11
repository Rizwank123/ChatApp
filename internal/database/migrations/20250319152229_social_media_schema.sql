-- +goose Up
-- +goose StatementBegin
-- Create posts table
CREATE TABLE IF NOT EXISTS "public"."posts" (
    "id" UUID NOT NULL DEFAULT gen_random_uuid(),
    "personnel_id" UUID NOT NULL,
    "caption" TEXT,
    "media_url" TEXT [] NOT NULL,
    "location" VARCHAR(255),
    "visibility" VARCHAR(20) NOT NULL DEFAULT 'PUBLIC',
    "created_at" TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    "updated_at" TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    "deleted_at" TIMESTAMPTZ,
    PRIMARY KEY ("id"),
    FOREIGN KEY ("personnel_id") REFERENCES "personnel"("id")
);

-- Create follows table
CREATE TABLE IF NOT EXISTS "public"."follows" (
    "id" UUID NOT NULL DEFAULT gen_random_uuid(),
    "follower_id" UUID NOT NULL,
    "following_id" UUID NOT NULL,
    "created_at" TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    "updated_at" TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    "deleted_at" TIMESTAMPTZ,
    PRIMARY KEY ("id"),
    FOREIGN KEY ("follower_id") REFERENCES "personnel"("id"),
    FOREIGN KEY ("following_id") REFERENCES "personnel"("id"),
    UNIQUE("follower_id", "following_id")
);

-- Create likes table
CREATE TABLE IF NOT EXISTS "public"."likes" (
    "id" UUID NOT NULL DEFAULT gen_random_uuid(),
    "personnel_id" UUID NOT NULL,
    "post_id" UUID NOT NULL,
    "created_at" TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    "updated_at" TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    "deleted_at" TIMESTAMPTZ,
    PRIMARY KEY ("id"),
    FOREIGN KEY ("personnel_id") REFERENCES "personnel"("id"),
    FOREIGN KEY ("post_id") REFERENCES "posts"("id"),
    UNIQUE("personnel_id", "post_id")
);

-- Create comments table
CREATE TABLE IF NOT EXISTS "public"."comments" (
    "id" UUID NOT NULL DEFAULT gen_random_uuid(),
    "personnel_id" UUID NOT NULL,
    "post_id" UUID NOT NULL,
    "parent_id" UUID,
    "content" TEXT NOT NULL,
    "created_at" TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    "updated_at" TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    "deleted_at" TIMESTAMPTZ,
    PRIMARY KEY ("id"),
    FOREIGN KEY ("personnel_id") REFERENCES "personnel"("id"),
    FOREIGN KEY ("post_id") REFERENCES "posts"("id"),
    FOREIGN KEY ("parent_id") REFERENCES "comments"("id")
);

-- Create hashtags table
CREATE TABLE IF NOT EXISTS "public"."hashtags" (
    "id" UUID NOT NULL DEFAULT gen_random_uuid(),
    "name" VARCHAR(255) NOT NULL UNIQUE,
    "created_at" TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    "updated_at" TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    "deleted_at" TIMESTAMPTZ,
    PRIMARY KEY ("id")
);

-- Create post_hashtags table
CREATE TABLE IF NOT EXISTS "public"."post_hashtags" (
    "id" UUID NOT NULL DEFAULT gen_random_uuid(),
    "post_id" UUID NOT NULL,
    "hashtag_id" UUID NOT NULL,
    "created_at" TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    "updated_at" TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    "deleted_at" TIMESTAMPTZ,
    PRIMARY KEY ("id"),
    FOREIGN KEY ("post_id") REFERENCES "posts"("id"),
    FOREIGN KEY ("hashtag_id") REFERENCES "hashtags"("id"),
    UNIQUE("post_id", "hashtag_id")
);

-- +goose StatementEnd
-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS "public"."post_hashtags";

DROP TABLE IF EXISTS "public"."hashtags";

DROP TABLE IF EXISTS "public"."comments";

DROP TABLE IF EXISTS "public"."likes";

DROP TABLE IF EXISTS "public"."follows";

DROP TABLE IF EXISTS "public"."posts";

-- +goose StatementEnd