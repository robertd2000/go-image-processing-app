CREATE TABLE image_processing_jobs (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    image_id        UUID NOT NULL,
    status          VARCHAR(20) NOT NULL DEFAULT 'pending',
    event_id        UUID,
    error_message   TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT fk_job_image
        FOREIGN KEY(image_id)
        REFERENCES images(id)
        ON DELETE CASCADE
);

CREATE INDEX idx_jobs_image_id ON image_processing_jobs(image_id);
CREATE INDEX idx_jobs_status ON image_processing_jobs(status);
