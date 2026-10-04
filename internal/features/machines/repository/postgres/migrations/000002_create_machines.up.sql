CREATE TABLE IF NOT EXISTS machines (
    id UUID PRIMARY KEY,
    label TEXT NOT NULL,
    coin_bank JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
