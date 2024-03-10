// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.25.0
// source: get_user_by_email.sql

package db

import (
	"context"
)

const getUserByEmail = `-- name: GetUserByEmail :one
SELECT id, email, username, passhash FROM users WHERE email = $1
`

func (q *Queries) GetUserByEmail(ctx context.Context, email string) (User, error) {
	row := q.db.QueryRow(ctx, getUserByEmail, email)
	var i User
	err := row.Scan(
		&i.ID,
		&i.Email,
		&i.Username,
		&i.Passhash,
	)
	return i, err
}
