CREATE TABLE users (
    id UUID PRIMARY KEY,
    email TEXT NOT NULL,
    display_name TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL,
    CONSTRAINT users_email_not_empty CHECK (email <> ''),
    CONSTRAINT users_display_name_not_empty CHECK (display_name <> '')
);

CREATE UNIQUE INDEX users_email_key ON users (lower(email));
