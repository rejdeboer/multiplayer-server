package sync

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rejdeboer/multiplayer-server/internal/db"
	"github.com/rejdeboer/multiplayer-server/pkg/reader"
)

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

// NOTE: Y-CRDT update v1 encoding format:
// 1. `clientsLen` | max 4 bytes
// 2. For each client:
//   - `blocksLen` | max 4 bytes
//   - `client` | max 4 bytes
//   - `clock` | max 4 bytes | starting clock for blocks
func DecodeUpdate(buf []byte) {
	r := reader.FromBuffer(buf)
	clientsLen, _ := r.ReadU32()
	clients := make(map[uint32]string, clientsLen)

	for _ = range clientsLen {
		blocksLen, _ := r.ReadU32()
		client, _ := r.ReadU32()
		clock, _ := r.ReadU32()

		blocks := make([]string, blocksLen)
		for i := range blocksLen {

		}
	}
}
