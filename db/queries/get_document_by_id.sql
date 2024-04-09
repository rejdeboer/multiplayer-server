-- name: GetDocumentByID :one
SELECT * FROM documents WHERE id = $1 AND owner_ID = $2;
