CREATE EXTENSION IF NOT EXISTS "pgcrypto";

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_type WHERE typname = 'transformation_status'
    ) THEN
        CREATE TYPE transformation_status AS ENUM (
            'pending',
            'processing',
            'completed',
            'failed'
        );
    END IF;
END $$;

CREATE TABLE transformations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    -- Original image
    image_id UUID NOT NULL,

    source_storage_key TEXT NOT NULL,
    source_mime_type VARCHAR(100) NOT NULL,

    source_width INT NOT NULL,
    source_height INT NOT NULL,

    -- Transformation
    transform_spec JSONB NOT NULL,
    transform_hash CHAR(64) NOT NULL,

    status transformation_status NOT NULL DEFAULT 'pending',

    -- Result image
    result_storage_key TEXT,
    result_mime_type VARCHAR(100),
    result_width INT,
    result_height INT,
    result_size BIGINT,

    -- Failure information
    error_message TEXT,

    -- Processing timestamps
    started_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,

    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT uq_transformations_image_hash
        UNIQUE (image_id, transform_hash)
);

CREATE INDEX idx_transformations_image
    ON transformations(image_id);

CREATE INDEX idx_transformations_status
    ON transformations(status);

CREATE INDEX idx_transformations_pending
    ON transformations(status, created_at);

CREATE INDEX idx_transformations_spec
    ON transformations
    USING GIN(transform_spec);