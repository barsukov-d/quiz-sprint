package handlers_test

import (
	"sync"
	"testing"

	appDuel "github.com/barsukov/quiz-sprint/backend/internal/application/quick_duel"
	"github.com/barsukov/quiz-sprint/backend/internal/infrastructure/http/handlers"
)

// mockLobbyConn records WriteJSON calls for assertions.
type mockLobbyConn struct {
	mu       sync.Mutex
	messages []interface{}
}

func (m *mockLobbyConn) WriteJSON(v interface{}) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.messages = append(m.messages, v)
	return nil
}

func TestDuelLobbyHub_IsConnected(t *testing.T) {
	hub := handlers.NewDuelLobbyHub(nil)

	if hub.IsConnected("p1") {
		t.Fatal("expected not connected before register")
	}

	conn := &mockLobbyConn{}
	hub.Register("p1", conn)

	if !hub.IsConnected("p1") {
		t.Fatal("expected connected after register")
	}

	hub.Unregister("p1")

	if hub.IsConnected("p1") {
		t.Fatal("expected not connected after unregister")
	}
}

func TestDuelLobbyHub_Notify(t *testing.T) {
	hub := handlers.NewDuelLobbyHub(nil)
	conn := &mockLobbyConn{}
	hub.Register("p1", conn)

	event := appDuel.LobbyEvent{Type: "challenge_accepted", Data: map[string]string{"challengeId": "abc"}}
	hub.Notify("p1", event)

	conn.mu.Lock()
	defer conn.mu.Unlock()
	if len(conn.messages) != 1 {
		t.Fatalf("expected 1 message, got %d", len(conn.messages))
	}
}

func TestDuelLobbyHub_Notify_NotConnected(t *testing.T) {
	hub := handlers.NewDuelLobbyHub(nil)
	// Should not panic when player not connected
	hub.Notify("nonexistent", appDuel.LobbyEvent{Type: "test"})
}

func TestDuelLobbyHub_NotifyBoth(t *testing.T) {
	hub := handlers.NewDuelLobbyHub(nil)
	c1, c2 := &mockLobbyConn{}, &mockLobbyConn{}
	hub.Register("p1", c1)
	hub.Register("p2", c2)

	hub.NotifyBoth("p1", "p2", appDuel.LobbyEvent{Type: "game_ready"})

	c1.mu.Lock()
	defer c1.mu.Unlock()
	c2.mu.Lock()
	defer c2.mu.Unlock()
	if len(c1.messages) != 1 || len(c2.messages) != 1 {
		t.Fatalf("expected 1 message each, got %d and %d", len(c1.messages), len(c2.messages))
	}
}
