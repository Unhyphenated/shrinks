-- Create the links table
CREATE TABLE IF NOT EXISTS links (
    id SERIAL PRIMARY KEY,
    long_url TEXT NOT NULL,
    short_url VARCHAR(10) NOT NULL UNIQUE,
    clicks INTEGER DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Optional: Add an index on short_url for lightning-fast redirects
CREATE INDEX IF NOT EXISTS idx_short_url ON links(short_url);