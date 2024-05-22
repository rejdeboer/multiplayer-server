package application

import (
	"context"
	"fmt"
	"net/http"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rejdeboer/multiplayer-server/internal/configuration"
	"github.com/rejdeboer/multiplayer-server/internal/logger"
	"github.com/rejdeboer/multiplayer-server/internal/routes"
)

var log = logger.Get()

type Application struct {
	pool     *pgxpool.Pool
	producer *kafka.Producer
	handler  http.Handler
	addr     string
}

func Build(settings configuration.Settings) Application {
	port := settings.Application.Port
	addr := fmt.Sprintf(":%d", port)

	pool := getDbConnectionPool(settings.Database)

	producer, err := kafka.NewProducer(&kafka.ConfigMap{
		"bootstrap.servers": settings.Application.KafkaEndpoint,
	})
	if err != nil {
		log.Fatal().Err(err).Msg("error creating kafka producer")
	}

	searchClient, err := elasticsearch.NewTypedClient(elasticsearch.Config{
		Addresses: []string{settings.Application.ElasticsearchEndpoint},
	})
	if err != nil {
		log.Fatal().Err(err).Msg("error creating elasticsearch client")
	}

	handler := routes.CreateHandler(settings, &routes.Env{
		Pool:         pool,
		Producer:     producer,
		Blob:         GetBlobClient(settings.Azure),
		SearchClient: searchClient,
	})

	return Application{
		addr:     addr,
		pool:     pool,
		producer: producer,
		handler:  handler,
	}
}

func (app *Application) Start() error {
	defer app.close()
	log.Info().Msg(fmt.Sprintf("Server listening on port %s", app.addr))
	return http.ListenAndServe(app.addr, app.handler)
}

func (app *Application) close() {
	app.pool.Close()
	app.producer.Close()
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
