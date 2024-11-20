CREATE TABLE sessions (
    id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id uuid REFERENCES users (id),
    refresh_token varchar(255) NOT NULL UNIQUE,
    expires_at timestamptz NOT NULL
);
