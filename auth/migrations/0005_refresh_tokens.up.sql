CREATE TABLE refresh_tokens (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id         UUID NOT NULL,

    token_hash      VARCHAR(255) NOT NULL,

    expires_at      TIMESTAMPTZ NOT NULL,
    revoked_at      TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),

    parent_id       UUID,
    family_id       UUID NOT NULL,

    CONSTRAINT fk_refresh_user
        FOREIGN KEY(user_id)
        REFERENCES auth_users(id)
        ON DELETE CASCADE
);

CREATE UNIQUE INDEX idx_refresh_token_hash ON refresh_tokens(token_hash);
CREATE INDEX idx_refresh_family_id ON refresh_tokens(family_id);
CREATE INDEX idx_refresh_user_id ON refresh_tokens(user_id);