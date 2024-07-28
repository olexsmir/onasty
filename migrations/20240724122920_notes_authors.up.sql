create table notes_authors (
  id uuid primary key default uuid_generate_v4(),
  note_id uuid references notes (id) on delete cascade,
  user_id uuid references users (id) on delete cascade,
  created_at timestamptz not null default now()
);
