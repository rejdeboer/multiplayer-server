package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

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
