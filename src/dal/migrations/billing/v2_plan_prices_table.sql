CREATE TABLE plan_prices(
    id SERIAL PRIMARY KEY,
    plan_id INT NOT NULL,
    currency VARCHAR(3) NOT NULL,
    cents BIGINT NOT NULL CHECK (cents >= 0),
    UNIQUE (plan_id, currency)
);

CREATE INDEX idx_plan_prices_plan_id on plan_prices(plan_id);
CREATE INDEX idx_plan_prices_plan_id_currency on plan_prices(plan_id, currency);