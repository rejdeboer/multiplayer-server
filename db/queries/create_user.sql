-- name: CreateUser :one
INSERT INTO users (email, passhash)
    VALUES ($1, $2)
RETURNING *;
