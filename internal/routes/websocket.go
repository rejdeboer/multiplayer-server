package routes

import (
	"net/http"

	gwebsocket "github.com/gorilla/websocket"
	"github.com/rejdeboer/multiplayer-server/internal/websocket"
	"github.com/rejdeboer/multiplayer-server/pkg/httperrors"
	"github.com/rs/zerolog"
)

var upgrader = gwebsocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func handleWebSocket(
	hub *websocket.Hub,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		log := zerolog.Ctx(ctx)

		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			httperrors.InternalServerError(w)
			log.Error().Err(err).Msg("websocket upgrade error")
			return
		}
		defer conn.Close()

		client := &websocket.Client{
			Context: websocket.CreateContext(ctx),
			Hub:     hub,
			Conn:    conn,
			Send:    make(chan []byte, 256),
		}
		client.Hub.Register <- client
		log.Info().Msg("new client registered")

		go client.WritePump()
		client.ReadPump()
	}
}
