package solo_marathon

import (
	"testing"

	"github.com/barsukov/quiz-sprint/backend/internal/domain/shared"
)

// TestNewPersonalBest_Success tests creating a new personal best
func TestNewPersonalBest_Success(t *testing.T) {
	playerID, _ := shared.NewUserID("user123")
	category := NewMarathonCategoryAll()
	now := int64(1000000)

	pb, err := NewPersonalBest(playerID, category, 10, 5000, now)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if pb == nil {
		t.Fatal("Expected personal best to be created")
	}
	if pb.ID().IsZero() {
		t.Error("Personal best ID should not be zero")
	}
	if pb.PlayerID() != playerID {
		t.Errorf("PlayerID = %v, want %v", pb.PlayerID(), playerID)
	}
	if pb.Category() != category {
		t.Errorf("Category = %v, want %v", pb.Category(), category)
	}
	if pb.BestStreak() != 10 {
		t.Errorf("BestStreak = %d, want %d", pb.BestStreak(), 10)
	}
	if pb.BestScore() != 5000 {
		t.Errorf("BestScore = %d, want %d", pb.BestScore(), 5000)
	}
	if pb.AchievedAt() != now {
		t.Errorf("AchievedAt = %d, want %d", pb.AchievedAt(), now)
	}
	if pb.UpdatedAt() != now {
		t.Errorf("UpdatedAt = %d, want %d", pb.UpdatedAt(), now)
	}
}

// TestNewPersonalBest_InvalidPlayerID tests validation
func TestNewPersonalBest_InvalidPlayerID(t *testing.T) {
	playerID := UserID{} // Zero ID
	category := NewMarathonCategoryAll()
	now := int64(1000000)

	pb, err := NewPersonalBest(playerID, category, 10, 5000, now)

	if err == nil {
		t.Error("Expected error for zero player ID")
	}
	if err != ErrInvalidPersonalBestID {
		t.Errorf("Expected ErrInvalidPersonalBestID, got %v", err)
	}
	if pb != nil {
		t.Error("Expected nil personal best for invalid input")
	}
}

// TestNewPersonalBest_NegativeValues tests negative streak/score handling
func TestNewPersonalBest_NegativeValues(t *testing.T) {
	playerID, _ := shared.NewUserID("user123")
	category := NewMarathonCategoryAll()
	now := int64(1000000)

	tests := []struct {
		name           string
		streak         int
		score          int
		expectedStreak int
		expectedScore  int
	}{
		{"Negative streak", -5, 1000, 0, 1000},
		{"Negative score", 10, -500, 10, 0},
		{"Both negative", -3, -200, 0, 0},
		{"Valid values", 15, 7500, 15, 7500},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pb, err := NewPersonalBest(playerID, category, tt.streak, tt.score, now)

			if err != nil {
				t.Fatalf("Expected no error, got %v", err)
			}
			if pb.BestStreak() != tt.expectedStreak {
				t.Errorf("BestStreak = %d, want %d", pb.BestStreak(), tt.expectedStreak)
			}
			if pb.BestScore() != tt.expectedScore {
				t.Errorf("BestScore = %d, want %d", pb.BestScore(), tt.expectedScore)
			}
		})
	}
}

// TestPersonalBest_UpdateIfBetter tests record update logic
func TestPersonalBest_UpdateIfBetter(t *testing.T) {
	playerID, _ := shared.NewUserID("user123")
	category := NewMarathonCategoryAll()
	now := int64(1000000)

	tests := []struct {
		name           string
		initialStreak  int
		initialScore   int
		newStreak      int
		newScore       int
		expectedUpdate bool
		expectedStreak int
		expectedScore  int
	}{
		{
			name:           "Higher streak",
			initialStreak:  10,
			initialScore:   5000,
			newStreak:      15,
			newScore:       4000,
			expectedUpdate: true,
			expectedStreak: 15,
			expectedScore:  4000,
		},
		{
			name:           "Same streak, higher score",
			initialStreak:  10,
			initialScore:   5000,
			newStreak:      10,
			newScore:       6000,
			expectedUpdate: true,
			expectedStreak: 10,
			expectedScore:  6000,
		},
		{
			name:           "Lower streak",
			initialStreak:  10,
			initialScore:   5000,
			newStreak:      8,
			newScore:       7000,
			expectedUpdate: false,
			expectedStreak: 10,
			expectedScore:  5000,
		},
		{
			name:           "Same streak, lower score",
			initialStreak:  10,
			initialScore:   5000,
			newStreak:      10,
			newScore:       4000,
			expectedUpdate: false,
			expectedStreak: 10,
			expectedScore:  5000,
		},
		{
			name:           "Same streak and score",
			initialStreak:  10,
			initialScore:   5000,
			newStreak:      10,
			newScore:       5000,
			expectedUpdate: false,
			expectedStreak: 10,
			expectedScore:  5000,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pb, _ := NewPersonalBest(playerID, category, tt.initialStreak, tt.initialScore, now)

			updated := pb.UpdateIfBetter(tt.newStreak, tt.newScore, now+1000)

			if updated != tt.expectedUpdate {
				t.Errorf("UpdateIfBetter() = %v, want %v", updated, tt.expectedUpdate)
			}
			if pb.BestStreak() != tt.expectedStreak {
				t.Errorf("BestStreak = %d, want %d", pb.BestStreak(), tt.expectedStreak)
			}
			if pb.BestScore() != tt.expectedScore {
				t.Errorf("BestScore = %d, want %d", pb.BestScore(), tt.expectedScore)
			}

			// If updated, check timestamps changed
			if tt.expectedUpdate {
				if pb.AchievedAt() != now+1000 {
					t.Errorf("AchievedAt should be updated to %d, got %d", now+1000, pb.AchievedAt())
				}
				if pb.UpdatedAt() != now+1000 {
					t.Errorf("UpdatedAt should be updated to %d, got %d", now+1000, pb.UpdatedAt())
				}
			} else {
				// If not updated, timestamps should be original
				if pb.AchievedAt() != now {
					t.Errorf("AchievedAt should remain %d, got %d", now, pb.AchievedAt())
				}
				if pb.UpdatedAt() != now {
					t.Errorf("UpdatedAt should remain %d, got %d", now, pb.UpdatedAt())
				}
			}
		})
	}
}

// TestPersonalBest_IsBetter tests comparison without mutation
func TestPersonalBest_IsBetter(t *testing.T) {
	playerID, _ := shared.NewUserID("user123")
	category := NewMarathonCategoryAll()
	now := int64(1000000)

	pb, _ := NewPersonalBest(playerID, category, 10, 5000, now)

	tests := []struct {
		name     string
		streak   int
		score    int
		expected bool
	}{
		{"Higher streak", 15, 4000, true},
		{"Same streak, higher score", 10, 6000, true},
		{"Lower streak", 8, 7000, false},
		{"Same streak, lower score", 10, 4000, false},
		{"Same streak and score", 10, 5000, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := pb.IsBetter(tt.streak, tt.score)

			if result != tt.expected {
				t.Errorf("IsBetter(%d, %d) = %v, want %v", tt.streak, tt.score, result, tt.expected)
			}

			// Verify no mutation (record unchanged)
			if pb.BestStreak() != 10 {
				t.Error("IsBetter should not mutate BestStreak")
			}
			if pb.BestScore() != 5000 {
				t.Error("IsBetter should not mutate BestScore")
			}
		})
	}
}

// TestReconstructPersonalBest tests reconstruction from persistence
func TestReconstructPersonalBest(t *testing.T) {
	playerID, _ := shared.NewUserID("user123")
	category := NewMarathonCategoryAll()
	id := NewPersonalBestID()
	now := int64(1000000)

	pb := ReconstructPersonalBest(
		id,
		playerID,
		category,
		25,
		12000,
		now,
		now+5000,
	)

	if pb == nil {
		t.Fatal("Expected personal best to be reconstructed")
	}
	if pb.ID() != id {
		t.Errorf("ID = %v, want %v", pb.ID(), id)
	}
	if pb.PlayerID() != playerID {
		t.Errorf("PlayerID = %v, want %v", pb.PlayerID(), playerID)
	}
	if pb.Category() != category {
		t.Errorf("Category = %v, want %v", pb.Category(), category)
	}
	if pb.BestStreak() != 25 {
		t.Errorf("BestStreak = %d, want %d", pb.BestStreak(), 25)
	}
	if pb.BestScore() != 12000 {
		t.Errorf("BestScore = %d, want %d", pb.BestScore(), 12000)
	}
	if pb.AchievedAt() != now {
		t.Errorf("AchievedAt = %d, want %d", pb.AchievedAt(), now)
	}
	if pb.UpdatedAt() != now+5000 {
		t.Errorf("UpdatedAt = %d, want %d", pb.UpdatedAt(), now+5000)
	}
}

// TestPersonalBestID_Operations tests ID value object operations
func TestPersonalBestID_Operations(t *testing.T) {
	// Test generation
	id1 := NewPersonalBestID()
	id2 := NewPersonalBestID()

	if id1.IsZero() {
		t.Error("Generated ID should not be zero")
	}
	if id1.String() == "" {
		t.Error("Generated ID should have string representation")
	}
	if id1.Equals(id2) {
		t.Error("Generated IDs should be unique")
	}

	// Test from string
	id3 := NewPersonalBestIDFromString("test-id-123")
	if id3.String() != "test-id-123" {
		t.Errorf("ID string = %s, want %s", id3.String(), "test-id-123")
	}

	// Test zero ID
	zeroID := PersonalBestID{}
	if !zeroID.IsZero() {
		t.Error("Empty ID should be zero")
	}

	// Test equals
	id4 := NewPersonalBestIDFromString("same-id")
	id5 := NewPersonalBestIDFromString("same-id")
	if !id4.Equals(id5) {
		t.Error("IDs with same string should be equal")
	}
}
