package routes

import (
	"net/http"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rejdeboer/multiplayer-server/internal/configuration"
	"github.com/rejdeboer/multiplayer-server/internal/middleware"
	"github.com/rs/cors"
	"github.com/rs/zerolog"
)

type Env struct {
	Pool     *pgxpool.Pool
	Producer *kafka.Producer
}

func CreateHandler(settings configuration.Settings, env *Env) http.Handler {
	authorized := middleware.WithAuth(settings.Application.SigningKey)

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		l := zerolog.Ctx(r.Context())
		l.Info().Msg("Hello from index")
		w.WriteHeader(http.StatusOK)
	})

	mux.HandleFunc("GET /document", authorized(env.listDocuments))
	mux.HandleFunc("GET /document/{id}", authorized(env.getDocument))
	mux.HandleFunc("DELETE /document/{id}", authorized(env.deleteDocument))
	mux.HandleFunc("POST /document", authorized(env.createDocument))

	mux.HandleFunc("POST /document/{document_id}/contributor", authorized(env.addDocumentContributor))

	mux.HandleFunc("POST /user", env.createUser)
	mux.HandleFunc("POST /token", env.getToken(settings.Application.SigningKey, settings.Application.TokenExpirationSeconds))

	handler := middleware.WithLogging(mux)
	handler = middleware.WithBlobStorage(handler, settings.Azure)

	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3000"},
		AllowedHeaders:   []string{"*"},
		AllowedMethods:   []string{http.MethodGet, http.MethodPost, http.MethodDelete, http.MethodOptions},
		AllowCredentials: true,
	})

	handler = c.Handler(handler)

	return handler
}
