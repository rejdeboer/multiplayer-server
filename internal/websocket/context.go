package websocket

import (
	"context"

	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/rejdeboer/multiplayer-server/internal/logger"
	"github.com/rs/zerolog"
)

type Context struct {
	Log        zerolog.Logger
	UserID     string
	BlobClient *azblob.Client
}

func CreateContext(ctx context.Context) Context {
	userID := ctx.Value("user_id").(string)
	log := logger.Get().With().
		Str("user_id", userID).
		Logger()

	return Context{
		Log:        log,
		UserID:     userID,
		BlobClient: ctx.Value("azblob").(*azblob.Client),
	}
}
