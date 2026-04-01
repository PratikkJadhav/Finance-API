CREATE TYPE txn_type as ENUM ('income' , 'expense');

CREATE TABLE transactions (
    id UUID PRIMARY KEY DEFAULT gen_random_UUID(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    amount NUMERIC(12 , 2) NOT NULL CHECK (amount > 0),
    type txn_type NOT NULL,
    category TEXT NOT NULL,
    description TEXT,
    date DATE NOT NULL,
    deleted_at TIMESTAMPZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_transactions_user_id ON transactions(user_id);
CREATE INDEX idx_transactions_date ON transactions(date);
CREATE INDEX idx_transactions_type ON transactions(type);
CREATE INDEX idx_transactions_category ON transactions(category);