package quick_duel

import (
	"github.com/barsukov/quiz-sprint/backend/internal/domain/quiz"
	"github.com/barsukov/quiz-sprint/backend/internal/domain/shared"
	"github.com/google/uuid"
	"math"
)

// Type aliases from other domains
type UserID = shared.UserID
type QuizID = quiz.QuizID
type QuestionID = quiz.QuestionID
type AnswerID = quiz.AnswerID
type Points = quiz.Points

// GameID uniquely identifies a duel game
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

// EloRating represents a player's ELO rating for matchmaking
type EloRating struct {
	rating      int
	gamesPlayed int
}

const (
	InitialEloRating = 1000
	MinEloRating     = 100
	KFactorNew       = 32 // K-factor for first 30 games
	KFactorRegular   = 16 // K-factor after 30 games
	NewPlayerGames   = 30 // Number of games considered "new player"
)

func NewEloRating() EloRating {
	return EloRating{
		rating:      InitialEloRating,
		gamesPlayed: 0,
	}
}

// ReconstructEloRating reconstructs EloRating from persistence
func ReconstructEloRating(rating int, gamesPlayed int) EloRating {
	return EloRating{
		rating:      rating,
		gamesPlayed: gamesPlayed,
	}
}

// KFactor returns the K-factor based on games played
func (e EloRating) KFactor() int {
	if e.gamesPlayed < NewPlayerGames {
		return KFactorNew
	}
	return KFactorRegular
}

// CalculateNewRating calculates new ELO rating after a game
// won: true if player won, false if lost
// opponentRating: opponent's ELO rating
func (e EloRating) CalculateNewRating(won bool, opponentRating int) EloRating {
	// Expected score (probability of winning)
	expectedScore := 1.0 / (1.0 + math.Pow(10, float64(opponentRating-e.rating)/400.0))

	// Actual score
	var actualScore float64
	if won {
		actualScore = 1.0
	} else {
		actualScore = 0.0
	}

	// Calculate rating change
	kFactor := float64(e.KFactor())
	ratingChange := kFactor * (actualScore - expectedScore)

	// Apply change
	newRating := e.rating + int(math.Round(ratingChange))

	// Enforce minimum rating
	if newRating < MinEloRating {
		newRating = MinEloRating
	}

	return EloRating{
		rating:      newRating,
		gamesPlayed: e.gamesPlayed + 1,
	}
}

// IsNewPlayer checks if player is still in "new player" bracket
func (e EloRating) IsNewPlayer() bool {
	return e.gamesPlayed < NewPlayerGames
}

// GetMatchmakingRange returns ELO range for matchmaking based on search duration
// searchSeconds: how long player has been searching
func (e EloRating) GetMatchmakingRange(searchSeconds int) (min int, max int) {
	var rangeDelta int

	switch {
	case searchSeconds < 5:
		rangeDelta = 50
	case searchSeconds < 10:
		rangeDelta = 100
	case searchSeconds < 15:
		rangeDelta = 200
	default:
		// After 15 seconds, match with anyone
		return MinEloRating, 9999
	}

	min = e.rating - rangeDelta
	max = e.rating + rangeDelta

	if min < MinEloRating {
		min = MinEloRating
	}

	return min, max
}

// Getters
func (e EloRating) Rating() int      { return e.rating }
func (e EloRating) GamesPlayed() int { return e.gamesPlayed }

// DuelPlayer represents a player in the duel (value object)
type DuelPlayer struct {
	userID       UserID
	username     string
	elo          EloRating
	score        int  // Current score in this game
	connected    bool // Connection status
	answersCount int  // Number of questions answered
}

func NewDuelPlayer(userID UserID, username string, elo EloRating) DuelPlayer {
	return DuelPlayer{
		userID:       userID,
		username:     username,
		elo:          elo,
		score:        0,
		connected:    true,
		answersCount: 0,
	}
}

// AddScore adds points to player's score (immutable)
func (dp DuelPlayer) AddScore(points int) DuelPlayer {
	return DuelPlayer{
		userID:       dp.userID,
		username:     dp.username,
		elo:          dp.elo,
		score:        dp.score + points,
		connected:    dp.connected,
		answersCount: dp.answersCount + 1,
	}
}

// IncrementAnswers increments answered questions count (immutable)
func (dp DuelPlayer) IncrementAnswers() DuelPlayer {
	return DuelPlayer{
		userID:       dp.userID,
		username:     dp.username,
		elo:          dp.elo,
		score:        dp.score,
		connected:    dp.connected,
		answersCount: dp.answersCount + 1,
	}
}

// SetConnected updates connection status (immutable)
func (dp DuelPlayer) SetConnected(connected bool) DuelPlayer {
	return DuelPlayer{
		userID:       dp.userID,
		username:     dp.username,
		elo:          dp.elo,
		score:        dp.score,
		connected:    connected,
		answersCount: dp.answersCount,
	}
}

// UpdateElo updates player's ELO (immutable)
func (dp DuelPlayer) UpdateElo(newElo EloRating) DuelPlayer {
	return DuelPlayer{
		userID:       dp.userID,
		username:     dp.username,
		elo:          newElo,
		score:        dp.score,
		connected:    dp.connected,
		answersCount: dp.answersCount,
	}
}

// Getters
func (dp DuelPlayer) UserID() UserID       { return dp.userID }
func (dp DuelPlayer) Username() string     { return dp.username }
func (dp DuelPlayer) Elo() EloRating       { return dp.elo }
func (dp DuelPlayer) Score() int           { return dp.score }
func (dp DuelPlayer) Connected() bool      { return dp.connected }
func (dp DuelPlayer) AnswersCount() int    { return dp.answersCount }

// WinStreak represents a player's winning streak
type WinStreak struct {
	currentStreak int
	bestStreak    int
}

func NewWinStreak() WinStreak {
	return WinStreak{
		currentStreak: 0,
		bestStreak:    0,
	}
}

// ReconstructWinStreak reconstructs from persistence
func ReconstructWinStreak(currentStreak int, bestStreak int) WinStreak {
	return WinStreak{
		currentStreak: currentStreak,
		bestStreak:    bestStreak,
	}
}

// IncrementWin increments streak on win (immutable)
func (ws WinStreak) IncrementWin() WinStreak {
	newCurrent := ws.currentStreak + 1
	newBest := ws.bestStreak

	if newCurrent > newBest {
		newBest = newCurrent
	}

	return WinStreak{
		currentStreak: newCurrent,
		bestStreak:    newBest,
	}
}

// ResetOnLoss resets current streak on loss (immutable)
func (ws WinStreak) ResetOnLoss() WinStreak {
	return WinStreak{
		currentStreak: 0,
		bestStreak:    ws.bestStreak,
	}
}

// GetBonusMultiplier returns reward multiplier based on current streak
func (ws WinStreak) GetBonusMultiplier() float64 {
	switch {
	case ws.currentStreak >= 10:
		return 1.5 // +50%
	case ws.currentStreak >= 5:
		return 1.25 // +25%
	case ws.currentStreak >= 3:
		return 1.1 // +10%
	default:
		return 1.0 // No bonus
	}
}

// IsMilestone checks if current streak is a milestone (3, 5, 10)
func (ws WinStreak) IsMilestone() bool {
	milestones := []int{3, 5, 10}
	for _, m := range milestones {
		if ws.currentStreak == m {
			return true
		}
	}
	return false
}

// Getters
func (ws WinStreak) CurrentStreak() int { return ws.currentStreak }
func (ws WinStreak) BestStreak() int    { return ws.bestStreak }

// SpeedBonus calculates bonus points based on answer time
func CalculateSpeedBonus(timeTakenMs int64) int {
	timeTakenSec := float64(timeTakenMs) / 1000.0

	switch {
	case timeTakenSec <= 3.0:
		return 50
	case timeTakenSec <= 5.0:
		return 25
	case timeTakenSec <= 7.0:
		return 10
	default:
		return 0
	}
}

// GameStatus represents the current status of a duel game
type GameStatus string

const (
	GameStatusWaitingStart GameStatus = "waiting_start" // Created, waiting for both players to be ready
	GameStatusInProgress   GameStatus = "in_progress"   // Active game
	GameStatusFinished     GameStatus = "finished"      // Completed
	GameStatusAbandoned    GameStatus = "abandoned"     // One/both players disconnected
)

// State transition diagram for Duel Game:
//
//   waiting_start ──(start)──> in_progress
//   in_progress ──(finish)──> finished (terminal)
//   waiting_start ──(both disconnect)──> abandoned (terminal)
//   in_progress ──(both disconnect)──> abandoned (terminal)
//
// Allowed transitions map
var duelGameTransitions = map[GameStatus][]GameStatus{
	GameStatusWaitingStart: {GameStatusInProgress, GameStatusAbandoned},
	GameStatusInProgress:   {GameStatusFinished, GameStatusAbandoned},
	GameStatusFinished:     {}, // Terminal state - no transitions allowed
	GameStatusAbandoned:    {}, // Terminal state - no transitions allowed
}

// CanTransitionTo checks if transition to target status is valid
func (gs GameStatus) CanTransitionTo(target GameStatus) bool {
	allowedTargets, exists := duelGameTransitions[gs]
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
	transitions, exists := duelGameTransitions[gs]
	return exists && len(transitions) == 0
}
