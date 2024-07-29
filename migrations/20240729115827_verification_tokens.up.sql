create table verification_tokens (
  id uuid primary key default uuid_generate_v4(),
  user_id uuid not null unique references users(id),
  token varchar(255) not null unique,
  created_at timestamptz not null default now(),
  expires_at timestamptz not null,
  used_at timestamptz default null
);
