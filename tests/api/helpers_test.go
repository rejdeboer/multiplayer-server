package api

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/brianvoe/gofakeit"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"github.com/rejdeboer/multiplayer-server/internal/configuration"
	"github.com/rejdeboer/multiplayer-server/internal/db"
	"github.com/rejdeboer/multiplayer-server/internal/middleware"
	"github.com/rejdeboer/multiplayer-server/internal/routes"
	"golang.org/x/crypto/bcrypt"
)

var once sync.Once
var settings configuration.Settings
var handler http.Handler
var dbpool *pgxpool.Pool
var blobHostAndPort string

type TestApp struct {
	handler  http.Handler
	user     TestUser
	document routes.DocumentResponse
	token    string
	settings configuration.ApplicationSettings
}

type TestUser struct {
	ID       uuid.UUID
	Email    string
	Username string
	Password string
}

func GetTestApp() TestApp {
	once.Do(func() {
		settings = configuration.ReadConfiguration("../../configuration")
		settings.Azure.BlobConnectionString = strings.ReplaceAll(settings.Azure.BlobConnectionString, "https", "http")
		settings.Azure.BlobConnectionString = strings.ReplaceAll(settings.Azure.BlobConnectionString, "azurite:10000", blobHostAndPort)

		handler = routes.NewRouter(settings.Application, nil)
		handler = middleware.WithLogging(handler)
		handler = middleware.WithDb(handler, dbpool)
		handler = middleware.WithBlobStorage(handler, settings.Azure)
	})

	user := createTestUser()
	token, err := routes.GetJwt(
		settings.Application.SigningKey,
		settings.Application.TokenExpirationSeconds,
		user.ID.String(),
		user.Username,
	)
	if err != nil {
		log.Fatalf("error creating test token: %s", err)
	}

	return TestApp{
		handler:  handler,
		user:     user,
		document: createTestDocument(user.ID),
		token:    token,
		settings: settings.Application,
	}
}

func TestMain(m *testing.M) {
	dockerPool := initDocker()

	postgresContainer := createPostgresContainer(dockerPool)
	databaseHostAndPort := postgresContainer.GetHostPort("5432/tcp")
	databaseUrl := fmt.Sprintf("postgres://postgres:postgres@%s/multiplayer?sslmode=disable", databaseHostAndPort)

	if err := dockerPool.Retry(waitPostgresContainerToBeReady(databaseUrl)); err != nil {
		log.Fatalf("postgres container not intialized: %s", err)
	}

	startMigration(databaseUrl)

	azuriteContainer := createAzuriteContainer(dockerPool)
	blobHostAndPort = azuriteContainer.GetHostPort("10000/tcp")
	if err := dockerPool.Retry(waitAzuriteContainerToBeReady(blobHostAndPort)); err != nil {
		log.Fatalf("azurite container not intialized: %s", err)
	}

	code := m.Run()

	if err := dockerPool.Purge(postgresContainer); err != nil {
		log.Fatalf("could not purge postgres: %s", err)
	}

	if err := dockerPool.Purge(azuriteContainer); err != nil {
		log.Fatalf("could not purge azurite: %s", err)
	}

	os.Exit(code)
}

func initDocker() *dockertest.Pool {
	pool, err := dockertest.NewPool("")
	if err != nil {
		log.Fatalf("could not construct docker pool: %s", err)
	}

	err = pool.Client.Ping()
	if err != nil {
		log.Fatalf("could not connect to Docker: %s", err)
	}

	pool.MaxWait = 120 * time.Second
	return pool
}

func createPostgresContainer(dockerPool *dockertest.Pool) *dockertest.Resource {
	container, err := dockerPool.RunWithOptions(&dockertest.RunOptions{
		Repository: "postgres",
		Tag:        "15-alpine3.18",
		Env: []string{
			"POSTGRES_PASSWORD=postgres",
			"POSTGRES_USER=postgres",
			"POSTGRES_DB=multiplayer",
			"listen_addresses = '*'",
		},
	}, func(config *docker.HostConfig) {
		config.AutoRemove = true
		config.RestartPolicy = docker.RestartPolicy{Name: "no"}
	})
	if err != nil {
		log.Fatalf("could not start postgres: %s", err)
	}

	container.Expire(120)
	return container
}

func createAzuriteContainer(dockerPool *dockertest.Pool) *dockertest.Resource {
	container, err := dockerPool.RunWithOptions(&dockertest.RunOptions{
		Repository: "mcr.microsoft.com/azure-storage/azurite",
		Tag:        "latest",
		Env: []string{
			"listen_addresses = '*'",
		},
	}, func(config *docker.HostConfig) {
		config.AutoRemove = true
		config.RestartPolicy = docker.RestartPolicy{Name: "no"}
	})
	if err != nil {
		log.Fatalf("could not start postgres: %s", err)
	}

	container.Expire(120)
	return container
}

func waitPostgresContainerToBeReady(url string) func() error {
	return func() error {
		ctx := context.Background()
		var err error
		dbpool, err = pgxpool.New(ctx, url)
		if err != nil {
			return err
		}

		return dbpool.Ping(ctx)
	}
}

func waitAzuriteContainerToBeReady(address string) func() error {
	return func() error {
		_, err := net.Dial("tcp", address)
		return err
	}
}

func startMigration(databaseUrl string) {
	db, err := sql.Open("pgx", databaseUrl)
	if err != nil {
		log.Fatalf("error open connection to apply migration: %s", err)
	}

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		log.Fatalf("could not init driver: %s", err)
	}

	migrate, err := migrate.NewWithDatabaseInstance(
		"file://../../db/migrations",
		"pgx", driver)
	if err != nil {
		log.Fatalf("could not apply the migration: %s", err)
	}

	migrate.Up()
}

func createTestUser() TestUser {
	q := db.New(dbpool)

	password := gofakeit.Password(true, true, true, true, false, 8)
	username := gofakeit.Username()
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Fatalf("error hashing test user password: %s", err)
	}

	passhash := string(bytes)

	user, err := q.CreateUser(context.Background(), db.CreateUserParams{
		Email:    gofakeit.Email(),
		Username: username,
		Passhash: passhash,
	})
	if err != nil {
		log.Fatalf("error storing test user in db: %s", err)
	}

	return TestUser{
		ID:       user.ID,
		Email:    user.Email,
		Username: user.Username,
		Password: password,
	}
}

func createTestDocument(ownerID uuid.UUID) routes.DocumentResponse {
	q := db.New(dbpool)

	name := gofakeit.Name()

	document, err := q.CreateDocument(context.Background(), db.CreateDocumentParams{
		OwnerID: ownerID,
		Name:    name,
	})
	if err != nil {
		log.Fatalf("error storing test document in db: %s", err)
	}

	err = q.CreateDocumentContributor(context.Background(), db.CreateDocumentContributorParams{
		DocumentID: document.ID,
		UserID:     ownerID,
	})
	if err != nil {
		log.Fatalf("error storing owner as contributor in db: %s", err)
	}

	return routes.DocumentResponse{
		ID:   document.ID,
		Name: document.Name,
	}
}
