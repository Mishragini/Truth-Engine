CREATE TABLE IF NOT EXISTS resources(
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title TEXT,
    search_response_id UUID REFERENCES search_responses(id) ON DELETE CASCADE
);


CREATE TABLE IF NOT EXISTS sources(
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    resource_id UUID REFERENCES resources(id) ON DELETE CASCADE,
    content TEXT[]
);