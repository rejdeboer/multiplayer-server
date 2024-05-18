package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rejdeboer/multiplayer-server/internal/routes"
)

func TestCreateContributor(t *testing.T) {
	testApp := GetTestApp()
	testDocID := testApp.document.ID

	otherUser := createTestUser()
	bodyBytes, err := json.Marshal(routes.DocumentContributorCreate{
		UserID: otherUser.ID,
	})
	if err != nil {
		t.Fatal(err)
	}

	t.Run("success response", func(t *testing.T) {
		req, err := http.NewRequest(
			http.MethodPost,
			"/document/"+testDocID.String()+"/contributor",
			bytes.NewReader(bodyBytes),
		)
		if err != nil {
			t.Fatal(err)
		}
		req.Header.Add("Authorization", "Bearer "+testApp.token)

		rr := httptest.NewRecorder()
		testApp.handler.ServeHTTP(rr, req)

		status := rr.Result().StatusCode
		if status != 202 {
			t.Errorf("expected %d got %d", 202, rr.Result().StatusCode)
		}
	})
}
