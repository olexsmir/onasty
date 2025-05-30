ALTER TABLE notes_authors
    ADD CONSTRAINT notes_authors_pair_user UNIQUE (note_id, user_id)
