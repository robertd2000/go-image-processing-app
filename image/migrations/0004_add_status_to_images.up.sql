ALTER TABLE images
    ADD COLUMN status VARCHAR(20) NOT NULL DEFAULT 'pending';

CREATE INDEX idx_images_status ON images(status);

UPDATE images SET status = 'completed';
