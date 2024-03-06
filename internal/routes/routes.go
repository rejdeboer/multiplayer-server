package routes

import (
	"github.com/rs/zerolog/log"
	"net/http"
)

func NewRouter() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Info().Msg("Hello from index")
	})

	return mux
}
