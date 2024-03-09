package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rejdeboer/multiplayer-server/internal/routes"
)

func TestGetToken(t *testing.T) {
	testApp := GetTestApp()
	testUser := testApp.user

	cases := []struct {
		input            routes.UserCredentials
		outputStatusCode int
	}{
		{
			outputStatusCode: 200,
			input: routes.UserCredentials{
				Email:    testUser.Email,
				Password: testUser.Password,
			},
		},
		{
			outputStatusCode: 401,
			input: routes.UserCredentials{
				Email:    testUser.Email + "x",
				Password: testUser.Password,
			},
		},
		{
			outputStatusCode: 401,
			input: routes.UserCredentials{
				Email:    testUser.Email,
				Password: testUser.Password + "x",
			},
		},
	}

	for _, testCase := range cases {
		bodyBytes, err := json.Marshal(testCase.input)
		if err != nil {
			t.Fatal(err)
		}

		req, err := http.NewRequest(http.MethodPost, "/token", bytes.NewReader(bodyBytes))
		if err != nil {
			t.Fatal(err)
		}

		rr := httptest.NewRecorder()

		testApp.handler.ServeHTTP(rr, req)

		if rr.Result().StatusCode != testCase.outputStatusCode {
			t.Errorf("expected %d got %d", testCase.outputStatusCode, rr.Result().StatusCode)
		}
	}
}
