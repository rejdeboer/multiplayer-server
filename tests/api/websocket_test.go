package api

import (
	"github.com/gorilla/websocket"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestWebsocket(t *testing.T) {
	testApp := GetTestApp()

	s := httptest.NewServer(testApp.handler)
	defer s.Close()

	u := "ws" + strings.TrimPrefix(s.URL, "http") + "/websocket"
	header := http.Header{
		"Authorization": {"Bearer " + testApp.token},
	}

	ws, _, err := websocket.DefaultDialer.Dial(u, header)
	if err != nil {
		t.Fatalf("error connecting to server: %v", err)
	}
	defer ws.Close()

	for i := 0; i < 10; i++ {
		if err := ws.WriteMessage(websocket.TextMessage, []byte("hello")); err != nil {
			t.Fatalf("error writing message: %v", err)
		}
	}
}
