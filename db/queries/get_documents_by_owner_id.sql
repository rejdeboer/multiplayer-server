-- name: GetDocumentsByOwnerID :many
SELECT * FROM documents WHERE owner_id = $1;
