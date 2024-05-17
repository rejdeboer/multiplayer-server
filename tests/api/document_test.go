package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/rejdeboer/multiplayer-server/internal/routes"
)

func TestCreateDocument(t *testing.T) {
	testApp := GetTestApp()
	documentName := "some document"

	bodyBytes, err := json.Marshal(routes.DocumentCreate{
		Name: documentName,
	})
	if err != nil {
		t.Fatal(err)
	}

	req, err := http.NewRequest(http.MethodPost, "/document", bytes.NewReader(bodyBytes))
	req.Header.Add("Authorization", "Bearer "+testApp.token)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	testApp.handler.ServeHTTP(rr, req)

	status := rr.Result().StatusCode
	if status != 200 {
		t.Errorf("expected %d got %d", 200, rr.Result().StatusCode)
	}

	var response routes.DocumentResponse
	err = json.NewDecoder(rr.Body).Decode(&response)
	if err != nil {
		t.Errorf("error decoding json response: %v", err)
	}

	if response.OwnerID != testApp.user.ID {
		t.Errorf("output user_id mismatch; expected %v; got %v", testApp.user.ID, response.ID)
	}

	if response.Name != documentName {
		t.Errorf("output name mismatch; expected %v; got %v", documentName, response.Name)
	}
}

func TestDeleteDocument(t *testing.T) {
	testApp := GetTestApp()
	testDocID := testApp.document.ID

	t.Run("success response", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodDelete, "/document/"+testDocID.String(), nil)
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

	t.Run("document not found", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodDelete, "/document/"+uuid.New().String(), nil)
		if err != nil {
			t.Fatal(err)
		}
		req.Header.Add("Authorization", "Bearer "+testApp.token)

		rr := httptest.NewRecorder()
		testApp.handler.ServeHTTP(rr, req)

		status := rr.Result().StatusCode
		if status != 404 {
			t.Errorf("expected %d got %d", 404, rr.Result().StatusCode)
		}
	})

	t.Run("user is not owner", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodDelete, "/document/"+testDocID.String(), nil)
		if err != nil {
			t.Fatal(err)
		}
		otherUser := createTestUser()
		token, err := routes.GetJwt(
			settings.Application.SigningKey,
			settings.Application.TokenExpirationSeconds,
			otherUser.ID.String(),
			otherUser.Username,
		)
		if err != nil {
			t.Fatalf("error creating test token: %s", err)
		}
		req.Header.Add("Authorization", "Bearer "+token)

		rr := httptest.NewRecorder()
		testApp.handler.ServeHTTP(rr, req)

		status := rr.Result().StatusCode
		if status != 404 {
			t.Errorf("expected %d got %d", 404, rr.Result().StatusCode)
		}
	})
}

func TestListDocuments(t *testing.T) {
	testApp := GetTestApp()
	testDocID := testApp.document.ID

	req, err := http.NewRequest(http.MethodGet, "/document", nil)
	req.Header.Add("Authorization", "Bearer "+testApp.token)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	testApp.handler.ServeHTTP(rr, req)

	status := rr.Result().StatusCode
	if status != 200 {
		t.Errorf("expected %d got %d", 200, rr.Result().StatusCode)
	}

	var response []routes.DocumentListItem
	err = json.NewDecoder(rr.Body).Decode(&response)
	if err != nil {
		t.Errorf("error decoding json response: %v", err)
	}

	testDocumentFound := false
	for _, doc := range response {
		if doc.ID == testDocID {
			testDocumentFound = true
		}
	}

	if !testDocumentFound {
		t.Errorf("test document %v not present in response: %v", testDocID, response)
	}
}

func TestGetDocument(t *testing.T) {
	testApp := GetTestApp()
	testDocID := testApp.document.ID

	t.Run("success response", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodGet, "/document/"+testDocID.String(), nil)
		if err != nil {
			t.Fatal(err)
		}
		req.Header.Add("Authorization", "Bearer "+testApp.token)

		rr := httptest.NewRecorder()
		testApp.handler.ServeHTTP(rr, req)

		status := rr.Result().StatusCode
		if status != 200 {
			t.Errorf("expected %d got %d", 200, rr.Result().StatusCode)
		}

		var response routes.DocumentResponse
		err = json.NewDecoder(rr.Body).Decode(&response)
		if err != nil {
			t.Errorf("error decoding json response: %v", err)
		}

		if testDocID != response.ID {
			t.Errorf("documents do not match; expected: %v; was: %v", testApp.document, response)
		}
	})

	t.Run("unauthorized", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodGet, "/document/"+testDocID.String(), nil)
		if err != nil {
			t.Fatal(err)
		}

		rr := httptest.NewRecorder()
		testApp.handler.ServeHTTP(rr, req)

		status := rr.Result().StatusCode
		if status != 401 {
			t.Errorf("expected %d got %d", 401, rr.Result().StatusCode)
		}
	})
}
