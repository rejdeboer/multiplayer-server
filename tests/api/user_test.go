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
	cases := []struct {
		input            routes.UserCreate
		outputStatusCode int
	}{
		{
			outputStatusCode: 200,
			input: routes.UserCreate{
				Email:    "rick.deboer@live.nl",
				Password: "Very$ecret1",
			},
		},
		{
			outputStatusCode: 400,
			input: routes.UserCreate{
				Email:    "",
				Password: "Very$ecret1",
			},
		},
		{
			outputStatusCode: 400,
			input: routes.UserCreate{
				Email:    "rick.deboerlive.nl",
				Password: "Very$ecret1",
			},
		},
		{
			outputStatusCode: 400,
			input: routes.UserCreate{
				Email:    "rick.deboer@live.nl",
				Password: "Very$ecret",
			},
		},
		{
			outputStatusCode: 400,
			input: routes.UserCreate{
				Email:    "rick.deboer@live.nl",
				Password: "Verysecret1",
			},
		},
	}

	handler := GetTestingHandler()
	for _, testCase := range cases {
		bodyBytes, err := json.Marshal(testCase.input)
		if err != nil {
			t.Fatal(err)
		}

		req, err := http.NewRequest(http.MethodPost, "/user", bytes.NewReader(bodyBytes))
		if err != nil {
			t.Fatal(err)
		}

		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		if rr.Result().StatusCode != testCase.outputStatusCode {
			t.Errorf("expected %d got %d", testCase.outputStatusCode, rr.Result().StatusCode)
		}
	}
}
