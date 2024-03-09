package routes

import (
	"encoding/json"
	"github.com/rejdeboer/multiplayer-server/internal/logger"
	"net/http"
)

var log = logger.Get()

type ErrorResponse struct {
	Message string
	Status  int
}

func internalServerError(w http.ResponseWriter) {
	writeError(w, "an unexpected error occured, please try again later", http.StatusInternalServerError)
}

func writeError(w http.ResponseWriter, message string, code int) {
	w.Header().Set("Content-Type", "text/json; charset=utf-8")
	w.WriteHeader(code)

	response, err := json.Marshal(ErrorResponse{
		Message: message,
		Status:  code,
	})
	if err != nil {
		log.Error().Err(err).Msg("error marshalling error response")
		return
	}

	w.Write(response)
}
