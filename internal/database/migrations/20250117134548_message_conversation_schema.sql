-- +goose Up
-- +goose StatementBegin
-- Drop dependent tables first in the correct order
DROP TABLE IF EXISTS "public"."message_stat";
DROP TABLE IF EXISTS "public"."messages";
DROP TABLE IF EXISTS "public"."conversation_participants";
DROP TABLE IF EXISTS "public"."conversations";
DROP TABLE IF EXISTS "public"."friendships";

-- Create friendships table
CREATE TABLE IF NOT EXISTS "public"."friendships" (
    "id" UUID NOT NULL DEFAULT gen_random_uuid(),
    "personnel_id" UUID NOT NULL,
    "friend_id" UUID NOT NULL,
    "status" FRIENDSHIP_STATUS NOT NULL DEFAULT 'PENDING',
    "created_at" TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    "updated_at" TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    "deleted_at" TIMESTAMPTZ,
    PRIMARY KEY ("id"),
    FOREIGN KEY ("personnel_id") REFERENCES "personnel"("id"),
    FOREIGN KEY ("friend_id") REFERENCES "personnel"("id"),
    UNIQUE("personnel_id", "friend_id")
);

-- Create conversations table
CREATE TABLE IF NOT EXISTS "public"."conversations" (
    "id" UUID NOT NULL DEFAULT gen_random_uuid(),
    "name" VARCHAR,
    "conversation_type" CONVERSATION_TYPE NOT NULL,
    "created_at" TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    "updated_at" TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    "deleted_at" TIMESTAMPTZ,
    PRIMARY KEY ("id")
);

-- Create conversation participants table
CREATE TABLE IF NOT EXISTS "public"."conversation_participants" (
    "id" UUID NOT NULL DEFAULT gen_random_uuid(),
    "conversation_id" UUID NOT NULL,
    "personnel_id" UUID NOT NULL,
    "joined_at" TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    "left_at" TIMESTAMPTZ,
    "created_at" TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    "updated_at" TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    "deleted_at" TIMESTAMPTZ,
    PRIMARY KEY ("id"),
    FOREIGN KEY ("conversation_id") REFERENCES "conversations"("id"),
    FOREIGN KEY ("personnel_id") REFERENCES "personnel"("id"),
    UNIQUE("conversation_id", "personnel_id")
);

-- Create messages table
CREATE TABLE IF NOT EXISTS "public"."messages" (
    "id" UUID NOT NULL DEFAULT gen_random_uuid(),
    "conversation_id" UUID NOT NULL,
    "sender_id" UUID NOT NULL,
    "message_type" MESSAGE_TYPE NOT NULL DEFAULT 'TEXT',
    "content" TEXT NOT NULL,
    "created_at" TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    "updated_at" TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    "deleted_at" TIMESTAMPTZ,
    PRIMARY KEY ("id"),
    FOREIGN KEY ("conversation_id") REFERENCES "conversations"("id"),
    FOREIGN KEY ("sender_id") REFERENCES "personnel"("id")
);

-- Create message status table
CREATE TABLE IF NOT EXISTS "public"."message_stat" (
    "id" UUID NOT NULL DEFAULT gen_random_uuid(),
    "message_id" UUID NOT NULL,
    "personnel_id" UUID NOT NULL,
    "status" MESSAGE_STATUS NOT NULL DEFAULT 'DELIVERED',
    "created_at" TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    "updated_at" TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    "deleted_at" TIMESTAMPTZ,
    PRIMARY KEY ("id"),
    FOREIGN KEY ("message_id") REFERENCES "messages"("id"),
    FOREIGN KEY ("personnel_id") REFERENCES "personnel"("id"),
    UNIQUE("message_id", "personnel_id")
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS "public"."message_stat";
DROP TABLE IF EXISTS "public"."messages";
DROP TABLE IF EXISTS "public"."conversation_participants";
DROP TABLE IF EXISTS "public"."conversations";
DROP TABLE IF EXISTS "public"."friendships";
-- +goose StatementEnd
