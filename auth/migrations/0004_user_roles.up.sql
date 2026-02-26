CREATE TABLE user_roles (
    id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role_id     UUID NOT NULL REFERENCES roles(id) ON DELETE CASCADE,

    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    modified_at TIMESTAMPTZ,
    deleted_at  TIMESTAMPTZ,

    created_by  UUID,
    modified_by UUID,
    deleted_by  UUID,

    CONSTRAINT unique_user_role UNIQUE(user_id, role_id)
);
