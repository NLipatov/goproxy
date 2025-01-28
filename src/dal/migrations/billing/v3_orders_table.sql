CREATE TYPE order_status AS ENUM ('NEW', 'PAID');

CREATE TABLE orders (
    id SERIAL PRIMARY KEY,
    email VARCHAR(255) NOT NULL,
    plan_id INT NOT NULL,
    status order_status DEFAULT 'NEW',
    created_at TIMESTAMP DEFAULT now()
);

CREATE INDEX idx_orders_email ON orders(email);
CREATE INDEX idx_orders_email_created_at ON orders(email, created_at);
CREATE INDEX idx_orders_plan_id ON orders(plan_id);
CREATE INDEX idx_orders_status ON orders(status);
