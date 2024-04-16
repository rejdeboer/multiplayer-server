package main

import (
	"context"
	"fmt"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rejdeboer/multiplayer-server/internal/configuration"
	"github.com/rejdeboer/multiplayer-server/internal/logger"
	"github.com/rejdeboer/multiplayer-server/internal/middleware"
	"github.com/rejdeboer/multiplayer-server/internal/routes"
)

var log = logger.Get()

func main() {
	settings := configuration.ReadConfiguration("./configuration")

	port := settings.Application.Port
	addr := fmt.Sprintf(":%d", port)

	pool := getDbConnectionPool(settings.Database)
	defer pool.Close()

	handler := createHandler(settings, pool)

	log.Info().Msg(fmt.Sprintf("Server listening on port %d", port))

	err := http.ListenAndServe(addr, handler)
	if err != nil {
		log.Error().Msg("failed to start server")
		panic(err)
	}
}

func createHandler(settings configuration.Settings, pool *pgxpool.Pool) http.Handler {
	handler := routes.NewRouter(settings.Auth)
	handler = middleware.WithLogging(handler)
	handler = middleware.WithDb(handler, pool)
	handler = middleware.WithBlobStorage(handler, settings.Azure)

	return handler
}

func getDbConnectionPool(settings configuration.DatabaseSettings) *pgxpool.Pool {
	dbUrl := fmt.Sprintf("postgresql://%s:%s@%s:%d/%s",
		settings.Username,
		settings.Password,
		settings.Host,
		settings.Port,
		settings.DbName,
	)

	if !settings.RequireSsl {
		dbUrl = dbUrl + "?sslmode=disable"
	} else {
		dbUrl = dbUrl + "?sslmode=verify-full&sslrootcert=" + settings.CertificatePath
	}

	pool, err := pgxpool.New(context.Background(), dbUrl)
	if err != nil {
		log.Error().Msg("failed to connect to db")
		panic(err)
	}

	return pool
}
