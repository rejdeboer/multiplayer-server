package websocket

import (
	"context"

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

func CreateContext(ctx context.Context) Context {
	userID := ctx.Value("user_id").(string)
	log := logger.Get().With().
		Str("user_id", userID).
		Str("cid", xid.New().String()).
		Logger()

	return Context{
		Log:    log,
		UserID: userID,
		Pool:   ctx.Value("pool").(*pgxpool.Pool),
	}
}
