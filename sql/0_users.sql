CREATE TABLE IF NOT EXISTS users (
    refresh_token  TEXT      NOT NULL,
    access_token   TEXT      NOT NULL,
    access_expires TIMESTAMP NOT NULL,
    character_id   INTEGER,
    owner_hash     TEXT,
    last_processed TIMESTAMP,

    PRIMARY KEY (refresh_token)
);
