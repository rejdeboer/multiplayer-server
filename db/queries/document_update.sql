-- name: GetDocumentClock :one
SELECT COALESCE(MAX(clock), 0) 
FROM document_updates
WHERE document_id = $1;

-- name: CreateDocumentUpdate :exec
INSERT INTO document_updates (document_id, clock, value)
    VALUES ($1, $2, $3);
