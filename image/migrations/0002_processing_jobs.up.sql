-- ==========================
-- PROCESSING TYPE ENUM
-- ==========================
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'processing_type') THEN
        CREATE TYPE processing_type AS ENUM (
            'resize',
            'crop',
            'rotate',
            'filter',
            'watermark',
            'compress',
            'format'
        );
    END IF;
END$$;


-- ==========================
-- PROCESSING JOBS
-- ==========================
CREATE TABLE processing_jobs (
    id              BIGSERIAL PRIMARY KEY,

    image_id        BIGINT NOT NULL,

    processing_type processing_type NOT NULL,
    parameters      JSONB,

    status          image_status NOT NULL DEFAULT 'pending',

    result_path     TEXT,
    error_message   TEXT,

    started_at      TIMESTAMPTZ,
    completed_at    TIMESTAMPTZ,
    duration        BIGINT, -- milliseconds

    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    modified_at     TIMESTAMPTZ,
    deleted_at      TIMESTAMPTZ,

    created_by      BIGINT NOT NULL,
    modified_by     BIGINT,
    deleted_by      BIGINT,

    CONSTRAINT fk_processing_image
        FOREIGN KEY(image_id)
        REFERENCES images(id)
        ON DELETE CASCADE
);

CREATE INDEX idx_jobs_image_id ON processing_jobs(image_id);
CREATE INDEX idx_jobs_status ON processing_jobs(status);
CREATE INDEX idx_jobs_deleted_at ON processing_jobs(deleted_at);
CREATE INDEX idx_jobs_parameters ON processing_jobs USING GIN (parameters);
