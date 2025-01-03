CREATE TABLE user_limits (
    id SERIAL PRIMARY KEY,
    user_id int NOT NULL,
    limit_bytes BIGINT DEFAULT NULL,
    valid_to TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT now()
);

create index on user_limits (user_id, valid_to)