CREATE TABLE public.schema_migrations (
    id SERIAL PRIMARY KEY,
    version INT NOT NULL UNIQUE,
    applied_at TIMESTAMP DEFAULT now()
);
