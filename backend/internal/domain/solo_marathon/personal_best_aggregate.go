package solo_marathon

import "github.com/google/uuid"

// PersonalBestID uniquely identifies a personal best record
type PersonalBestID struct {
	value string
}

func NewPersonalBestID() PersonalBestID {
	return PersonalBestID{value: uuid.New().String()}
}

func NewPersonalBestIDFromString(value string) PersonalBestID {
	return PersonalBestID{value: value}
}

func (id PersonalBestID) String() string {
	return id.value
}

func (id PersonalBestID) IsZero() bool {
	return id.value == ""
}

func (id PersonalBestID) Equals(other PersonalBestID) bool {
	return id.value == other.value
}

// PersonalBest is the aggregate root for player's best marathon record per category
type PersonalBest struct {
	id         PersonalBestID
	playerID   UserID
	category   MarathonCategory
	bestStreak int   // Best streak achieved
	bestScore  int   // Best score achieved
	achievedAt int64 // Unix timestamp when record was set
	updatedAt  int64 // Unix timestamp of last update
}

// NewPersonalBest creates a new personal best record
func NewPersonalBest(
	playerID UserID,
	category MarathonCategory,
	streak int,
	score int,
	achievedAt int64,
) (*PersonalBest, error) {
	if playerID.IsZero() {
		return nil, ErrInvalidPersonalBestID
	}

	if streak < 0 {
		streak = 0
	}

	if score < 0 {
		score = 0
	}

	return &PersonalBest{
		id:         NewPersonalBestID(),
		playerID:   playerID,
		category:   category,
		bestStreak: streak,
		bestScore:  score,
		achievedAt: achievedAt,
		updatedAt:  achievedAt,
	}, nil
}

// UpdateIfBetter updates the record if new streak/score is better
// Returns true if record was updated
func (pb *PersonalBest) UpdateIfBetter(newStreak int, newScore int, achievedAt int64) bool {
	// Check if new streak is better
	if newStreak > pb.bestStreak {
		pb.bestStreak = newStreak
		pb.bestScore = newScore
		pb.achievedAt = achievedAt
		pb.updatedAt = achievedAt
		return true
	}

	// If streak is same but score is better
	if newStreak == pb.bestStreak && newScore > pb.bestScore {
		pb.bestScore = newScore
		pb.achievedAt = achievedAt
		pb.updatedAt = achievedAt
		return true
	}

	return false
}

// IsBetter checks if given streak/score is better than current record
func (pb *PersonalBest) IsBetter(streak int, score int) bool {
	if streak > pb.bestStreak {
		return true
	}
	if streak == pb.bestStreak && score > pb.bestScore {
		return true
	}
	return false
}

// Getters
func (pb *PersonalBest) ID() PersonalBestID          { return pb.id }
func (pb *PersonalBest) PlayerID() UserID            { return pb.playerID }
func (pb *PersonalBest) Category() MarathonCategory  { return pb.category }
func (pb *PersonalBest) BestStreak() int             { return pb.bestStreak }
func (pb *PersonalBest) BestScore() int              { return pb.bestScore }
func (pb *PersonalBest) AchievedAt() int64           { return pb.achievedAt }
func (pb *PersonalBest) UpdatedAt() int64            { return pb.updatedAt }

// ReconstructPersonalBest reconstructs a PersonalBest from persistence
// Used by repository when loading from database
func ReconstructPersonalBest(
	id PersonalBestID,
	playerID UserID,
	category MarathonCategory,
	bestStreak int,
	bestScore int,
	achievedAt int64,
	updatedAt int64,
) *PersonalBest {
	return &PersonalBest{
		id:         id,
		playerID:   playerID,
		category:   category,
		bestStreak: bestStreak,
		bestScore:  bestScore,
		achievedAt: achievedAt,
		updatedAt:  updatedAt,
	}
}
