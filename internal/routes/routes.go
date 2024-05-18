package routes

import (
	"net/http"

	"github.com/rejdeboer/multiplayer-server/internal/configuration"
	"github.com/rejdeboer/multiplayer-server/internal/routes/middleware"
	"github.com/rs/zerolog"
)

func NewRouter(settings configuration.ApplicationSettings) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		l := zerolog.Ctx(r.Context())
		l.Info().Msg("Hello from index")
		w.WriteHeader(http.StatusOK)
	})

	mux.HandleFunc("GET /document", middleware.WithAuth(listDocuments, settings.SigningKey))
	mux.HandleFunc("GET /document/{id}", middleware.WithAuth(getDocument, settings.SigningKey))
	mux.HandleFunc("DELETE /document/{id}", middleware.WithAuth(deleteDocument, settings.SigningKey))
	mux.HandleFunc("POST /document", middleware.WithAuth(createDocument, settings.SigningKey))

	mux.HandleFunc("POST /document/{document_id}/contributor/{user_id}", middleware.WithAuth(addContributor, settings.SigningKey))

	mux.HandleFunc("POST /user", createUser)
	mux.HandleFunc("POST /token", getToken(settings.SigningKey, settings.TokenExpirationSeconds))

	return mux
}
