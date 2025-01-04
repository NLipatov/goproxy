CREATE TABLE plans (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    limit_bytes BIGINT DEFAULT NULL,
    duration_days INT NOT NULL,
    created_at TIMESTAMP DEFAULT now() NOT NULL
);

CREATE INDEX idx_plan_name ON plans (name);

INSERT INTO plans (name, limit_bytes, duration_days, created_at) VALUES ('Free', 200000000, 1, now())