package application

import (
	"fmt"
	"net/http"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rejdeboer/multiplayer-server/internal/configuration"
	"github.com/rejdeboer/multiplayer-server/internal/logger"
	"github.com/rejdeboer/multiplayer-server/internal/routes"
	"github.com/segmentio/kafka-go"
)

var log = logger.Get()

type Application struct {
	pool     *pgxpool.Pool
	producer *kafka.Writer
	handler  http.Handler
	addr     string
}

func Build(settings configuration.Settings) Application {
	port := settings.Application.Port
	addr := fmt.Sprintf(":%d", port)

	pool := GetDbConnectionPool(settings.Database)

	producer := &kafka.Writer{
		Addr:     kafka.TCP(settings.Application.KafkaEndpoint),
		Balancer: &kafka.LeastBytes{},
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

func GetSearchClient(endpoint string) *elasticsearch.TypedClient {
	client, err := elasticsearch.NewTypedClient(elasticsearch.Config{
		Addresses: []string{endpoint},
	})
	if err != nil {
		log.Fatal().Err(err).Msg("error creating search client")
	}
	return client
}
