CREATE TABLE plan_features (
    id SERIAL PRIMARY KEY,
    plan_id INT NOT NULL REFERENCES plans(id) ON DELETE CASCADE,
    feature_id INT NOT NULL REFERENCES features(id) ON DELETE CASCADE,
    UNIQUE(plan_id, feature_id)
);

CREATE INDEX idx_plan_features_plan_id ON plan_features (plan_id);
CREATE INDEX idx_plan_features_feature_id ON plan_features (feature_id);