package routes

import (
	"net/http"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/rejdeboer/multiplayer-server/internal/configuration"
	"github.com/rejdeboer/multiplayer-server/internal/routes/middleware"
	"github.com/rs/zerolog"
)

func NewRouter(settings configuration.ApplicationSettings, producer *kafka.Producer) http.Handler {
	authorized := middleware.WithAuth(settings.SigningKey)

	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		l := zerolog.Ctx(r.Context())
		l.Info().Msg("Hello from index")
		w.WriteHeader(http.StatusOK)
	})

	mux.HandleFunc("GET /document", authorized(listDocuments))
	mux.HandleFunc("GET /document/{id}", authorized(getDocument))
	mux.HandleFunc("DELETE /document/{id}", authorized(deleteDocument))
	mux.HandleFunc("POST /document", authorized(createDocument))

	mux.HandleFunc("POST /document/{document_id}/contributor", authorized(addContributor))

	mux.HandleFunc("POST /user", createUser(producer))
	mux.HandleFunc("POST /token", getToken(settings.SigningKey, settings.TokenExpirationSeconds))

	return mux
}
