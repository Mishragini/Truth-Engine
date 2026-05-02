CREATE EXTENSION IF NOT EXISTS vector ;

CREATE TABLE IF NOT EXISTS search_responses (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    query TEXT NOT NULL UNIQUE,
    -- 1536 is for openai change it if you change the model to gen embedding
    query_embedding VECTOR(768),
    response TEXT,
    created_at TIMESTAMPTZ DEFAULT now(),
    -- 7 days of the initial query 
    valid_till TIMESTAMPTZ DEFAULT (now() + interval '7 days')
);

CREATE INDEX ON search_responses(valid_till);

CREATE INDEX ON search_responses 
USING hnsw (query_embedding vector_cosine_ops);

ANALYZE search_responses;