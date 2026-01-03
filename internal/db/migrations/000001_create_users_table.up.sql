CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY,

    created_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ,
    deleted_at TIMESTAMPTZ,

    username VARCHAR(255) NOT NULL UNIQUE,
    password TEXT NOT NULL,

    is_passenger BOOLEAN DEFAULT false,
    is_driver BOOLEAN DEFAULT false,
    is_admin BOOLEAN DEFAULT false
    );

-- Add indexes for performance
CREATE INDEX IF NOT EXISTS idx_users_deleted_at ON users(deleted_at);
CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);