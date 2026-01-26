package party_mode

import (
	"testing"

	"github.com/barsukov/quiz-sprint/backend/internal/domain/shared"
)

// TestNewPartyRoom_Success tests successful room creation
func TestNewPartyRoom_Success(t *testing.T) {
	hostID, _ := shared.NewUserID("host123")
	settings := NewRoomSettings()
	now := int64(1000000)

	room, err := NewPartyRoom(hostID, "HostPlayer", "Test Room", settings, now)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if room == nil {
		t.Fatal("Expected room to be created")
	}
	if room.ID().IsZero() {
		t.Error("Room ID should not be zero")
	}
	if room.Code().IsZero() {
		t.Error("Room code should not be zero")
	}
	if room.Status() != RoomStatusLobby {
		t.Errorf("Status = %v, want %v", room.Status(), RoomStatusLobby)
	}
	if room.PlayerCount() != 1 {
		t.Errorf("PlayerCount = %d, want 1 (host)", room.PlayerCount())
	}
	if !room.HostID().Equals(hostID) {
		t.Errorf("HostID = %v, want %v", room.HostID(), hostID)
	}

	// Check event emitted
	events := room.Events()
	if len(events) != 1 {
		t.Errorf("Expected 1 event, got %d", len(events))
	}
}

// TestNewPartyRoom_InvalidInputs tests validation
func TestNewPartyRoom_InvalidInputs(t *testing.T) {
	hostID, _ := shared.NewUserID("host123")
	settings := NewRoomSettings()
	now := int64(1000000)

	tests := []struct {
		name        string
		hostID      UserID
		settings    RoomSettings
		expectedErr error
	}{
		{
			name:        "Invalid host ID",
			hostID:      UserID{},
			settings:    settings,
			expectedErr: ErrInvalidRoomID,
		},
		{
			name:   "Invalid settings (too few players)",
			hostID: hostID,
			settings: RoomSettings{
				maxPlayers:      1,
				questionsCount:  15,
				timePerQuestion: 15,
			},
			expectedErr: ErrInvalidRoomSettings,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			room, err := NewPartyRoom(tt.hostID, "Test", "Room", tt.settings, now)

			if err == nil {
				t.Error("Expected error, got nil")
			}
			if err != tt.expectedErr {
				t.Errorf("Expected error %v, got %v", tt.expectedErr, err)
			}
			if room != nil {
				t.Error("Expected nil room for invalid input")
			}
		})
	}
}

// TestPartyRoom_JoinPlayer tests player joining
func TestPartyRoom_JoinPlayer(t *testing.T) {
	hostID, _ := shared.NewUserID("host123")
	settings := NewRoomSettings()
	now := int64(1000000)

	room, _ := NewPartyRoom(hostID, "Host", "Test Room", settings, now)
	room.Events() // Clear creation event

	// Join player
	player2ID, _ := shared.NewUserID("player2")
	err := room.JoinPlayer(player2ID, "Player2", now+1000)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if room.PlayerCount() != 2 {
		t.Errorf("PlayerCount = %d, want 2", room.PlayerCount())
	}

	// Check event
	events := room.Events()
	if len(events) != 1 {
		t.Errorf("Expected 1 event, got %d", len(events))
	}

	// Verify player in room
	players := room.Players()
	found := false
	for _, p := range players {
		if p.UserID().Equals(player2ID) {
			found = true
			if p.Username() != "Player2" {
				t.Errorf("Username = %s, want Player2", p.Username())
			}
			if p.IsHost() {
				t.Error("Player2 should not be host")
			}
		}
	}
	if !found {
		t.Error("Player2 not found in room")
	}
}

// TestPartyRoom_JoinPlayer_AlreadyInRoom tests duplicate join
func TestPartyRoom_JoinPlayer_AlreadyInRoom(t *testing.T) {
	hostID, _ := shared.NewUserID("host123")
	settings := NewRoomSettings()
	now := int64(1000000)

	room, _ := NewPartyRoom(hostID, "Host", "Test Room", settings, now)

	// Try to join host again
	err := room.JoinPlayer(hostID, "Host", now+1000)

	if err == nil {
		t.Error("Expected error when joining twice")
	}
	if err != ErrPlayerAlreadyInRoom {
		t.Errorf("Expected ErrPlayerAlreadyInRoom, got %v", err)
	}
}

// TestPartyRoom_JoinPlayer_RoomFull tests full room
func TestPartyRoom_JoinPlayer_RoomFull(t *testing.T) {
	hostID, _ := shared.NewUserID("host123")
	settings := RoomSettings{
		maxPlayers:      2, // Only 2 players allowed
		questionsCount:  15,
		timePerQuestion: 15,
	}
	now := int64(1000000)

	room, _ := NewPartyRoom(hostID, "Host", "Test Room", settings, now)

	// Join one player
	player2ID, _ := shared.NewUserID("player2")
	room.JoinPlayer(player2ID, "Player2", now+1000)

	// Try to join third player
	player3ID, _ := shared.NewUserID("player3")
	err := room.JoinPlayer(player3ID, "Player3", now+2000)

	if err == nil {
		t.Error("Expected error when room is full")
	}
	if err != ErrRoomFull {
		t.Errorf("Expected ErrRoomFull, got %v", err)
	}
}

// TestPartyRoom_RemovePlayer tests player leaving
func TestPartyRoom_RemovePlayer(t *testing.T) {
	hostID, _ := shared.NewUserID("host123")
	settings := NewRoomSettings()
	now := int64(1000000)

	room, _ := NewPartyRoom(hostID, "Host", "Test Room", settings, now)

	// Add player
	player2ID, _ := shared.NewUserID("player2")
	room.JoinPlayer(player2ID, "Player2", now+1000)
	room.Events() // Clear events

	// Remove player
	err := room.RemovePlayer(player2ID, now+2000)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if room.PlayerCount() != 1 {
		t.Errorf("PlayerCount = %d, want 1 (host remains)", room.PlayerCount())
	}

	// Check event
	events := room.Events()
	if len(events) != 1 {
		t.Errorf("Expected 1 event, got %d", len(events))
	}
}

// TestPartyRoom_RemovePlayer_HostTransfer tests host transfer
func TestPartyRoom_RemovePlayer_HostTransfer(t *testing.T) {
	hostID, _ := shared.NewUserID("host123")
	settings := NewRoomSettings()
	now := int64(1000000)

	room, _ := NewPartyRoom(hostID, "Host", "Test Room", settings, now)

	// Add player
	player2ID, _ := shared.NewUserID("player2")
	room.JoinPlayer(player2ID, "Player2", now+1000)
	room.Events() // Clear events

	// Remove host
	err := room.RemovePlayer(hostID, now+2000)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Player2 should become host
	if !room.HostID().Equals(player2ID) {
		t.Errorf("HostID = %v, want %v (Player2 should be new host)", room.HostID(), player2ID)
	}

	// Check events (left + host changed)
	events := room.Events()
	if len(events) < 2 {
		t.Errorf("Expected at least 2 events, got %d", len(events))
	}
}

// TestPartyRoom_SetPlayerReady tests ready status
func TestPartyRoom_SetPlayerReady(t *testing.T) {
	hostID, _ := shared.NewUserID("host123")
	settings := NewRoomSettings()
	now := int64(1000000)

	room, _ := NewPartyRoom(hostID, "Host", "Test Room", settings, now)

	// Add player
	player2ID, _ := shared.NewUserID("player2")
	room.JoinPlayer(player2ID, "Player2", now+1000)
	room.Events() // Clear events

	// Set player ready
	err := room.SetPlayerReady(player2ID, true, now+2000)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Check player ready status
	players := room.Players()
	for _, p := range players {
		if p.UserID().Equals(player2ID) {
			if !p.IsReady() {
				t.Error("Player should be ready")
			}
		}
	}

	// Check event
	events := room.Events()
	if len(events) != 1 {
		t.Errorf("Expected 1 event, got %d", len(events))
	}
}

// TestPartyRoom_CanStartGame tests start game validation
func TestPartyRoom_CanStartGame(t *testing.T) {
	hostID, _ := shared.NewUserID("host123")
	settings := NewRoomSettings()
	now := int64(1000000)

	room, _ := NewPartyRoom(hostID, "Host", "Test Room", settings, now)

	// Try to start with only 1 player
	err := room.CanStartGame(hostID)
	if err != ErrNotEnoughPlayers {
		t.Errorf("Expected ErrNotEnoughPlayers, got %v", err)
	}

	// Add player
	player2ID, _ := shared.NewUserID("player2")
	room.JoinPlayer(player2ID, "Player2", now+1000)

	// Try to start with not ready player
	err = room.CanStartGame(hostID)
	if err != ErrNotAllPlayersReady {
		t.Errorf("Expected ErrNotAllPlayersReady, got %v", err)
	}

	// Set player ready
	room.SetPlayerReady(player2ID, true, now+2000)

	// Now should be able to start
	err = room.CanStartGame(hostID)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Non-host cannot start
	err = room.CanStartGame(player2ID)
	if err != ErrOnlyHostCanStart {
		t.Errorf("Expected ErrOnlyHostCanStart, got %v", err)
	}
}

// TestPartyRoom_StartGame tests game start
func TestPartyRoom_StartGame(t *testing.T) {
	hostID, _ := shared.NewUserID("host123")
	settings := NewRoomSettings()
	now := int64(1000000)

	room, _ := NewPartyRoom(hostID, "Host", "Test Room", settings, now)

	// Start game
	err := room.StartGame()

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if room.Status() != RoomStatusPlaying {
		t.Errorf("Status = %v, want %v", room.Status(), RoomStatusPlaying)
	}
}

// TestPartyRoom_ReconstructPartyRoom tests reconstruction
func TestPartyRoom_ReconstructPartyRoom(t *testing.T) {
	hostID, _ := shared.NewUserID("host123")
	roomID := NewRoomID()
	code := NewRoomCodeFromString("ABC-123")
	settings := NewRoomSettings()
	now := int64(1000000)

	players := []RoomPlayer{
		NewRoomPlayer(hostID, "Host", true, now),
	}

	room := ReconstructPartyRoom(
		roomID,
		code,
		"Test Room",
		hostID,
		settings,
		players,
		RoomStatusLobby,
		now,
		now+3600,
	)

	if room == nil {
		t.Fatal("Expected room to be reconstructed")
	}
	if room.ID() != roomID {
		t.Errorf("ID = %v, want %v", room.ID(), roomID)
	}
	if room.Code() != code {
		t.Errorf("Code = %v, want %v", room.Code(), code)
	}

	// Events should be empty after reconstruction
	events := room.Events()
	if len(events) != 0 {
		t.Errorf("Reconstructed room should have no events, got %d", len(events))
	}
}
