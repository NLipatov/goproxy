CREATE TABLE invoices (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL,
    ext_id VARCHAR(36) NOT NULL UNIQUE,
    status VARCHAR(50) NOT NULL DEFAULT 'new',
    email VARCHAR NOT NULL,
    offer_id VARCHAR(36) NOT NULL references offers,
    periodicity VARCHAR(50) NOT NULL,
    currency VARCHAR(3) NOT NULL,
    payment_method VARCHAR(60) NOT NULL,
    buyer_language VARCHAR(3) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_offer_id ON invoices (offer_id);
CREATE INDEX idx_user_id_status ON invoices (user_id, status);
CREATE INDEX idx_created_at ON invoices (created_at);
