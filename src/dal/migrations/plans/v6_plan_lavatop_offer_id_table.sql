CREATE TABLE plan_lavatop_offer(
    id SERIAL PRIMARY KEY,
    plan_id INT NOT NULL REFERENCES plans(id) ON DELETE CASCADE,
    offer_id UUID NOT NULL    
);

CREATE INDEX idx_plan_lavatop_offer_plan_id on plan_lavatop_offer(plan_id);
CREATE INDEX idx_plan_lavatop_offer_offer_id on plan_lavatop_offer(offer_id);