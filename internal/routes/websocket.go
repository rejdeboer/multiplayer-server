package routes

import (
	"net/http"

	"github.com/google/uuid"
	gwebsocket "github.com/gorilla/websocket"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rejdeboer/multiplayer-server/internal/sync"
	"github.com/rejdeboer/multiplayer-server/internal/websocket"
	"github.com/rejdeboer/multiplayer-server/pkg/httperrors"
	"github.com/rs/zerolog"
)

var upgrader = gwebsocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func handleSync(
	hub *websocket.Hub,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		log := zerolog.Ctx(ctx)

		docID, err := uuid.Parse(r.PathValue("document-id"))
		if err != nil {
			httperrors.Write(w, "please provide a valid uuid", http.StatusBadRequest)
			log.Error().Err(err).Msg("invalid uuid")
			return
		}

		userID, _ := uuid.Parse(ctx.Value("user_id").(string))
		pool := ctx.Value("pool").(*pgxpool.Pool)

		document, err := sync.FetchDoc(pool, docID, userID)
		// TODO: Handle internal server error
		if err != nil {
			httperrors.Write(w, "document not found", http.StatusNotFound)
			log.Error().Err(err).Str("document_id", docID.String()).Msg("document not found")
			return
		}

		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			httperrors.InternalServerError(w)
			log.Error().Err(err).Msg("websocket upgrade error")
			return
		}
		defer conn.Close()

		client := &websocket.Client{
			Context: websocket.CreateContext(ctx, docID),
			// TODO: Create session struct that manages clients connected to the same document
			Doc: sync.Doc{
				ID:          docID,
				StateVector: document.StateVector,
			},
			Hub:  hub,
			Conn: conn,
			Send: make(chan []byte, 256),
		}
		client.Hub.Register <- client
		log.Info().Str("document_id", docID.String()).Msg("new client registered")

		go client.WritePump()
		client.ReadPump()
	}
}
