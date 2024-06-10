package application

import (
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/rejdeboer/multiplayer-server/internal/configuration"
	"github.com/rejdeboer/multiplayer-server/internal/logger"
)

func GetBlobClient(settings configuration.AzureSettings) *azblob.Client {
	l := logger.Get()

	var client *azblob.Client
	var err error
	if settings.BlobConnectionString != "" {
		// Local
		client, err = azblob.NewClientFromConnectionString(settings.BlobConnectionString, nil)
	} else {
		// Production
		credential, err := azidentity.NewDefaultAzureCredential(nil)
		if err != nil {
			l.Fatal().Err(err).Msg("error getting azure blob storage credentials")
		}
		client, err = azblob.NewClient(settings.BlobStorageEndpoint, credential, nil)
	}
	if err != nil {
		l.Fatal().Err(err).Msg("error creating azure blob storage client")
	}

	return client
}
