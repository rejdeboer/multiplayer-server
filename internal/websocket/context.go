package websocket

import (
	"context"

	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
)

type Context struct {
	UserID     string
	BlobClient *azblob.Client
}

func FromCtx(ctx context.Context) Context {
	return Context{
		UserID:     ctx.Value("user_id").(string),
		BlobClient: ctx.Value("azblob").(*azblob.Client),
	}
}
