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
END$$;

CREATE TABLE transformations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    image_id UUID NOT NULL,

    storage_key TEXT NOT NULL,

    mime_type VARCHAR(100) NOT NULL,

    width INT NOT NULL,
    height INT NOT NULL,

    transform_spec JSONB NOT NULL,
    transform_hash VARCHAR(64) NOT NULL,

    status transformation_status NOT NULL DEFAULT 'pending',

    result_key TEXT,

    error_message TEXT,

    started_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,

    duration_ms BIGINT,

    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_transformations_image_id
ON transformations(image_id);

CREATE INDEX idx_transformations_status
ON transformations(status);

CREATE INDEX idx_transformations_spec
ON transformations USING GIN(transform_spec);

CREATE UNIQUE INDEX uniq_transformations_image_hash
ON transformations(image_id, transform_hash);