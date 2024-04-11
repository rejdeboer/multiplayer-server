package sync

import (
	"context"
	"fmt"
	"slices"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rejdeboer/multiplayer-server/internal/db"
)

type Doc struct {
	ID          uuid.UUID
	StateVector []byte
}

func FetchDoc(pool *pgxpool.Pool, docID uuid.UUID, userID uuid.UUID) (Doc, error) {
	q := db.New(pool)
	dbDoc, err := q.GetDocumentByID(context.Background(), docID)
	if err != nil {
		return Doc{}, err
	}

	if dbDoc.OwnerID != userID && !slices.Contains(dbDoc.SharedWith, userID) {
		return Doc{}, fmt.Errorf("user %v does not have access to document %v", userID.String(), docID.String())
	}

	doc := Doc{
		ID:          dbDoc.ID,
		StateVector: dbDoc.StateVector,
	}

	return doc, nil
}
