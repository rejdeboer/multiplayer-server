package routes

import (
	"net/http"
	"strings"

	"github.com/dgrijalva/jwt-go"
	gwebsocket "github.com/gorilla/websocket"
	"github.com/rejdeboer/multiplayer-server/internal/configuration"
	"github.com/rejdeboer/multiplayer-server/internal/websocket"
	"github.com/rs/zerolog"
)

var upgrader = gwebsocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func handleWebSocket(
	hub *websocket.Hub,
	appSettings configuration.ApplicationSettings,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		log := zerolog.Ctx(ctx)

		authHeader := r.Header["Authorization"]
		if authHeader == nil || len(authHeader) == 0 {
			writeError(w, "please provide an 'Authorization' header", http.StatusUnauthorized)
			log.Error().Msg("received request without auth header")
			return
		}

		authHeaderParts := strings.Split(authHeader[0], " ")
		if len(authHeaderParts) != 2 || authHeaderParts[0] != "Bearer" {
			writeError(w, "please provide a valid bearer token in the auth header", http.StatusUnauthorized)
			log.Error().Str("header", authHeader[0]).Msg("malformed auth header")
			return
		}

		token := authHeaderParts[1]
		_, err := verifyAndGetUser(token, appSettings.SigningKey)
		if err != nil {
			writeError(w, "invalid jwt token", http.StatusUnauthorized)
			log.Error().Err(err).Msg("invalid jwt token")
			return
		}

		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			internalServerError(w)
			log.Error().Err(err).Msg("websocket upgrade error")
			return
		}
		defer conn.Close()

		client := &websocket.Client{Hub: hub, Conn: conn, Send: make(chan []byte, 256)}
		client.Hub.Register <- client

		go client.WritePump()
		go client.ReadPump()
	}
}

func verifyAndGetUser(token string, verificationSecret string) (string, error) {
	claims := jwt.MapClaims{}
	_, err := jwt.ParseWithClaims(token, claims, func(t *jwt.Token) (interface{}, error) {
		return []byte(verificationSecret), nil
	})
	if err != nil {
		return "", err
	}

	return claims["username"].(string), nil
}
