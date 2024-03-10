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
		outputBody       interface{}
	}{
		{
			outputStatusCode: 200,
			outputBody: routes.UserResponse{
				ID:       "",
				Email:    "rick.deboer@live.nl",
				Username: "rejdeboer",
			},
			input: routes.UserCreate{
				Email:    "rick.deboer@live.nl",
				Username: "rejdeboer",
				Password: "Very$ecret1",
			},
		},
		{
			outputStatusCode: 400,
			outputBody: routes.ErrorResponse{
				Message: "invalid email address",
				Status:  400,
			},
			input: routes.UserCreate{
				Email:    "",
				Username: "rejdeboer",
				Password: "Very$ecret1",
			},
		},
		{
			outputStatusCode: 400,
			outputBody: routes.ErrorResponse{
				Message: "invalid email address",
				Status:  400,
			},
			input: routes.UserCreate{
				Email:    "rick.deboerlive.nl",
				Username: "rejdeboer",
				Password: "Very$ecret1",
			},
		},
		{
			outputStatusCode: 400,
			outputBody: routes.ErrorResponse{
				Message: "password must contain at least one digit",
				Status:  400,
			},
			input: routes.UserCreate{
				Email:    "rick.deboer@live.nl",
				Username: "rejdeboer",
				Password: "Very$ecret",
			},
		},
		{
			outputStatusCode: 400,
			outputBody: routes.ErrorResponse{
				Message: "password must contain at least one special character",
				Status:  400,
			},
			input: routes.UserCreate{
				Email:    "rick.deboer@live.nl",
				Password: "Verysecret1",
			},
		},
	}

	testApp := GetTestApp()
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

		testApp.handler.ServeHTTP(rr, req)

		status := rr.Result().StatusCode
		if status != testCase.outputStatusCode {
			t.Errorf("expected %d got %d", testCase.outputStatusCode, rr.Result().StatusCode)
		}

		if status != 200 {
			var response routes.ErrorResponse
			err = json.NewDecoder(rr.Body).Decode(&response)
			if err != nil {
				t.Errorf("error decoding json response: %v", err)
			}
			if response != testCase.outputBody {
				t.Errorf("output body mismatch; expected %v; got %v", testCase.outputBody, response)
			}
			return
		}

		var response routes.UserResponse
		err = json.NewDecoder(rr.Body).Decode(&response)
		if err != nil {
			t.Errorf("error decoding json response: %v", err)
		}

		response.ID = ""
		if response != testCase.outputBody {
			t.Errorf("output body mismatch; expected %v; got %v", testCase.outputBody, response)
		}
	}
}
