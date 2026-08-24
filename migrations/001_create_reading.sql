CREATE TABLE IF NOT EXISTS reading (
    name        TEXT             NOT NULL,
    timestamp   TIMESTAMPTZ      NOT NULL,
    temperature DOUBLE PRECISION NOT NULL
);

CREATE INDEX IF NOT EXISTS reading_name_timestamp_idx
    ON reading (name, timestamp DESC);
