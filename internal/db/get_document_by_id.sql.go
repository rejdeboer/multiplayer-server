// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.25.0
// source: get_document_by_id.sql

package db

import (
	"context"

	"github.com/google/uuid"
)

const getDocumentByID = `-- name: GetDocumentByID :one
SELECT id, name, owner_id, shared_with, state_vector FROM documents WHERE id = $1 AND owner_ID = $2
`

type GetDocumentByIDParams struct {
	ID      uuid.UUID
	OwnerID uuid.UUID
}

func (q *Queries) GetDocumentByID(ctx context.Context, arg GetDocumentByIDParams) (Document, error) {
	row := q.db.QueryRow(ctx, getDocumentByID, arg.ID, arg.OwnerID)
	var i Document
	err := row.Scan(
		&i.ID,
		&i.Name,
		&i.OwnerID,
		&i.SharedWith,
		&i.StateVector,
	)
	return i, err
}
