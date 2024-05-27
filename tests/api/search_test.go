package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rejdeboer/multiplayer-server/internal/routes"
)

func TestSearchUsers(t *testing.T) {
	testApp := GetTestApp()
	testUser := testApp.user

	t.Run("success response", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodGet, "/user", nil)
		if err != nil {
			t.Fatal(err)
		}
		req.Header.Add("Authorization", "Bearer "+testApp.token)

		rr := httptest.NewRecorder()

		testApp.handler.ServeHTTP(rr, req)

		status := rr.Result().StatusCode
		if status != 200 {
			t.Errorf("expected %d got %d", 200, status)
		}

		var response []routes.UserListItem
		err = json.NewDecoder(rr.Body).Decode(&response)
		if err != nil {
			t.Errorf("error decoding json response: %v", err)
		}

		testUserFound := false
		for _, doc := range response {
			if doc.ID == testUser.ID {
				testUserFound = true
			}
		}

		if !testUserFound {
			t.Errorf("test document %v not present in response: %v", testUser.ID, response)
		}
	})
}
