CREATE TABLE user_plans (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL,
    plan_id INT NOT NULL DEFAULT 1 REFERENCES plans(id) ON DELETE CASCADE,
    valid_to TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT now() NOT NULL
);

CREATE INDEX active_user_plans_idx ON user_plans (user_id, valid_to DESC);