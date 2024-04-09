package routes

import (
	"net/http"

	"github.com/rejdeboer/multiplayer-server/internal/configuration"
	"github.com/rejdeboer/multiplayer-server/internal/routes/middleware"
	"github.com/rejdeboer/multiplayer-server/internal/websocket"
	"github.com/rs/zerolog"
)

func NewRouter(settings configuration.ApplicationSettings) http.Handler {
	mux := http.NewServeMux()

	hub := websocket.NewHub()
	go hub.Run()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		l := zerolog.Ctx(r.Context())
		l.Info().Msg("Hello from index")
		w.WriteHeader(http.StatusOK)
	})

	mux.HandleFunc("GET /document", middleware.WithAuth(listDocuments, settings.SigningKey))
	mux.HandleFunc("POST /document", middleware.WithAuth(createDocument, settings.SigningKey))

	mux.HandleFunc("POST /user", createUser)
	mux.HandleFunc("POST /token", getToken(settings.SigningKey))
	mux.HandleFunc("GET /websocket", middleware.WithAuth(handleWebSocket(hub), settings.SigningKey))

	return mux
}
