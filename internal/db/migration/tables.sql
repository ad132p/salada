CREATE TABLE posts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title VARCHAR(255) NOT NULL,
    slug VARCHAR(255) NOT NULL UNIQUE,
    content TEXT NOT NULL,
    author_id UUID NULL,
    author_name VARCHAR(255) NULL,
    published_at TIMESTAMP WITH TIME ZONE NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS sessions (
    id TEXT PRIMARY KEY,
    data BYTEA,
    created_on TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    expires_on TIMESTAMP WITH TIME ZONE
);

-- Optional: Add an index on expires_on for faster cleanup
CREATE INDEX IF NOT EXISTS idx_sessions_expires_on ON sessions (expires_on);

-- Also, your 'users' table (example structure):
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(255) UNIQUE NOT NULL,
    password TEXT NOT NULL, -- Storing hashed password
    role VARCHAR(50) NOT NULL DEFAULT 'anon',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
