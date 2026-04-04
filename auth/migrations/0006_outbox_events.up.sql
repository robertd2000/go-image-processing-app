CREATE TABLE outbox_events (
    id UUID PRIMARY KEY,
    event_type TEXT NOT NULL,
    topic TEXT NOT NULL,
    key TEXT,
    payload JSONB NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    processed_at TIMESTAMP
);

CREATE INDEX idx_outbox_unprocessed
ON outbox_events (processed_at)
WHERE processed_at IS NULL;