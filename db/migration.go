package main

import (
	"context"
	"database/sql"
	"log"
	"os"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/rejdeboer/multiplayer-server/internal/application"
	"github.com/rejdeboer/multiplayer-server/internal/configuration"
)

func main() {
	settings := configuration.ReadConfiguration("./configuration")

	environment := os.Getenv("ENVIRONMENT")
	if environment != "local" && environment != "" {
		settings.Database.Password = getAzAccessToken()
	}

	db, err := sql.Open("pgx", application.GetDbConnectionString(settings.Database))
	if err != nil {
		log.Fatalf("error open db connection: %s", err)
	}
	defer db.Close()

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		log.Fatalf("could not init driver: %s", err)
	}

	migrate, err := migrate.NewWithDatabaseInstance(
		"file://migrations",
		"pgx", driver)
	if err != nil {
		log.Fatalf("could not apply the migration: %s", err)
	}

	migrate.Up()
	log.Println("migrated database")
}

func getAzAccessToken() string {
	credential, err := azidentity.NewWorkloadIdentityCredential(nil)
	if err != nil {
		log.Fatalf("could not get azwi credential: %s", err)
	}
	token, err := credential.GetToken(context.Background(), policy.TokenRequestOptions{})
	if err != nil {
		log.Fatalf("could not get az access token: %s", err)
	}
	return token.Token
}
