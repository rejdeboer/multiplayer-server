package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dgrijalva/jwt-go"
	"github.com/rejdeboer/multiplayer-server/internal/routes"
	"github.com/rejdeboer/multiplayer-server/pkg/httperrors"
)

func TestGetToken(t *testing.T) {
	testApp := GetTestApp()
	testUser := testApp.user

	cases := []struct {
		name                 string
		input                routes.UserCredentials
		outputStatusCode     int
		expectedErrorMessage string
	}{
		{
			name:                 "wrong email",
			outputStatusCode:     401,
			expectedErrorMessage: "invalid email or password",
			input: routes.UserCredentials{
				Email:    testUser.Email + "x",
				Password: testUser.Password,
			},
		},
		{
			name:                 "wrong password",
			outputStatusCode:     401,
			expectedErrorMessage: "invalid email or password",
			input: routes.UserCredentials{
				Email:    testUser.Email,
				Password: testUser.Password + "x",
			},
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
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

			var response httperrors.Response
			err = json.NewDecoder(rr.Body).Decode(&response)
			if err != nil {
				t.Errorf("error decoding response: %s", err)
			}

			if response.Message != testCase.expectedErrorMessage {
				t.Errorf("expected error message: %s; got: %s", testCase.expectedErrorMessage, response.Message)
			}
		})
	}

	t.Run("success response", func(t *testing.T) {
		input := routes.UserCredentials{
			Email:    testUser.Email,
			Password: testUser.Password,
		}

		bodyBytes, err := json.Marshal(input)
		if err != nil {
			t.Fatal(err)
		}

		req, err := http.NewRequest(http.MethodPost, "/token", bytes.NewReader(bodyBytes))
		if err != nil {
			t.Fatal(err)
		}

		rr := httptest.NewRecorder()

		testApp.handler.ServeHTTP(rr, req)

		var response routes.TokenResponse
		err = json.NewDecoder(rr.Body).Decode(&response)
		if err != nil {
			t.Errorf("error decoding response: %s", err)
		}

		claims := jwt.MapClaims{}
		_, err = jwt.ParseWithClaims(response.Token, claims, func(t *jwt.Token) (interface{}, error) {
			return []byte(testApp.settings.SigningKey), nil
		})
		if err != nil {
			t.Errorf("error decoding jwt: %s", err)
		}

		if claims["username"] != testUser.Username {
			t.Errorf("expected username: %s; got: %s", testUser.Username, claims["username"])
		}
	})
}
