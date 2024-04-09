-- name: CreateDocument :one
INSERT INTO documents (name, owner_id)
    VALUES ($1, $2)
RETURNING *;
