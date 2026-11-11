-- +goose Up
-- +goose StatementBegin
DROP TYPE IF EXISTS USER_ROLE;

CREATE TYPE USER_ROLE AS ENUM ('ADMIN', 'USER');

DROP TABLE IF EXISTS "public"."users";

CREATE TABLE "public"."users" (
    "id" UUID NOT NULL DEFAULT gen_random_uuid(),
    "user_name" VARCHAR NOT NULL UNIQUE,
    "password" VARCHAR,
    "role" USER_ROLE NOT NULL,
    "created_at" TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    "updated_at" TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    "deleted_at" TIMESTAMPTZ,
    PRIMARY KEY ("id")
);

-- +goose StatementEnd
-- +goose Down
-- +goose StatementBegin
DROP TYPE IF EXISTS USER_ROLE;

DROP TABLE IF EXISTS users;

-- +goose StatementEnd