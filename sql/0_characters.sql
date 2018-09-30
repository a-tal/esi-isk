CREATE TABLE IF NOT EXISTS characters (
    character_id     INTEGER          NOT NULL,
    corporation_id   INTEGER          NOT NULL,
    alliance_id      INTEGER          NOT NULL,
    received         BIGINT           NOT NULL,
    received_isk     DOUBLE PRECISION NOT NULL,
    received_30      BIGINT           NOT NULL,
    received_isk_30  DOUBLE PRECISION NOT NULL,
    donated          BIGINT           NOT NULL,
    donated_isk      DOUBLE PRECISION NOT NULL,
    donated_30       BIGINT           NOT NULL,
    donated_isk_30   DOUBLE PRECISION NOT NULL,
    last_donated     TIMESTAMP,
    last_received    TIMESTAMP,
    good_standing    BOOLEAN          NOT NULL DEFAULT false,

    PRIMARY KEY (character_id)
);
