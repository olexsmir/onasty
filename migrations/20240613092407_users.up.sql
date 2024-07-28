create table users (
  id uuid primary key default uuid_generate_v4(),
  username varchar(255) not null unique,
  email varchar(255) not null unique,
  password varchar(255) not null,
  activated boolean not null default false,
  created_at timestamptz not null default now(),
  last_login_at timestamptz not null default now()
);
