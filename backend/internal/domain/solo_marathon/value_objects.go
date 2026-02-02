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

// ResetForContinue resets lives to 1 for continue mechanic (immutable)
func (ls LivesSystem) ResetForContinue(now int64) LivesSystem {
	return LivesSystem{
		maxLives:      ls.maxLives,
		currentLives:  1, // Continue always gives exactly 1 life
		regenInterval: ls.regenInterval,
		lastUpdate:    now,
	}
}

// Label returns visual representation of lives (e.g., "❤️❤️🖤")
func (ls LivesSystem) Label() string {
	hearts := ""
	for i := 0; i < ls.currentLives; i++ {
		hearts += "❤️"
	}
	for i := ls.currentLives; i < ls.maxLives; i++ {
		hearts += "🖤"
	}
	return hearts
}

// Getters
func (ls LivesSystem) CurrentLives() int { return ls.currentLives }
func (ls LivesSystem) MaxLives() int     { return ls.maxLives }
func (ls LivesSystem) LastUpdate() int64 { return ls.lastUpdate }

// BonusType represents the type of bonus
type BonusType string

const (
	BonusShield     BonusType = "shield"      // Protect from one wrong answer (activated before answering)
	BonusFiftyFifty BonusType = "fifty_fifty"  // Remove 2 incorrect answers
	BonusSkip       BonusType = "skip"         // Skip question without losing life
	BonusFreeze     BonusType = "freeze"       // Add 10 seconds to timer (stackable)
)

// BonusInventory manages available bonuses
type BonusInventory struct {
	shield     int // Available shield bonuses
	fiftyFifty int // Available 50/50 bonuses
	skip       int // Available skip bonuses
	freeze     int // Available freeze bonuses (+10 sec)
}

const (
	DefaultShield     = 2
	DefaultFiftyFifty = 1
	DefaultSkip       = 0
	DefaultFreeze     = 3
)

func NewBonusInventory() BonusInventory {
	return BonusInventory{
		shield:     DefaultShield,
		fiftyFifty: DefaultFiftyFifty,
		skip:       DefaultSkip,
		freeze:     DefaultFreeze,
	}
}

// NewBonusInventoryFromSelected creates BonusInventory from player-selected bonuses
func NewBonusInventoryFromSelected(shield, fiftyFifty, skip, freeze int) BonusInventory {
	return BonusInventory{
		shield:     max(0, shield),
		fiftyFifty: max(0, fiftyFifty),
		skip:       max(0, skip),
		freeze:     max(0, freeze),
	}
}

// ReconstructBonusInventory reconstructs BonusInventory from persistence
func ReconstructBonusInventory(shield, fiftyFifty, skip, freeze int) BonusInventory {
	return BonusInventory{
		shield:     shield,
		fiftyFifty: fiftyFifty,
		skip:       skip,
		freeze:     freeze,
	}
}

// UseBonus decreases bonus count (immutable - returns new BonusInventory)
func (bi BonusInventory) UseBonus(bonusType BonusType) (BonusInventory, error) {
	switch bonusType {
	case BonusShield:
		if bi.shield <= 0 {
			return bi, ErrNoBonusesAvailable
		}
		return BonusInventory{
			shield:     bi.shield - 1,
			fiftyFifty: bi.fiftyFifty,
			skip:       bi.skip,
			freeze:     bi.freeze,
		}, nil

	case BonusFiftyFifty:
		if bi.fiftyFifty <= 0 {
			return bi, ErrNoBonusesAvailable
		}
		return BonusInventory{
			shield:     bi.shield,
			fiftyFifty: bi.fiftyFifty - 1,
			skip:       bi.skip,
			freeze:     bi.freeze,
		}, nil

	case BonusSkip:
		if bi.skip <= 0 {
			return bi, ErrNoBonusesAvailable
		}
		return BonusInventory{
			shield:     bi.shield,
			fiftyFifty: bi.fiftyFifty,
			skip:       bi.skip - 1,
			freeze:     bi.freeze,
		}, nil

	case BonusFreeze:
		if bi.freeze <= 0 {
			return bi, ErrNoBonusesAvailable
		}
		return BonusInventory{
			shield:     bi.shield,
			fiftyFifty: bi.fiftyFifty,
			skip:       bi.skip,
			freeze:     bi.freeze - 1,
		}, nil

	default:
		return bi, ErrInvalidBonusType
	}
}

// HasBonus checks if specific bonus is available
func (bi BonusInventory) HasBonus(bonusType BonusType) bool {
	switch bonusType {
	case BonusShield:
		return bi.shield > 0
	case BonusFiftyFifty:
		return bi.fiftyFifty > 0
	case BonusSkip:
		return bi.skip > 0
	case BonusFreeze:
		return bi.freeze > 0
	default:
		return false
	}
}

// Count returns quantity of a specific bonus type
func (bi BonusInventory) Count(bonusType BonusType) int {
	switch bonusType {
	case BonusShield:
		return bi.shield
	case BonusFiftyFifty:
		return bi.fiftyFifty
	case BonusSkip:
		return bi.skip
	case BonusFreeze:
		return bi.freeze
	default:
		return 0
	}
}

// Getters
func (bi BonusInventory) Shield() int     { return bi.shield }
func (bi BonusInventory) FiftyFifty() int { return bi.fiftyFifty }
func (bi BonusInventory) Skip() int       { return bi.skip }
func (bi BonusInventory) Freeze() int     { return bi.freeze }

// DifficultyLevel represents progression levels in marathon
type DifficultyLevel string

const (
	DifficultyBeginner DifficultyLevel = "beginner" // Questions 1-10
	DifficultyMedium   DifficultyLevel = "medium"   // Questions 11-30
	DifficultyHard     DifficultyLevel = "hard"     // Questions 31-50
	DifficultyMaster   DifficultyLevel = "master"   // Questions 51+
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

// UpdateFromQuestionIndex calculates difficulty level based on question index (1-based)
// Per docs: 1-10: easy/medium, 11-30: medium, 31-50: medium/hard, 51+: hard
func (dp DifficultyProgression) UpdateFromQuestionIndex(questionIndex int) DifficultyProgression {
	var level DifficultyLevel

	switch {
	case questionIndex <= 10:
		level = DifficultyBeginner
	case questionIndex <= 30:
		level = DifficultyMedium
	case questionIndex <= 50:
		level = DifficultyHard
	default:
		level = DifficultyMaster
	}

	return DifficultyProgression{level: level}
}

// GetDistribution returns question difficulty distribution for current level
// Per docs: 1-10: 80% easy 20% medium, 11-30: 100% medium, 31-50: 70% medium 30% hard, 51+: 100% hard
func (dp DifficultyProgression) GetDistribution() map[string]float64 {
	switch dp.level {
	case DifficultyBeginner:
		return map[string]float64{"easy": 0.8, "medium": 0.2, "hard": 0.0}
	case DifficultyMedium:
		return map[string]float64{"easy": 0.0, "medium": 1.0, "hard": 0.0}
	case DifficultyHard:
		return map[string]float64{"easy": 0.0, "medium": 0.7, "hard": 0.3}
	case DifficultyMaster:
		return map[string]float64{"easy": 0.0, "medium": 0.0, "hard": 1.0}
	default:
		return map[string]float64{"easy": 0.8, "medium": 0.2, "hard": 0.0}
	}
}

// GetTimeLimit returns time limit in seconds based on question index (1-based)
// Per docs: 1-10: 15s, 11-25: 12s, 26-50: 10s, 51+: 8s
func (dp DifficultyProgression) GetTimeLimit(questionIndex int) int {
	switch {
	case questionIndex <= 10:
		return 15
	case questionIndex <= 25:
		return 12
	case questionIndex <= 50:
		return 10
	default:
		return 8
	}
}

func (dp DifficultyProgression) Level() DifficultyLevel {
	return dp.level
}

// GameStatus represents the current status of a marathon game
type GameStatus string

const (
	GameStatusInProgress GameStatus = "in_progress"
	GameStatusGameOver   GameStatus = "game_over"  // 0 lives, continue offered (intermediate)
	GameStatusCompleted  GameStatus = "completed"  // Game ended normally (terminal)
	GameStatusAbandoned  GameStatus = "abandoned"  // Player quit voluntarily (terminal)
)

// State transition diagram for Marathon Game:
//
//   in_progress ──(no lives)──> game_over (intermediate: continue offered)
//   game_over ──(continue used)──> in_progress
//   game_over ──(decline continue / quit)──> completed (terminal)
//   in_progress ──(abandon)──> abandoned (terminal)
//
// Allowed transitions map
var marathonGameTransitions = map[GameStatus][]GameStatus{
	GameStatusInProgress: {GameStatusGameOver, GameStatusAbandoned},
	GameStatusGameOver:   {GameStatusInProgress, GameStatusCompleted}, // Can continue or finish
	GameStatusCompleted:  {}, // Terminal state - no transitions allowed
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

// PaymentMethod represents how a continue is paid for
type PaymentMethod string

const (
	PaymentCoins PaymentMethod = "coins"
	PaymentAd    PaymentMethod = "ad"
)

// ContinueCostCalculator calculates continue cost based on continue count
type ContinueCostCalculator struct{}

// GetCost returns coin cost for the next continue: 200, 400, 600, 800, ...
func (ccc ContinueCostCalculator) GetCost(continueCount int) int {
	return 200 + (continueCount * 200)
}

// HasAdOption returns true if ad-based continue is available (first 3 continues only)
func (ccc ContinueCostCalculator) HasAdOption(continueCount int) bool {
	return continueCount < 3
}

// Milestones for marathon progress tracking
var MarathonMilestones = []int{25, 50, 100, 200, 500}

// GetNextMilestone returns the next milestone target and remaining questions
func GetNextMilestone(currentScore int) (next int, remaining int) {
	for _, m := range MarathonMilestones {
		if currentScore < m {
			return m, m - currentScore
		}
	}
	// Past all milestones
	return 0, 0
}
