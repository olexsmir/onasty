create table users (
  id uuid primary key default uuid_generate_v4(),
  username varchar(255) not null,
  email varchar(255) not null unique,
  password varchar(255) not null,
  created_at timestamptz not null default now(),
  last_login_at timestamptz not null default now()
);
