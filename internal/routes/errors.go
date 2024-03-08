package routes

import "net/http"

func internalServerError(w http.ResponseWriter) {
	http.Error(w, "an unexpected error occured, please try again later", http.StatusInternalServerError)
}
