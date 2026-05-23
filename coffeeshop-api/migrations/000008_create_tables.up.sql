CREATE TABLE IF NOT EXISTS tables (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    number      TEXT NOT NULL UNIQUE,
    token       TEXT NOT NULL UNIQUE,
    is_active   BOOLEAN DEFAULT true,
    created_at  TIMESTAMPTZ DEFAULT now()
);
