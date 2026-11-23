CREATE TABLE IF NOT EXISTS sales_purchases (
    id UUID PRIMARY KEY,
    machine_id UUID NOT NULL,
    product_id UUID NOT NULL,
    price_cents INTEGER NOT NULL,
    total_inserted_cents INTEGER NOT NULL,
    change_cents INTEGER NOT NULL,
    change_coins JSONB NOT NULL,
    purchased_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_sales_purchases_machine_id ON sales_purchases (machine_id);
