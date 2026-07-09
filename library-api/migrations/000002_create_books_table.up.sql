CREATE TABLE books (
    id BIGSERIAL PRIMARY KEY,
    author_id BIGINT NOT NULL REFERENCES authors(id) ON DELETE RESTRICT,
    title TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    status TEXT NOT NULL DEFAULT 'to_read',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT books_status_check
    CHECK (status IN ('to_read', 'reading', 'finished'))
);
