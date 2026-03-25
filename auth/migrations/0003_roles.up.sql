CREATE TABLE roles (
    id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name        VARCHAR(20) NOT NULL UNIQUE,

    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    modified_at TIMESTAMPTZ,
    deleted_at  TIMESTAMPTZ,

    created_by      UUID,
    modified_by     UUID,
    deleted_by      UUID
);

INSERT INTO roles (name)
VALUES ('ADMIN'), ('USER');