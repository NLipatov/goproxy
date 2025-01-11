CREATE TABLE invoices (
    id SERIAL PRIMARY KEY,
    ext_id UUID NOT NULL UNIQUE,
    user_id INT NOT NULL,
    offer_id UUID NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'new' CHECK (
      status IN (
                 'new',
                 'in-progress',
                 'completed',
                 'failed',
                 'cancelled',
                 'subscription-active',
                 'subscription-expired',
                 'subscription-cancelled',
                 'subscription-failed'
          )
      ),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_offer_id ON invoices (offer_id);
CREATE INDEX idx_user_id_status ON invoices (user_id, status);
CREATE INDEX idx_created_at ON invoices (created_at);
