CREATE TABLE IF NOT EXISTS donations (
    transaction_id BIGINT           NOT NULL,
    donator        INTEGER          NOT NULL,
    receiver       INTEGER          NOT NULL,
    "timestamp"    TIMESTAMP        NOT NULL,
    note           VARCHAR(40)      NOT NULL,
    amount         DOUBLE PRECISION NOT NULL,
    PRIMARY KEY (transaction_id)
);
