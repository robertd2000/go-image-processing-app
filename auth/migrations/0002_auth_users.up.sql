CREATE TABLE auth_users (
    id            UUID PRIMARY KEY,
    username      VARCHAR(20) NOT NULL UNIQUE,
    email         VARCHAR(64) UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    status        TEXT NOT NULL,

    created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_auth_users_email ON auth_users(email);