package solo_marathon

import (
	"github.com/barsukov/quiz-sprint/backend/internal/domain/quiz"
	"github.com/barsukov/quiz-sprint/backend/internal/domain/shared"
	"github.com/google/uuid"
)

// Type aliases from other domains
type UserID = shared.UserID
type QuizID = quiz.QuizID
type QuestionID = quiz.QuestionID
type AnswerID = quiz.AnswerID
type Points = quiz.Points

// GameID uniquely identifies a marathon game
type GameID struct {
	value string
}

func NewGameID() GameID {
	return GameID{value: uuid.New().String()}
}

func NewGameIDFromString(value string) GameID {
	return GameID{value: value}
}

func (id GameID) String() string {
	return id.value
}

func (id GameID) IsZero() bool {
	return id.value == ""
}

func (id GameID) Equals(other GameID) bool {
	return id.value == other.value
}

// MarathonCategory represents the quiz category for marathon mode
type MarathonCategory struct {
	id   quiz.CategoryID
	name string
	isAllCategories bool // true for "all categories" mode
}

func NewMarathonCategoryAll() MarathonCategory {
	return MarathonCategory{
		id:              quiz.CategoryID{},
		name:            "all",
		isAllCategories: true,
	}
}

func NewMarathonCategory(categoryID quiz.CategoryID, name string) MarathonCategory {
	return MarathonCategory{
		id:              categoryID,
		name:            name,
		isAllCategories: false,
	}
}

func (mc MarathonCategory) CategoryID() quiz.CategoryID { return mc.id }
func (mc MarathonCategory) Name() string                { return mc.name }
func (mc MarathonCategory) IsAllCategories() bool       { return mc.isAllCategories }

// LivesSystem manages the lives mechanic (max 3 lives, 1 regenerates every 4 hours)
type LivesSystem struct {
	maxLives      int
	currentLives  int
	regenInterval int64 // 4 hours in seconds
	lastUpdate    int64 // Unix timestamp
}

const (
	MaxLives            = 3
	LifeRegenInterval   = 4 * 60 * 60 // 4 hours in seconds
)

func NewLivesSystem(now int64) LivesSystem {
	return LivesSystem{
		maxLives:      MaxLives,
		currentLives:  MaxLives,
		regenInterval: LifeRegenInterval,
		lastUpdate:    now,
	}
}

// ReconstructLivesSystem reconstructs LivesSystem from persistence
func ReconstructLivesSystem(currentLives int, lastUpdate int64) LivesSystem {
	return LivesSystem{
		maxLives:      MaxLives,
		currentLives:  currentLives,
		regenInterval: LifeRegenInterval,
		lastUpdate:    lastUpdate,
	}
}

// LoseLife decreases lives by 1 (immutable - returns new LivesSystem)
func (ls LivesSystem) LoseLife(now int64) LivesSystem {
	newLives := ls.currentLives - 1
	if newLives < 0 {
		newLives = 0
	}

	return LivesSystem{
		maxLives:      ls.maxLives,
		currentLives:  newLives,
		regenInterval: ls.regenInterval,
		lastUpdate:    now,
	}
}

// RegenerateLives calculates and returns updated LivesSystem based on time passed
func (ls LivesSystem) RegenerateLives(now int64) LivesSystem {
	if ls.currentLives >= ls.maxLives {
		return ls
	}

	timePassed := now - ls.lastUpdate
	livesRegened := int(timePassed / ls.regenInterval)

	if livesRegened == 0 {
		return ls
	}

	newLives := ls.currentLives + livesRegened
	if newLives > ls.maxLives {
		newLives = ls.maxLives
	}

	return LivesSystem{
		maxLives:      ls.maxLives,
		currentLives:  newLives,
		regenInterval: ls.regenInterval,
		lastUpdate:    now,
	}
}

// AddLives adds lives (from purchase, ad, etc.) - immutable
func (ls LivesSystem) AddLives(amount int, now int64) LivesSystem {
	newLives := ls.currentLives + amount
	if newLives > ls.maxLives {
		newLives = ls.maxLives
	}

	return LivesSystem{
		maxLives:      ls.maxLives,
		currentLives:  newLives,
		regenInterval: ls.regenInterval,
		lastUpdate:    now,
	}
}

// HasLives checks if player has at least one life
func (ls LivesSystem) HasLives() bool {
	return ls.currentLives > 0
}

// TimeToNextLife calculates seconds until next life regenerates
func (ls LivesSystem) TimeToNextLife(now int64) int64 {
	if ls.currentLives >= ls.maxLives {
		return 0
	}

	timeSinceLastUpdate := now - ls.lastUpdate
	timeUntilNext := ls.regenInterval - (timeSinceLastUpdate % ls.regenInterval)

	return timeUntilNext
}

// Getters
func (ls LivesSystem) CurrentLives() int { return ls.currentLives }
func (ls LivesSystem) MaxLives() int     { return ls.maxLives }
func (ls LivesSystem) LastUpdate() int64 { return ls.lastUpdate }

// HintType represents the type of hint
type HintType string

const (
	HintFiftyFifty HintType = "fifty_fifty" // Remove 2 incorrect answers
	HintExtraTime  HintType = "extra_time"  // Add 10 seconds to timer
	HintSkip       HintType = "skip"        // Skip question without losing life
)

// HintsSystem manages available hints
type HintsSystem struct {
	fiftyFifty int // Available 50/50 hints
	extraTime  int // Available +10 sec hints
	skip       int // Available skip hints
}

const (
	DefaultFiftyFifty = 3
	DefaultExtraTime  = 2
	DefaultSkip       = 1
)

func NewHintsSystem() HintsSystem {
	return HintsSystem{
		fiftyFifty: DefaultFiftyFifty,
		extraTime:  DefaultExtraTime,
		skip:       DefaultSkip,
	}
}

// ReconstructHintsSystem reconstructs HintsSystem from persistence
func ReconstructHintsSystem(fiftyFifty, extraTime, skip int) HintsSystem {
	return HintsSystem{
		fiftyFifty: fiftyFifty,
		extraTime:  extraTime,
		skip:       skip,
	}
}

// UseHint decreases hint count (immutable - returns new HintsSystem)
func (hs HintsSystem) UseHint(hintType HintType) (HintsSystem, error) {
	switch hintType {
	case HintFiftyFifty:
		if hs.fiftyFifty <= 0 {
			return hs, ErrNoHintsAvailable
		}
		return HintsSystem{
			fiftyFifty: hs.fiftyFifty - 1,
			extraTime:  hs.extraTime,
			skip:       hs.skip,
		}, nil

	case HintExtraTime:
		if hs.extraTime <= 0 {
			return hs, ErrNoHintsAvailable
		}
		return HintsSystem{
			fiftyFifty: hs.fiftyFifty,
			extraTime:  hs.extraTime - 1,
			skip:       hs.skip,
		}, nil

	case HintSkip:
		if hs.skip <= 0 {
			return hs, ErrNoHintsAvailable
		}
		return HintsSystem{
			fiftyFifty: hs.fiftyFifty,
			extraTime:  hs.extraTime,
			skip:       hs.skip - 1,
		}, nil

	default:
		return hs, ErrInvalidHintType
	}
}

// HasHint checks if specific hint is available
func (hs HintsSystem) HasHint(hintType HintType) bool {
	switch hintType {
	case HintFiftyFifty:
		return hs.fiftyFifty > 0
	case HintExtraTime:
		return hs.extraTime > 0
	case HintSkip:
		return hs.skip > 0
	default:
		return false
	}
}

// Getters
func (hs HintsSystem) FiftyFifty() int { return hs.fiftyFifty }
func (hs HintsSystem) ExtraTime() int  { return hs.extraTime }
func (hs HintsSystem) Skip() int       { return hs.skip }

// DifficultyLevel represents progression levels in marathon
type DifficultyLevel string

const (
	DifficultyBeginner DifficultyLevel = "beginner" // 1-5 questions
	DifficultyMedium   DifficultyLevel = "medium"   // 6-15 questions
	DifficultyHard     DifficultyLevel = "hard"     // 16-30 questions
	DifficultyExpert   DifficultyLevel = "expert"   // 31-50 questions
	DifficultyMaster   DifficultyLevel = "master"   // 51+ questions
)

// DifficultyProgression manages adaptive difficulty
type DifficultyProgression struct {
	level DifficultyLevel
}

func NewDifficultyProgression() DifficultyProgression {
	return DifficultyProgression{
		level: DifficultyBeginner,
	}
}

// UpdateFromStreak calculates difficulty level based on current streak
func (dp DifficultyProgression) UpdateFromStreak(streak int) DifficultyProgression {
	var level DifficultyLevel

	switch {
	case streak <= 5:
		level = DifficultyBeginner
	case streak <= 15:
		level = DifficultyMedium
	case streak <= 30:
		level = DifficultyHard
	case streak <= 50:
		level = DifficultyExpert
	default:
		level = DifficultyMaster
	}

	return DifficultyProgression{level: level}
}

// GetDistribution returns question difficulty distribution for current level
func (dp DifficultyProgression) GetDistribution() map[string]float64 {
	switch dp.level {
	case DifficultyBeginner:
		return map[string]float64{"easy": 0.8, "medium": 0.2, "hard": 0.0}
	case DifficultyMedium:
		return map[string]float64{"easy": 0.5, "medium": 0.4, "hard": 0.1}
	case DifficultyHard:
		return map[string]float64{"easy": 0.2, "medium": 0.5, "hard": 0.3}
	case DifficultyExpert:
		return map[string]float64{"easy": 0.1, "medium": 0.4, "hard": 0.5}
	case DifficultyMaster:
		return map[string]float64{"easy": 0.0, "medium": 0.3, "hard": 0.7}
	default:
		return map[string]float64{"easy": 0.8, "medium": 0.2, "hard": 0.0}
	}
}

// GetTimeLimit returns time limit in seconds for current level
func (dp DifficultyProgression) GetTimeLimit(streak int) int {
	switch {
	case streak <= 10:
		return 20 // seconds
	case streak <= 25:
		return 15
	case streak <= 50:
		return 12
	default:
		return 10
	}
}

func (dp DifficultyProgression) Level() DifficultyLevel {
	return dp.level
}

// GameStatus represents the current status of a marathon game
type GameStatus string

const (
	GameStatusInProgress GameStatus = "in_progress"
	GameStatusFinished   GameStatus = "finished"
	GameStatusAbandoned  GameStatus = "abandoned" // Player quit voluntarily
)

// State transition diagram for Marathon Game:
//
//   in_progress ──(no lives)──> finished (terminal)
//   in_progress ──(abandon)──> abandoned (terminal)
//
// Allowed transitions map
var marathonGameTransitions = map[GameStatus][]GameStatus{
	GameStatusInProgress: {GameStatusFinished, GameStatusAbandoned},
	GameStatusFinished:   {}, // Terminal state - no transitions allowed
	GameStatusAbandoned:  {}, // Terminal state - no transitions allowed
}

// CanTransitionTo checks if transition to target status is valid
func (gs GameStatus) CanTransitionTo(target GameStatus) bool {
	allowedTargets, exists := marathonGameTransitions[gs]
	if !exists {
		return false
	}

	for _, allowedTarget := range allowedTargets {
		if allowedTarget == target {
			return true
		}
	}

	return false
}

// IsTerminal checks if status is a terminal state (no further transitions)
func (gs GameStatus) IsTerminal() bool {
	transitions, exists := marathonGameTransitions[gs]
	return exists && len(transitions) == 0
}
