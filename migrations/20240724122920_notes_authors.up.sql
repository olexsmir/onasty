CREATE TABLE notes_authors (
    id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
    note_id uuid REFERENCES notes (id) ON DELETE CASCADE,
    user_id uuid REFERENCES users (id) ON DELETE CASCADE,
    created_at timestamptz NOT NULL DEFAULT now()
);
