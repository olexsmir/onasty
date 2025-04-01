ALTER TABLE notes
    ADD COLUMN "read" boolean DEFAULT FALSE NOT NULL,
    ADD COLUMN "read_at" timestamptz DEFAULT '1970-01-01 00:00:00+00'::timestamptz NOT NULL;
