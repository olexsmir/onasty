create table sessions (
  id uuid primary key default uuid_generate_v4(),
  user_id uuid references users (id),
  refresh_token varchar(255) not null unique,
  expires_at timestamptz not null
);
