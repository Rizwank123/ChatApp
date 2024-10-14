-- +goose Up
-- +goose StatementBegin
CREATE TYPE ACTIVATION_STATUS AS ENUM ('ACTIVE', 'DISABLED');

DROP TABLE IF EXISTS "public"."personnel";

CREATE TABLE "public"."personnel" (
    "id" UUID NOT NULL DEFAULT gen_random_uuid(),
    -- Assuming Base includes an ID
    "first_name" VARCHAR NOT NULL,
    "last_name" VARCHAR NOT NULL,
    "gender" VARCHAR,
    -- Assuming Gender is an ENUM or a VARCHAR
    "email" VARCHAR,
    -- Removed UNIQUE since it's part of a composite unique key
    "mobile" VARCHAR NOT NULL,
    "role" VARCHAR NOT NULL,
    "avatar" VARCHAR,
    "address" jsonb DEFAULT '{}' :: jsonb,
    -- Assuming UserRole is an ENUM or a VARCHAR
    "user_id" UUID NOT NULL UNIQUE,
    -- Assuming UserID is unique
    "activation_status" VARCHAR NOT NULL,
    -- Assuming ActivationStatus is an ENUM or a VARCHAR
    "created_at" TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    -- Assuming BaseAudit includes created_at
    "updated_at" TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    -- Assuming BaseAudit includes updated_at
    "deleted_at" TIMESTAMPTZ,
    -- Assuming BaseAudit includes deleted_at
    PRIMARY KEY ("id"),
    UNIQUE ("first_name", "last_name", "email"),
    -- Unique combination constraint
    FOREIGN KEY ("user_id") REFERENCES "public"."users"("id") -- Foreign key constraint
);

-- +goose StatementEnd
-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS "public"."personnel";

DROP TYPE IF EXISTS ACTIVATION_STATUS;

-- +goose StatementEnd