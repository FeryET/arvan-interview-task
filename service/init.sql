\c db;
CREATE TABLE IF NOT EXISTS ip_cache (
    id SERIAL PRIMARY KEY, -- Unique identifier for each entry
    ip VARCHAR(64),                    -- Stores IP addresses, max length of IPv6 is 45 characters
    country VARCHAR(64)                -- Stores country names, max length 64 to cover the longest names, the United Kingdom is 56 characters.
);