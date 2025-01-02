CREATE TABLE outbox (
    id SERIAL PRIMARY KEY,
    payload JSONB NOT NULL,
    published BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT now()
);
