create table notes_authors (
  id uuid primary key default uuid_generate_v4(),
  note_id uuid references notes (id),
  user_id uuid references users (id)
);
