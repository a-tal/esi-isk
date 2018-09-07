CREATE TABLE IF NOT EXISTS contractItems (
    id          BIGINT  NOT NULL,  -- record_id
    contract_id INTEGER NOT NULL,
    type_id     INTEGER NOT NULL,
    item_id     BIGINT  NOT NULL,
    quantity    INTEGER NOT NULL,
    PRIMARY KEY (id)
);
