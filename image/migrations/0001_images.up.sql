-- ==========================
-- EXTENSIONS
-- ==========================
CREATE EXTENSION IF NOT EXISTS "pgcrypto";


-- ==========================
-- IMAGE STATUS TYPE
-- ==========================
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'image_status') THEN
        CREATE TYPE image_status AS ENUM (
            'pending',
            'processing',
            'done',
            'failed'
        );
    END IF;
END$$;


-- ==========================
-- IMAGES TABLE
-- ==========================
CREATE TABLE images (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID NOT NULL,

    original_name   VARCHAR(255) NOT NULL,
    storage_key     TEXT NOT NULL,

    file_size       BIGINT NOT NULL,
    mime_type       VARCHAR(100) NOT NULL,

    width           INT NOT NULL,
    height          INT NOT NULL,

    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_images_user_id ON images(user_id);

