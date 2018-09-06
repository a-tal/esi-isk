CREATE TABLE IF NOT EXISTS users (
    refresh_token    TEXT      NOT NULL,
    access_token     TEXT      NOT NULL,
    access_expires   TIMESTAMP NOT NULL,
    character_id     INTEGER   NOT NULL,
    owner_hash       TEXT      NOT NULL,
    last_processed   TIMESTAMP,
    last_journal_id  BIGINT,
    last_contract_id BIGINT,

    PRIMARY KEY (refresh_token)
);
