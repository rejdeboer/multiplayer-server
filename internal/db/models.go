// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.25.0

package db

import (
	"github.com/jackc/pgx/v5/pgtype"
)

type Document struct {
	ID      pgtype.UUID
	OwnerID pgtype.UUID
	Content []byte
}

type User struct {
	ID       pgtype.UUID
	Email    string
	Username string
	Passhash string
}
