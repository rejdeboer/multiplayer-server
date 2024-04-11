-- name: GetDocumentByID :one
SELECT * FROM documents WHERE id = $1;
