-- name: CreateUser :one
INSERT INTO users (email, username, passhash)
    VALUES ($1, $2, $3)
RETURNING *;
