CREATE TABLE documents (
    id uuid PRIMARY KEY DEFAULT UUID_GENERATE_V4(),
    name text NOT NULL,
    owner_id uuid DEFAULT UUID_GENERATE_V4(),
    content bytea,
    CONSTRAINT fk_user
          FOREIGN KEY(owner_id) 
            REFERENCES users(id)
);

CREATE INDEX IF NOT EXISTS "idx_documents_id" ON "documents" ("id");
