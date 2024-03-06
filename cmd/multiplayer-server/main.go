package main

import (
	"fmt"
	"net/http"

	"github.com/joho/godotenv"
	"github.com/rejdeboer/multiplayer-server/internal/logger"
	"github.com/rejdeboer/multiplayer-server/internal/middleware"
	"github.com/rejdeboer/multiplayer-server/internal/routes"
)

func main() {
	godotenv.Load(".env")
	logger := logger.Get()

	port := 8080
	addr := fmt.Sprintf(":%d", port)

	handler := middleware.WithMiddleware(routes.NewRouter())

	logger.Info().Msg(fmt.Sprintf("Server listening on port %d", port))

	err := http.ListenAndServe(addr, handler)
	if err != nil {
		panic(err)
	}
}
