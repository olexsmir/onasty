CREATE TYPE provider_enum AS ENUM (
    'google',
    'github'
);

CREATE TABLE oauth_identities (
    id uuid PRIMARY KEY DEFAULT uuid_generate_v4 (),
    user_id uuid REFERENCES users (id) ON DELETE CASCADE,
    provider provider_enum NOT NULL,
    provider_id varchar(50),
    created_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE (provider, provider_id)
);
