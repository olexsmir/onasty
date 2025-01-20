ALTER TABLE notes
    ADD COLUMN "is_read" boolean DEFAULT FALSE NOT NULL,
    ADD COLUMN "read_at" timestamp;
