package routes

import (
	"net/http"

	"github.com/rejdeboer/multiplayer-server/internal/configuration"
	"github.com/rs/zerolog"
)

func NewRouter(settings configuration.ApplicationSettings) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		l := zerolog.Ctx(r.Context())
		l.Info().Msg("Hello from index")
		w.WriteHeader(http.StatusOK)
	})

	mux.HandleFunc("POST /user", createUser)
	mux.HandleFunc("POST /token", getToken(settings.SigningKey))

	return mux
}
