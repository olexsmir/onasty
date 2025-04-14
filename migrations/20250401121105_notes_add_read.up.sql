ALTER TABLE notes
    ADD COLUMN "read_at" timestamptz NOT NULL DEFAULT '0001-01-01 00:00:00';
