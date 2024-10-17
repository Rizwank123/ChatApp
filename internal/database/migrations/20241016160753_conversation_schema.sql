-- +goose Up
-- +goose StatementBegin
-- DROP TABLE IF EXISTS "public"."chats";
-- DROP TABLE IF EXISTS "public"."chat_participants";
CREATE TYPE status AS ENUM ('Sent', 'Delivered', 'Read');

DROP TABLE IF EXISTS "public"."messages";

DROP TABLE IF EXISTS "public"."message_status";

CREATE TABLE "public"."messages" (
    "id" UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    -- UUID for the message
    "sender_id" UUID REFERENCES "public"."personnel"(id) ON DELETE CASCADE,
    -- Chat ID (foreign key)
    "receiver_id" UUID REFERENCES "public"."personnel"(id) ON DELETE CASCADE,
    -- Sender ID (foreign key)
    "content" TEXT NOT NULL,
    -- Content of the message
    "created_at" TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    "deleted_at" TIMESTAMPTZ
);

CREATE TABLE "public"."message_status" (
    "id" UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    "message_id" UUID REFERENCES "public"."messages"(id) ON DELETE CASCADE,
    "status" status DEFAULT "Sent",
    "updated_at" TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP FOREIGN KEY (message_id) REFERENCES "public"."messages"(id) ON DELETE CASCADE
) -- +goose StatementEnd
-- +goose Down
CREATE TYPE status AS ENUM ('Sent', 'Delivered', 'Read');

DROP TABLE IF EXISTS "public"."messages";

DROP TABLE IF EXISTS "public"."message_status";

-- +goose StatementBegin
-- +goose StatementEnd