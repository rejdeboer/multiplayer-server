package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rejdeboer/multiplayer-server/internal/routes"
)

func TestCreateUser(t *testing.T) {
	body := routes.UserCreate{
		Email:    "rick.deboer@live.nl",
		Password: "secret",
	}
	bodyBytes, err := json.Marshal(body)
	if err != nil {
		t.Fatal(err)
	}

	req, err := http.NewRequest(http.MethodPost, "/user", bytes.NewReader(bodyBytes))
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()

	handler := GetTestingHandler()
	handler.ServeHTTP(rr, req)

	if rr.Result().StatusCode != 200 {
		t.Errorf("expected 200 got %d", rr.Result().StatusCode)
	}
}
