package document

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rejdeboer/multiplayer-server/internal/db"
)

type Doc struct {
	ID          uuid.UUID
	StateVector []byte
}

func (doc Doc) StoreUpdate(pool *pgxpool.Pool, value []byte) error {
	ctx := context.Background()

	tx, err := pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	q := db.New(pool).WithTx(tx)

	currentClock, err := q.GetDocumentClock(ctx, doc.ID)
	if err != nil {
		return err
	}

	if err = q.CreateDocumentUpdate(ctx, db.CreateDocumentUpdateParams{
		DocumentID: doc.ID,
		Clock:      currentClock.(int32) + 1,
		Value:      value,
	}); err != nil {
		return err
	}

	return tx.Commit(ctx)
}
