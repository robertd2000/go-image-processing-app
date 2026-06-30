CREATE TABLE transformations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    image_id UUID NOT NULL,

    transform_spec JSONB NOT NULL,
    transform_hash VARCHAR(64) NOT NULL,

    status image_status NOT NULL DEFAULT 'pending',

    result_key TEXT,
    error_message TEXT,

    started_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,
    duration BIGINT,

    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT fk_transform_image
        FOREIGN KEY(image_id)
        REFERENCES images(id)
        ON DELETE CASCADE
);

CREATE INDEX idx_transform_image_id
ON transformations(image_id);

CREATE INDEX idx_transform_status
ON transformations(status);

CREATE INDEX idx_transform_spec
ON transformations USING GIN(transform_spec);

CREATE UNIQUE INDEX uniq_image_transform
ON transformations(image_id, transform_hash);