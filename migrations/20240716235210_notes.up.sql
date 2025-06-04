CREATE TABLE notes (
    id uuid PRIMARY KEY DEFAULT uuid_generate_v4 (),
    content text NOT NULL,
    slug varchar(255) NOT NULL UNIQUE,
    burn_before_expiration boolean DEFAULT FALSE,
    created_at timestamptz NOT NULL DEFAULT now(),
    expires_at timestamptz
);
