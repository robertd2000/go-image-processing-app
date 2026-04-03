CREATE TABLE users (
    id UUID PRIMARY KEY,

    username VARCHAR(20) UNIQUE,
    first_name TEXT,
    last_name TEXT,
    email VARCHAR(64) UNIQUE,

    avatar_url TEXT,

    status TEXT NOT NULL DEFAULT 'active',
    role TEXT NOT NULL DEFAULT 'user',

    last_seen_at TIMESTAMP,

    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP
);

CREATE INDEX idx_users_username ON users(username);
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_status ON users(status);