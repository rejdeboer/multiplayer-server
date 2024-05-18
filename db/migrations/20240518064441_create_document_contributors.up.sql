CREATE TABLE IF NOT EXISTS document_contributors (
    document_id uuid NOT NULL REFERENCES documents(id),
    user_id uuid NOT NULL REFERENCES users(id),
    PRIMARY KEY(document_id, user_id)
);
