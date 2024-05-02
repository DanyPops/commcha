package server

import (
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
  "context"

	"github.com/DanyPops/logues/domain"
	"github.com/gorilla/websocket"
)

func TestUserNoAuthClientSendMessage(t *testing.T) {
	t.Run("Send message via unautenticated user", func(t *testing.T) {
  })
}
func TestUserAuthClientSendMessage(t *testing.T) {
	t.Run("Send message via authenticated user", func(t *testing.T) {
    ctx := context.Background()
		c := domain.NewChannel()
    cc := domain.NewClientConfig().
      SetShortDeadline()

		go c.Start()
		l := NewServer(ctx, c, cc)
		server := httptest.NewServer(l)
		defer server.Close()

    authURL := server.URL + "/auth"
    req := httptest.NewRequest("POST", authURL, nil)
    resp := httptest.NewRecorder()
    l.userAuthenticator(resp, req)
    otp := struct{
      OTP string `json:"otp"`
    }{}
    _ = json.NewDecoder(resp.Body).Decode(&otp)

		wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/ws?otp=" + otp.OTP
		ws, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
		if err != nil {
			t.Fatalf("could not open a ws connection on %s: %v", wsURL, err)
		}
		defer ws.Close()

		msg := &domain.Message{
			Content: "Hello!",
		}
		if err := ws.WriteJSON(msg); err != nil {
			t.Fatalf("couldn't send JSON over WebSocket: %v", err)
		}

		time.Sleep(10 * time.Millisecond)

		got := &domain.Message{}
		if err := ws.ReadJSON(got); err != nil {
			t.Fatalf("couldn't read JSON over WebSocket: %v", err)
		}

		want := "dani: Hello!"
		if got.String() != want {
			t.Errorf("Got %s, want %s", got, want)
		}
	})
}
