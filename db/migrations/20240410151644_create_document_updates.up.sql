CREATE TABLE document_updates (
    document_id uuid NOT NULL,
    clock integer NOT NULL,
    value bytea NOT NULL,
    PRIMARY KEY(document_id, clock),
    UNIQUE(document_id, value)
);
