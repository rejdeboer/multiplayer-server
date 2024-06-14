package application

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rejdeboer/multiplayer-server/internal/configuration"
)

func GetDbConnectionPool(settings configuration.DatabaseSettings) *pgxpool.Pool {
	pool, err := pgxpool.New(context.Background(), GetDbConnectionString(settings))
	if err != nil {
		log.Error().Msg("failed to connect to db")
		panic(err)
	}

	return pool
}

func GetDbConnectionString(settings configuration.DatabaseSettings) string {
	dbUrl := fmt.Sprintf("postgresql://%s:%s@%s:%d/%s",
		settings.Username,
		settings.Password,
		settings.Host,
		settings.Port,
		settings.DbName,
	)

	if !settings.RequireSsl {
		dbUrl = dbUrl + "?sslmode=disable"
	}
	// else {
	// 	dbUrl = dbUrl + "?sslmode=verify-full&sslrootcert=" + settings.CertificatePath
	// }

	return dbUrl
}
