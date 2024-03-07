package api

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCreateUser(t *testing.T) {
	// server := httptest.NewServer(GetTestingHandler())
	// defer server.Close()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	handler := GetTestingHandler()
	handler.ServeHTTP(w, req)
}
