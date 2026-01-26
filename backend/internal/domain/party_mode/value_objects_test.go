package party_mode

import (
	"strings"
	"testing"
)

// TestRoomID_Operations tests RoomID value object
func TestRoomID_Operations(t *testing.T) {
	id1 := NewRoomID()
	id2 := NewRoomID()

	if id1.IsZero() {
		t.Error("Generated ID should not be zero")
	}
	if id1.Equals(id2) {
		t.Error("Generated IDs should be unique")
	}

	id3 := NewRoomIDFromString("test-room-id")
	if id3.String() != "test-room-id" {
		t.Errorf("ID string = %s, want %s", id3.String(), "test-room-id")
	}

	zeroID := RoomID{}
	if !zeroID.IsZero() {
		t.Error("Empty ID should be zero")
	}
}

// TestGameID_Operations tests GameID value object
func TestGameID_Operations(t *testing.T) {
	id1 := NewGameID()
	id2 := NewGameID()

	if id1.IsZero() {
		t.Error("Generated ID should not be zero")
	}
	if id1.Equals(id2) {
		t.Error("Generated IDs should be unique")
	}

	id3 := NewGameIDFromString("test-game-id")
	if id3.String() != "test-game-id" {
		t.Errorf("ID string = %s, want %s", id3.String(), "test-game-id")
	}
}

// TestRoomCode_Generation tests room code generation
func TestRoomCode_Generation(t *testing.T) {
	code := GenerateRoomCode()

	if code.IsZero() {
		t.Error("Generated code should not be zero")
	}

	codeStr := code.String()
	if len(codeStr) != 7 { // ABC-123 format
		t.Errorf("Code length = %d, want 7 (ABC-123 format)", len(codeStr))
	}

	parts := strings.Split(codeStr, "-")
	if len(parts) != 2 {
		t.Errorf("Code should have 2 parts separated by dash, got %d", len(parts))
	}

	if len(parts[0]) != 3 {
		t.Errorf("Letters part length = %d, want 3", len(parts[0]))
	}

	if len(parts[1]) != 3 {
		t.Errorf("Digits part length = %d, want 3", len(parts[1]))
	}

	// Check letters are uppercase
	for _, c := range parts[0] {
		if c < 'A' || c > 'Z' {
			t.Errorf("Expected uppercase letter, got %c", c)
		}
	}

	// Check digits are numeric
	for _, c := range parts[1] {
		if c < '0' || c > '9' {
			t.Errorf("Expected digit, got %c", c)
		}
	}
}

// TestRoomCode_FromString tests room code normalization
func TestRoomCode_FromString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"Uppercase", "ABC-123", "ABC-123"},
		{"Lowercase (normalized)", "abc-123", "ABC-123"},
		{"With spaces", " abc-123 ", "ABC-123"},
		{"Mixed case", "AbC-123", "ABC-123"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code := NewRoomCodeFromString(tt.input)

			if code.String() != tt.expected {
				t.Errorf("Code = %s, want %s", code.String(), tt.expected)
			}
		})
	}
}

// TestRoomCode_Equals tests room code equality
func TestRoomCode_Equals(t *testing.T) {
	code1 := NewRoomCodeFromString("ABC-123")
	code2 := NewRoomCodeFromString("ABC-123")
	code3 := NewRoomCodeFromString("XYZ-999")

	if !code1.Equals(code2) {
		t.Error("Identical codes should be equal")
	}

	if code1.Equals(code3) {
		t.Error("Different codes should not be equal")
	}
}

// TestRoomSettings_NewRoomSettings tests default settings
func TestRoomSettings_NewRoomSettings(t *testing.T) {
	settings := NewRoomSettings()

	if settings.MaxPlayers() != 6 {
		t.Errorf("Default max players = %d, want 6", settings.MaxPlayers())
	}
	if settings.QuestionsCount() != 15 {
		t.Errorf("Default questions = %d, want 15", settings.QuestionsCount())
	}
	if settings.TimePerQuestion() != 15 {
		t.Errorf("Default time per question = %d, want 15", settings.TimePerQuestion())
	}
	if settings.Difficulty() != "mix" {
		t.Errorf("Default difficulty = %s, want mix", settings.Difficulty())
	}
	if !settings.ShowCorrectAnswer() {
		t.Error("Default should show correct answer")
	}
	if !settings.ShowPlayerAnswers() {
		t.Error("Default should show player answers")
	}
	if !settings.ShowCurrentScore() {
		t.Error("Default should show current score")
	}

	// Validate default settings
	if err := settings.Validate(); err != nil {
		t.Errorf("Default settings should be valid, got error: %v", err)
	}
}

// TestRoomSettings_Validate tests settings validation
func TestRoomSettings_Validate(t *testing.T) {
	tests := []struct {
		name        string
		settings    RoomSettings
		expectError bool
	}{
		{
			name:        "Valid settings",
			settings:    NewRoomSettings(),
			expectError: false,
		},
		{
			name: "Too few players",
			settings: RoomSettings{
				maxPlayers:      1,
				questionsCount:  15,
				timePerQuestion: 15,
			},
			expectError: true,
		},
		{
			name: "Too many players",
			settings: RoomSettings{
				maxPlayers:      9,
				questionsCount:  15,
				timePerQuestion: 15,
			},
			expectError: true,
		},
		{
			name: "Too few questions",
			settings: RoomSettings{
				maxPlayers:      4,
				questionsCount:  5,
				timePerQuestion: 15,
			},
			expectError: true,
		},
		{
			name: "Too many questions",
			settings: RoomSettings{
				maxPlayers:      4,
				questionsCount:  35,
				timePerQuestion: 15,
			},
			expectError: true,
		},
		{
			name: "Time too short",
			settings: RoomSettings{
				maxPlayers:      4,
				questionsCount:  15,
				timePerQuestion: 5,
			},
			expectError: true,
		},
		{
			name: "Time too long",
			settings: RoomSettings{
				maxPlayers:      4,
				questionsCount:  15,
				timePerQuestion: 35,
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.settings.Validate()

			if tt.expectError {
				if err == nil {
					t.Error("Expected validation error, got nil")
				}
				if err != ErrInvalidRoomSettings {
					t.Errorf("Expected ErrInvalidRoomSettings, got %v", err)
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, got %v", err)
				}
			}
		})
	}
}

// TestRoomStatus_CanTransitionTo tests room status transitions
func TestRoomStatus_CanTransitionTo(t *testing.T) {
	tests := []struct {
		name     string
		from     RoomStatus
		to       RoomStatus
		expected bool
	}{
		{"lobby -> playing", RoomStatusLobby, RoomStatusPlaying, true},
		{"lobby -> closed", RoomStatusLobby, RoomStatusClosed, true},
		{"playing -> closed", RoomStatusPlaying, RoomStatusClosed, true},
		{"playing -> lobby", RoomStatusPlaying, RoomStatusLobby, false},
		{"closed -> lobby", RoomStatusClosed, RoomStatusLobby, false},
		{"closed -> playing", RoomStatusClosed, RoomStatusPlaying, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.from.CanTransitionTo(tt.to)
			if result != tt.expected {
				t.Errorf("CanTransitionTo(%s -> %s) = %v, want %v",
					tt.from, tt.to, result, tt.expected)
			}
		})
	}
}

// TestRoomStatus_IsTerminal tests terminal state detection
func TestRoomStatus_IsTerminal(t *testing.T) {
	tests := []struct {
		status   RoomStatus
		expected bool
	}{
		{RoomStatusLobby, false},
		{RoomStatusPlaying, false},
		{RoomStatusClosed, true},
	}

	for _, tt := range tests {
		t.Run(string(tt.status), func(t *testing.T) {
			if tt.status.IsTerminal() != tt.expected {
				t.Errorf("IsTerminal(%s) = %v, want %v",
					tt.status, tt.status.IsTerminal(), tt.expected)
			}
		})
	}
}

// TestGameStatus_CanTransitionTo tests game status transitions
func TestGameStatus_CanTransitionTo(t *testing.T) {
	tests := []struct {
		name     string
		from     GameStatus
		to       GameStatus
		expected bool
	}{
		{"in_progress -> finished", GameStatusInProgress, GameStatusFinished, true},
		{"finished -> in_progress", GameStatusFinished, GameStatusInProgress, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.from.CanTransitionTo(tt.to)
			if result != tt.expected {
				t.Errorf("CanTransitionTo(%s -> %s) = %v, want %v",
					tt.from, tt.to, result, tt.expected)
			}
		})
	}
}

// TestGameStatus_IsTerminal tests game terminal state
func TestGameStatus_IsTerminal(t *testing.T) {
	tests := []struct {
		status   GameStatus
		expected bool
	}{
		{GameStatusInProgress, false},
		{GameStatusFinished, true},
	}

	for _, tt := range tests {
		t.Run(string(tt.status), func(t *testing.T) {
			if tt.status.IsTerminal() != tt.expected {
				t.Errorf("IsTerminal(%s) = %v, want %v",
					tt.status, tt.status.IsTerminal(), tt.expected)
			}
		})
	}
}
