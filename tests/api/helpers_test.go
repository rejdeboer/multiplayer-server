package api

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/brianvoe/gofakeit"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
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
var testApp TestApp
var dbpool *pgxpool.Pool

type TestApp struct {
	handler  http.Handler
	user     routes.UserCreate
	settings configuration.ApplicationSettings
}

func GetTestApp() TestApp {
	once.Do(func() {
		settings := configuration.ReadConfiguration("../../configuration")

		handler := routes.NewRouter(settings.Application)
		handler = middleware.WithLogging(handler)
		handler = middleware.WithDb(handler, dbpool)

		user := createTestUser()

		testApp = TestApp{
			handler:  handler,
			user:     user,
			settings: settings.Application,
		}
	})

	return testApp
}

func TestMain(m *testing.M) {
	dockerPool := initDocker()

	postgresContainer := createPostgresContainer(dockerPool)
	hostAndPort := postgresContainer.GetHostPort("5432/tcp")
	databaseUrl := fmt.Sprintf("postgres://postgres:postgres@%s/multiplayer?sslmode=disable", hostAndPort)

	if err := dockerPool.Retry(waitPostgresContainerToBeReady(databaseUrl)); err != nil {
		log.Fatalf("postgres container not intialized: %s", err)
	}

	startMigration(databaseUrl)

	code := m.Run()

	if err := dockerPool.Purge(postgresContainer); err != nil {
		log.Fatalf("could not purge postgres: %s", err)
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

func createTestUser() routes.UserCreate {
	q := db.New(dbpool)

	password := gofakeit.Password(true, true, true, true, false, 8)
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Fatalf("error hashing test user password: %s", err)
	}

	passhash := string(bytes)

	user, err := q.CreateUser(context.Background(), db.CreateUserParams{
		Email:    gofakeit.Email(),
		Passhash: passhash,
	})
	if err != nil {
		log.Fatalf("error storing test user in db: %s", err)
	}

	return routes.UserCreate{
		Email:    user.Email,
		Password: password,
	}
}
