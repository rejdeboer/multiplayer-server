package routes

import (
	"github.com/rs/zerolog"
	"net/http"
)

func NewRouter() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		l := zerolog.Ctx(r.Context())
		l.Info().Msg("Hello from index")
		w.WriteHeader(http.StatusOK)
	})

	return mux
}
