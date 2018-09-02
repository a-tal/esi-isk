CREATE TABLE IF NOT EXISTS characters (
    character_id     INTEGER          NOT NULL,
    character_name   VARCHAR(40)      NOT NULL,
    corporation_id   INTEGER          NOT NULL,
    corporation_name VARCHAR(40)      NOT NULL,
    received         BIGINT           NOT NULL,
    received_isk     DOUBLE PRECISION NOT NULL,
    donated          BIGINT           NOT NULL,
    donated_isk      DOUBLE PRECISION NOT NULL,
    joined           TIMESTAMP        NOT NULL,
    last_seen        TIMESTAMP        NOT NULL,
    last_donated     TIMESTAMP,
    last_received    TIMESTAMP,

    PRIMARY KEY (character_id)
);
