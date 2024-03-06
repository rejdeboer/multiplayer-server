package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/joho/godotenv"
	"github.com/rejdeboer/multiplayer-server/internal/logger"
	"github.com/rejdeboer/multiplayer-server/internal/routes"
)

func main() {
	godotenv.Load(".env")
	logger := logger.Get()

	port := 8080
	addr := fmt.Sprintf(":%d", port)

	handler := withLogging(routes.NewRouter())

	logger.Info().Msg(fmt.Sprintf("Server listening on port %d", port))

	err := http.ListenAndServe(addr, handler)
	if err != nil {
		panic(err)
	}
}

func withLogging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		l := logger.Get()

		defer func() {
			l.
				Info().
				Str("method", r.Method).
				Str("url", r.RequestURI).
				Dur("elapsed_ms", time.Since(start)).
				Msg("incoming request")
		}()

		next.ServeHTTP(w, r)
	})

}
