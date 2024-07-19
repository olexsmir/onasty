alter table notes
add column user_id uuid references users (id);
