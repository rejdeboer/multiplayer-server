// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.25.0

package db

import (
	"github.com/google/uuid"
)

type Document struct {
	ID          uuid.UUID
	Name        string
	OwnerID     uuid.UUID
	StateVector []byte
}

type DocumentContributor struct {
	DocumentID uuid.UUID
	UserID     uuid.UUID
}

type DocumentUpdate struct {
	DocumentID uuid.UUID
	Clock      int32
	Value      []byte
}

type User struct {
	ID       uuid.UUID
	Email    string
	Username string
	Passhash string
	ImageUrl *string
}
