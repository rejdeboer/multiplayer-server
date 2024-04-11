package websocket

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rejdeboer/multiplayer-server/internal/logger"
	"github.com/rs/xid"
	"github.com/rs/zerolog"
)

type Context struct {
	Log    zerolog.Logger
	UserID string
	Pool   *pgxpool.Pool
}

func CreateContext(ctx context.Context, docID uuid.UUID) Context {
	userID := ctx.Value("user_id").(string)
	log := logger.Get().With().
		Str("user_id", userID).
		Str("document_id", docID.String()).
		Str("cid", xid.New().String()).
		Logger()

	return Context{
		Log:    log,
		UserID: userID,
		Pool:   ctx.Value("pool").(*pgxpool.Pool),
	}
}
