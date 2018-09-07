CREATE TABLE IF NOT EXISTS contracts (
    contract_id INTEGER          NOT NULL,
    donator     INTEGER          NOT NULL,
    receiver    INTEGER          NOT NULL,
    location    BIGINT           NOT NULL,
    issued      TIMESTAMP        NOT NULL,
    expires     TIMESTAMP        NOT NULL,
    accepted    BOOLEAN          NOT NULL,
    value       DOUBLE PRECISION NOT NULL,
    note        TEXT             NOT NULL,
    PRIMARY KEY (contract_id)
);
