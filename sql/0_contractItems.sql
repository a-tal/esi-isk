CREATE TABLE IF NOT EXISTS contractItems (
    id          SERIAL  NOT NULL,
    contract_id INTEGER NOT NULL,
    type_id     INTEGER NOT NULL,
    item_id     BIGINT,
    quantity    BIGINT  NOT NULL,
    PRIMARY KEY (id)
);
