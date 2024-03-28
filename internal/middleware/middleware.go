package middleware

import (
	"context"
	"net/http"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rejdeboer/multiplayer-server/internal/configuration"
	"github.com/rejdeboer/multiplayer-server/internal/logger"
	"github.com/rs/zerolog/hlog"
)

func WithLogging(next http.Handler) http.Handler {
	l := logger.Get()
	hlogHandler := hlog.NewHandler(l)

	accessHandler := hlog.AccessHandler(func(r *http.Request, status, size int, duration time.Duration) {
		hlog.FromRequest(r).Info().
			Str("method", r.Method).
			Stringer("url", r.URL).
			Int("status_code", status).
			Dur("elapsed_ms", duration).
			Msg("")
	})

	userAgentHandler := hlog.UserAgentHandler("user_agent")
	requestIdHandler := hlog.RequestIDHandler("req_id", "Request-Id")

	return hlogHandler(accessHandler(userAgentHandler(requestIdHandler(next))))
}

func WithDb(next http.Handler, pool *pgxpool.Pool) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), "pool", pool)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func WithBlobStorage(next http.Handler, settings configuration.AzureSettings) http.Handler {
	l := logger.Get()

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		credentials, err := azidentity.NewDefaultAzureCredential(nil)
		if err != nil {
			l.Error().Err(err).Msg("error getting azure blob storage credentials")
			panic(err)
		}

		url := settings.BlobStorageEndpoint + settings.StorageAccountName
		client, err := azblob.NewClient(url, credentials, nil)
		if err != nil {
			l.Error().Err(err).Msg("error creating azure blob storage client")
			panic(err)
		}

		ctx := context.WithValue(r.Context(), "blob_storage", client)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
