-- name: CreateDocument :one
INSERT INTO documents (name)
    VALUES ($1)
RETURNING *;
