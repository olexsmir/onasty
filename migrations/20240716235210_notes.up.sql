create table notes (
  id uuid primary key default uuid_generate_v4(),
  content text not null,
  slug varchar(255) not null unique,
  burn_before_expiration boolean default false,
  created_at timestamptz not null default now(),
  expires_at timestamptz
);
