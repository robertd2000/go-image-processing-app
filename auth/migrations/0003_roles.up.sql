CREATE TABLE roles (
    id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name        VARCHAR(20) NOT NULL UNIQUE,

    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    modified_at TIMESTAMPTZ,
    deleted_at  TIMESTAMPTZ,

    created_by  BIGINT,
    modified_by BIGINT,
    deleted_by  BIGINT
);

INSERT INTO roles (name, created_by)
VALUES ('ADMIN', 1), ('USER', 1);
