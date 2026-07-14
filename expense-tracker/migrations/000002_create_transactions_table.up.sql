CREATE TABLE transactions (
    id BIGSERIAL PRIMARY KEY,
    category_id BIGINT NOT NULL REFERENCES categories(id) ON DELETE RESTRICT,
    type TEXT NOT NULL,
    amount BIGINT NOT NULL,
    note TEXT NOT NULL DEFAULT '',
    date DATE NOT NULL DEFAULT CURRENT_DATE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT transactions_type_check
    CHECK (type IN ('income', 'expense')),

    CONSTRAINT transactions_amount_positive
    CHECK (amount > 0)
);
