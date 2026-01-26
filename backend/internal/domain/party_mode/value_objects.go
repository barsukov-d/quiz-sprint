package party_mode

import (
	"fmt"
	"github.com/barsukov/quiz-sprint/backend/internal/domain/quiz"
	"github.com/barsukov/quiz-sprint/backend/internal/domain/shared"
	"github.com/google/uuid"
	"math/rand"
	"strings"
	"time"
)

// Type aliases from other domains
type UserID = shared.UserID
type QuizID = quiz.QuizID
type QuestionID = quiz.QuestionID
type AnswerID = quiz.AnswerID
type CategoryID = quiz.CategoryID

// RoomID uniquely identifies a party room
type RoomID struct {
	value string
}

func NewRoomID() RoomID {
	return RoomID{value: uuid.New().String()}
}

func NewRoomIDFromString(value string) RoomID {
	return RoomID{value: value}
}

func (id RoomID) String() string {
	return id.value
}

func (id RoomID) IsZero() bool {
	return id.value == ""
}

func (id RoomID) Equals(other RoomID) bool {
	return id.value == other.value
}

// RoomCode represents a human-readable room code (ABC-123)
type RoomCode struct {
	value string // Format: ABC-123 (3 uppercase letters + 3 digits)
}

func GenerateRoomCode() RoomCode {
	rand.Seed(time.Now().UnixNano())

	// Generate 3 random uppercase letters
	letters := ""
	for i := 0; i < 3; i++ {
		letters += string(rune('A' + rand.Intn(26)))
	}

	// Generate 3 random digits
	digits := fmt.Sprintf("%03d", rand.Intn(1000))

	code := fmt.Sprintf("%s-%s", letters, digits)
	return RoomCode{value: code}
}

func NewRoomCodeFromString(value string) RoomCode {
	// Normalize: uppercase, trim spaces
	normalized := strings.ToUpper(strings.TrimSpace(value))
	return RoomCode{value: normalized}
}

func (rc RoomCode) String() string {
	return rc.value
}

func (rc RoomCode) IsZero() bool {
	return rc.value == ""
}

func (rc RoomCode) Equals(other RoomCode) bool {
	return rc.value == other.value
}

// GameID uniquely identifies a party game (active game session)
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

// RoomSettings defines the game configuration
type RoomSettings struct {
	maxPlayers       int
	questionsCount   int
	timePerQuestion  int // seconds
	categories       []CategoryID
	difficulty       string // "easy", "mix", "hard"
	showCorrectAnswer bool
	showPlayerAnswers bool
	showCurrentScore bool
}

func NewRoomSettings() RoomSettings {
	return RoomSettings{
		maxPlayers:        6,
		questionsCount:    15,
		timePerQuestion:   15,
		categories:        []CategoryID{},
		difficulty:        "mix",
		showCorrectAnswer: true,
		showPlayerAnswers: true,
		showCurrentScore:  true,
	}
}

// Validation
func (rs RoomSettings) Validate() error {
	if rs.maxPlayers < 2 || rs.maxPlayers > 8 {
		return ErrInvalidRoomSettings
	}
	if rs.questionsCount < 10 || rs.questionsCount > 30 {
		return ErrInvalidRoomSettings
	}
	if rs.timePerQuestion < 10 || rs.timePerQuestion > 30 {
		return ErrInvalidRoomSettings
	}
	return nil
}

// Getters
func (rs RoomSettings) MaxPlayers() int           { return rs.maxPlayers }
func (rs RoomSettings) QuestionsCount() int       { return rs.questionsCount }
func (rs RoomSettings) TimePerQuestion() int      { return rs.timePerQuestion }
func (rs RoomSettings) Categories() []CategoryID  {
	copy := make([]CategoryID, len(rs.categories))
	for i, c := range rs.categories {
		copy[i] = c
	}
	return copy
}
func (rs RoomSettings) Difficulty() string        { return rs.difficulty }
func (rs RoomSettings) ShowCorrectAnswer() bool   { return rs.showCorrectAnswer }
func (rs RoomSettings) ShowPlayerAnswers() bool   { return rs.showPlayerAnswers }
func (rs RoomSettings) ShowCurrentScore() bool    { return rs.showCurrentScore }

// RoomPlayer represents a player in the room lobby
type RoomPlayer struct {
	userID    UserID
	username  string
	isHost    bool
	isReady   bool
	connected bool
	joinedAt  int64
}

func NewRoomPlayer(userID UserID, username string, isHost bool, joinedAt int64) RoomPlayer {
	return RoomPlayer{
		userID:    userID,
		username:  username,
		isHost:    isHost,
		isReady:   false, // New players are not ready by default
		connected: true,
		joinedAt:  joinedAt,
	}
}

// SetReady sets player ready status (immutable)
func (rp RoomPlayer) SetReady(ready bool) RoomPlayer {
	return RoomPlayer{
		userID:    rp.userID,
		username:  rp.username,
		isHost:    rp.isHost,
		isReady:   ready,
		connected: rp.connected,
		joinedAt:  rp.joinedAt,
	}
}

// SetHost sets player as host (immutable)
func (rp RoomPlayer) SetHost(isHost bool) RoomPlayer {
	return RoomPlayer{
		userID:    rp.userID,
		username:  rp.username,
		isHost:    isHost,
		isReady:   rp.isReady,
		connected: rp.connected,
		joinedAt:  rp.joinedAt,
	}
}

// SetConnected updates connection status (immutable)
func (rp RoomPlayer) SetConnected(connected bool) RoomPlayer {
	return RoomPlayer{
		userID:    rp.userID,
		username:  rp.username,
		isHost:    rp.isHost,
		isReady:   rp.isReady,
		connected: connected,
		joinedAt:  rp.joinedAt,
	}
}

// Getters
func (rp RoomPlayer) UserID() UserID    { return rp.userID }
func (rp RoomPlayer) Username() string  { return rp.username }
func (rp RoomPlayer) IsHost() bool      { return rp.isHost }
func (rp RoomPlayer) IsReady() bool     { return rp.isReady }
func (rp RoomPlayer) Connected() bool   { return rp.connected }
func (rp RoomPlayer) JoinedAt() int64   { return rp.joinedAt }

// PartyPlayer represents a player in active game (with score)
type PartyPlayer struct {
	userID       UserID
	username     string
	score        int
	connected    bool
	answersCount int
}

func NewPartyPlayer(userID UserID, username string) PartyPlayer {
	return PartyPlayer{
		userID:       userID,
		username:     username,
		score:        0,
		connected:    true,
		answersCount: 0,
	}
}

// AddScore adds points (immutable)
func (pp PartyPlayer) AddScore(points int) PartyPlayer {
	return PartyPlayer{
		userID:       pp.userID,
		username:     pp.username,
		score:        pp.score + points,
		connected:    pp.connected,
		answersCount: pp.answersCount + 1,
	}
}

// IncrementAnswers increments answer count (immutable)
func (pp PartyPlayer) IncrementAnswers() PartyPlayer {
	return PartyPlayer{
		userID:       pp.userID,
		username:     pp.username,
		score:        pp.score,
		connected:    pp.connected,
		answersCount: pp.answersCount + 1,
	}
}

// SetConnected updates connection (immutable)
func (pp PartyPlayer) SetConnected(connected bool) PartyPlayer {
	return PartyPlayer{
		userID:       pp.userID,
		username:     pp.username,
		score:        pp.score,
		connected:    connected,
		answersCount: pp.answersCount,
	}
}

// Getters
func (pp PartyPlayer) UserID() UserID     { return pp.userID }
func (pp PartyPlayer) Username() string   { return pp.username }
func (pp PartyPlayer) Score() int         { return pp.score }
func (pp PartyPlayer) Connected() bool    { return pp.connected }
func (pp PartyPlayer) AnswersCount() int  { return pp.answersCount }

// CalculateSpeedBonus calculates bonus based on answer time (Party Mode scoring)
func CalculateSpeedBonus(timeTakenMs int64) int {
	timeTakenSec := float64(timeTakenMs) / 1000.0

	switch {
	case timeTakenSec <= 2.0:
		return 75
	case timeTakenSec <= 4.0:
		return 50
	case timeTakenSec <= 6.0:
		return 35
	case timeTakenSec <= 8.0:
		return 20
	case timeTakenSec <= 10.0:
		return 10
	default:
		return 0
	}
}

// CalculatePositionBonus calculates bonus for being Nth to answer correctly
func CalculatePositionBonus(position int) int {
	switch position {
	case 1:
		return 25 // First to answer
	case 2:
		return 15 // Second
	case 3:
		return 10 // Third
	default:
		return 0
	}
}

// RoomStatus represents the current status of a party room
type RoomStatus string

const (
	RoomStatusLobby   RoomStatus = "lobby"   // Waiting for players
	RoomStatusPlaying RoomStatus = "playing" // Game in progress
	RoomStatusClosed  RoomStatus = "closed"  // Room closed
)

// State transition diagram for Party Room:
//
//   lobby ──(start game)──> playing
//   lobby ──(empty)──> closed (terminal)
//   playing ──(empty)──> closed (terminal)
//
// Allowed transitions map
var partyRoomTransitions = map[RoomStatus][]RoomStatus{
	RoomStatusLobby:   {RoomStatusPlaying, RoomStatusClosed},
	RoomStatusPlaying: {RoomStatusClosed},
	RoomStatusClosed:  {}, // Terminal state - no transitions allowed
}

// CanTransitionTo checks if transition to target status is valid
func (rs RoomStatus) CanTransitionTo(target RoomStatus) bool {
	allowedTargets, exists := partyRoomTransitions[rs]
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
func (rs RoomStatus) IsTerminal() bool {
	transitions, exists := partyRoomTransitions[rs]
	return exists && len(transitions) == 0
}

// GameStatus represents the current status of a party game
type GameStatus string

const (
	GameStatusInProgress GameStatus = "in_progress"
	GameStatusFinished   GameStatus = "finished"
)

// State transition diagram for Party Game:
//
//   in_progress ──(all questions completed)──> finished (terminal)
//
// Allowed transitions map
var partyGameTransitions = map[GameStatus][]GameStatus{
	GameStatusInProgress: {GameStatusFinished},
	GameStatusFinished:   {}, // Terminal state - no transitions allowed
}

// CanTransitionTo checks if transition to target status is valid
func (gs GameStatus) CanTransitionTo(target GameStatus) bool {
	allowedTargets, exists := partyGameTransitions[gs]
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
	transitions, exists := partyGameTransitions[gs]
	return exists && len(transitions) == 0
}
