CREATE TABLE users (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    username        VARCHAR(20) NOT NULL UNIQUE,
    first_name      VARCHAR(15),
    last_name       VARCHAR(25),
    mobile_number   VARCHAR(11) UNIQUE,
    email           VARCHAR(64) UNIQUE,
    password_hash   VARCHAR(255) NOT NULL,
    enabled         BOOLEAN NOT NULL DEFAULT TRUE,

    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    modified_at     TIMESTAMPTZ,
    deleted_at      TIMESTAMPTZ,

    created_by      BIGINT,
    modified_by     BIGINT,
    deleted_by      BIGINT
);

CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_deleted_at ON users(deleted_at);
