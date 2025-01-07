CREATE TABLE outbox (
    id SERIAL PRIMARY KEY,
    payload JSONB NOT NULL,
    published BOOLEAN DEFAULT FALSE,
    event_type varchar(100) NOT NULL,
    created_at TIMESTAMP DEFAULT now()
);
