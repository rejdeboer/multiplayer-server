CREATE TABLE documents (
    id uuid PRIMARY KEY DEFAULT UUID_GENERATE_V4(),
    name text NOT NULL,
    owner_id uuid NOT NULL REFERENCES users(id),
    shared_with uuid[] DEFAULT ARRAY[]::uuid[],
    state_vector bytea,
);

CREATE INDEX IF NOT EXISTS "idx_documents_id" ON "documents" ("id");
