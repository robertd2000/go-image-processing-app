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
    id              BIGSERIAL PRIMARY KEY,
    user_id         BIGINT NOT NULL,

    original_name   VARCHAR(255) NOT NULL,
    file_name       VARCHAR(255) NOT NULL UNIQUE,
    file_path       TEXT NOT NULL,
    file_size       BIGINT NOT NULL,
    mime_type       VARCHAR(100) NOT NULL,

    width           INT NOT NULL,
    height          INT NOT NULL,

    status          image_status NOT NULL DEFAULT 'pending',

    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    modified_at     TIMESTAMPTZ,
    deleted_at      TIMESTAMPTZ,

    created_by      BIGINT NOT NULL,
    modified_by     BIGINT,
    deleted_by      BIGINT
);

CREATE INDEX idx_images_user_id ON images(user_id);
CREATE INDEX idx_images_status ON images(status);
CREATE INDEX idx_images_deleted_at ON images(deleted_at);
