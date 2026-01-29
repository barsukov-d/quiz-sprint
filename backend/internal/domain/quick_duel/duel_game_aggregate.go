package quick_duel

import (
	"github.com/barsukov/quiz-sprint/backend/internal/domain/quiz"
)

const (
	QuestionsPerDuel = 7  // Fixed: 7 questions per duel
	TimePerQuestionSec = 10 // 10 seconds per question
	BasePointsCorrect = 100 // Base points for correct answer
	MinAnswerTimeMs = 500 // Anti-cheat: minimum 0.5 sec
)

// RoundAnswer tracks a player's answer for a round
type RoundAnswer struct {
	playerID  UserID
	answerID  AnswerID
	timeTaken int64 // milliseconds
	isCorrect bool
	points    int
}

// DuelGame is the aggregate root for Quick Duel mode (1v1 PvP)
type DuelGame struct {
	id            GameID
	player1       DuelPlayer
	player2       DuelPlayer
	questionIDs   []QuestionID // 7 question IDs
	currentRound  int          // Current round (1-7, 0 = not started)
	status        GameStatus
	roundAnswers  map[int][]RoundAnswer // Round number -> answers
	startedAt     int64        // Unix timestamp when game started
	finishedAt    int64        // Unix timestamp when finished (0 if not finished)

	// Domain events collected during operations
	events []Event
}

// NewDuelGame creates a new duel game after matchmaking
func NewDuelGame(
	player1 DuelPlayer,
	player2 DuelPlayer,
	questionIDs []QuestionID,
	createdAt int64,
) (*DuelGame, error) {
	// Validate
	if player1.UserID().IsZero() || player2.UserID().IsZero() {
		return nil, ErrInvalidGameID
	}

	if len(questionIDs) != QuestionsPerDuel {
		return nil, ErrInvalidRound
	}

	// Create
	game := &DuelGame{
		id:           NewGameID(),
		player1:      player1,
		player2:      player2,
		questionIDs:  questionIDs,
		currentRound: 0, // Not started yet
		status:       GameStatusWaitingStart,
		roundAnswers: make(map[int][]RoundAnswer),
		startedAt:    0,
		finishedAt:   0,
		events:       make([]Event, 0),
	}

	// Publish DuelGameCreated event
	game.events = append(game.events, NewDuelGameCreatedEvent(
		game.id,
		player1,
		player2,
		questionIDs,
		createdAt,
	))

	return game, nil
}

// Start starts the game (both players ready)
func (dg *DuelGame) Start(startedAt int64) error {
	if dg.status != GameStatusWaitingStart {
		return ErrGameNotActive
	}

	// Validate state transition
	if !dg.status.CanTransitionTo(GameStatusInProgress) {
		return ErrInvalidGameStatus
	}

	dg.status = GameStatusInProgress
	dg.startedAt = startedAt
	dg.currentRound = 1 // Start first round

	// Publish DuelGameStarted event
	dg.events = append(dg.events, NewDuelGameStartedEvent(
		dg.id,
		dg.player1.UserID(),
		dg.player2.UserID(),
		startedAt,
	))

	// Publish first RoundStarted event
	questionID := dg.questionIDs[0]
	dg.events = append(dg.events, NewRoundStartedEvent(
		dg.id,
		dg.currentRound,
		questionID,
		startedAt,
	))

	return nil
}

// SubmitAnswerResult holds result of submitting an answer
type SubmitAnswerResult struct {
	IsCorrect       bool
	PointsEarned    int
	PlayerScore     int
	OpponentScore   int
	RoundNumber     int
	BothAnswered    bool // True if both players answered this round
	IsGameFinished  bool
	WinnerID        *UserID // Set if game finished
}

// SubmitAnswer processes a player's answer for current round
func (dg *DuelGame) SubmitAnswer(
	playerID UserID,
	answerID AnswerID,
	timeTaken int64,
	question *quiz.Question,
	answeredAt int64,
) (*SubmitAnswerResult, error) {
	// 1. Validate game state
	if dg.status != GameStatusInProgress {
		return nil, ErrGameNotActive
	}

	// 2. Validate player is in game
	if !dg.isPlayerInGame(playerID) {
		return nil, ErrPlayerNotInGame
	}

	// 3. Anti-cheat: validate answer time
	if timeTaken < MinAnswerTimeMs {
		return nil, ErrInvalidAnswerTime
	}

	// 4. Check if player already answered this round
	if dg.hasPlayerAnsweredRound(playerID, dg.currentRound) {
		return nil, ErrPlayerAlreadyAnswered
	}

	// 5. Validate answer
	answer, err := question.GetAnswer(answerID)
	if err != nil {
		return nil, err
	}

	isCorrect := answer.IsCorrect()

	// 6. Calculate points
	points := 0
	if isCorrect {
		basePoints := BasePointsCorrect
		speedBonus := CalculateSpeedBonus(timeTaken)
		points = basePoints + speedBonus
	}

	// 7. Record answer
	roundAnswer := RoundAnswer{
		playerID:  playerID,
		answerID:  answerID,
		timeTaken: timeTaken,
		isCorrect: isCorrect,
		points:    points,
	}
	dg.roundAnswers[dg.currentRound] = append(dg.roundAnswers[dg.currentRound], roundAnswer)

	// 8. Update player score
	if dg.player1.UserID().Equals(playerID) {
		dg.player1 = dg.player1.AddScore(points)
	} else {
		dg.player2 = dg.player2.AddScore(points)
	}

	// 9. Publish PlayerAnswered event
	dg.events = append(dg.events, NewPlayerAnsweredEvent(
		dg.id,
		playerID,
		question.ID(),
		answerID,
		timeTaken,
		isCorrect,
		points,
		answeredAt,
	))

	// 10. Check if both players answered
	bothAnswered := len(dg.roundAnswers[dg.currentRound]) == 2

	result := &SubmitAnswerResult{
		IsCorrect:      isCorrect,
		PointsEarned:   points,
		PlayerScore:    dg.getPlayerScore(playerID),
		OpponentScore:  dg.getOpponentScore(playerID),
		RoundNumber:    dg.currentRound,
		BothAnswered:   bothAnswered,
		IsGameFinished: false,
		WinnerID:       nil,
	}

	// 11. If both answered, complete round
	if bothAnswered {
		if err := dg.completeRound(answeredAt); err != nil {
			return nil, err
		}

		// Check if game finished
		if dg.status == GameStatusFinished {
			result.IsGameFinished = true
			result.WinnerID = dg.determineWinner()
		}
	}

	return result, nil
}

// completeRound completes current round and starts next (or finishes game)
func (dg *DuelGame) completeRound(completedAt int64) error {
	// 1. Publish RoundCompleted event
	answers := dg.roundAnswers[dg.currentRound]
	player1Answered := false
	player2Answered := false

	for _, ans := range answers {
		if dg.player1.UserID().Equals(ans.playerID) {
			player1Answered = true
		} else {
			player2Answered = true
		}
	}

	dg.events = append(dg.events, NewRoundCompletedEvent(
		dg.id,
		dg.currentRound,
		dg.player1.Score(),
		dg.player2.Score(),
		player1Answered,
		player2Answered,
		completedAt,
	))

	// 2. Check if all rounds completed
	if dg.currentRound >= QuestionsPerDuel {
		return dg.finishGame(completedAt)
	}

	// 3. Start next round
	dg.currentRound++
	questionID := dg.questionIDs[dg.currentRound-1]

	dg.events = append(dg.events, NewRoundStartedEvent(
		dg.id,
		dg.currentRound,
		questionID,
		completedAt,
	))

	return nil
}

// finishGame finishes the game and calculates ELO changes
func (dg *DuelGame) finishGame(finishedAt int64) error {
	// Validate state transition
	if !dg.status.CanTransitionTo(GameStatusFinished) {
		return ErrInvalidGameStatus
	}

	dg.status = GameStatusFinished
	dg.finishedAt = finishedAt

	// Determine winner
	var player1Won bool
	if dg.player1.Score() > dg.player2.Score() {
		player1Won = true
	} else if dg.player1.Score() < dg.player2.Score() {
		player1Won = false
	} else {
		// Draw - both get draw result for ELO
		player1Won = false // Will be treated as draw in ELO calculation
	}

	// Calculate new ELO ratings
	player1NewElo := dg.player1.Elo().CalculateNewRating(player1Won, dg.player2.Elo().Rating())
	player2NewElo := dg.player2.Elo().CalculateNewRating(!player1Won, dg.player1.Elo().Rating())

	// Update players' ELO
	dg.player1 = dg.player1.UpdateElo(player1NewElo)
	dg.player2 = dg.player2.UpdateElo(player2NewElo)

	// Determine winner ID
	var winnerID *UserID
	if dg.player1.Score() > dg.player2.Score() {
		id := dg.player1.UserID()
		winnerID = &id
	} else if dg.player2.Score() > dg.player1.Score() {
		id := dg.player2.UserID()
		winnerID = &id
	}
	// else: draw, winnerID = nil

	// Publish DuelGameFinished event
	dg.events = append(dg.events, NewDuelGameFinishedEvent(
		dg.id,
		winnerID,
		dg.player1,
		dg.player2,
		player1NewElo,
		player2NewElo,
		finishedAt,
	))

	return nil
}

// HandlePlayerDisconnect handles player disconnect
func (dg *DuelGame) HandlePlayerDisconnect(playerID UserID, disconnectedAt int64) error {
	if !dg.isPlayerInGame(playerID) {
		return ErrPlayerNotInGame
	}

	// Update connection status
	if dg.player1.UserID().Equals(playerID) {
		dg.player1 = dg.player1.SetConnected(false)
	} else {
		dg.player2 = dg.player2.SetConnected(false)
	}

	// Publish PlayerDisconnected event
	dg.events = append(dg.events, NewPlayerDisconnectedEvent(
		dg.id,
		playerID,
		disconnectedAt,
	))

	// Check if both disconnected
	if !dg.player1.Connected() && !dg.player2.Connected() {
		// Validate state transition
		if !dg.status.CanTransitionTo(GameStatusAbandoned) {
			return ErrInvalidGameStatus
		}

		dg.status = GameStatusAbandoned
		return ErrBothPlayersDisconnected
	}

	return nil
}

// HandlePlayerReconnect handles player reconnect
func (dg *DuelGame) HandlePlayerReconnect(playerID UserID, reconnectedAt int64) error {
	if !dg.isPlayerInGame(playerID) {
		return ErrPlayerNotInGame
	}

	// Update connection status
	if dg.player1.UserID().Equals(playerID) {
		dg.player1 = dg.player1.SetConnected(true)
	} else {
		dg.player2 = dg.player2.SetConnected(true)
	}

	// Publish PlayerReconnected event
	dg.events = append(dg.events, NewPlayerReconnectedEvent(
		dg.id,
		playerID,
		reconnectedAt,
	))

	return nil
}

// Helper methods

func (dg *DuelGame) isPlayerInGame(playerID UserID) bool {
	return dg.player1.UserID().Equals(playerID) || dg.player2.UserID().Equals(playerID)
}

func (dg *DuelGame) hasPlayerAnsweredRound(playerID UserID, round int) bool {
	answers, exists := dg.roundAnswers[round]
	if !exists {
		return false
	}

	for _, ans := range answers {
		if ans.playerID.Equals(playerID) {
			return true
		}
	}
	return false
}

func (dg *DuelGame) getPlayerScore(playerID UserID) int {
	if dg.player1.UserID().Equals(playerID) {
		return dg.player1.Score()
	}
	return dg.player2.Score()
}

func (dg *DuelGame) getOpponentScore(playerID UserID) int {
	if dg.player1.UserID().Equals(playerID) {
		return dg.player2.Score()
	}
	return dg.player1.Score()
}

func (dg *DuelGame) determineWinner() *UserID {
	if dg.player1.Score() > dg.player2.Score() {
		id := dg.player1.UserID()
		return &id
	} else if dg.player2.Score() > dg.player1.Score() {
		id := dg.player2.UserID()
		return &id
	}
	return nil // Draw
}

// Getters
func (dg *DuelGame) ID() GameID            { return dg.id }
func (dg *DuelGame) Player1() DuelPlayer   { return dg.player1 }
func (dg *DuelGame) Player2() DuelPlayer   { return dg.player2 }
func (dg *DuelGame) QuestionIDs() []QuestionID {
	copy := make([]QuestionID, len(dg.questionIDs))
	for i, qid := range dg.questionIDs {
		copy[i] = qid
	}
	return copy
}
func (dg *DuelGame) CurrentRound() int     { return dg.currentRound }
func (dg *DuelGame) Status() GameStatus    { return dg.status }
func (dg *DuelGame) StartedAt() int64      { return dg.startedAt }
func (dg *DuelGame) FinishedAt() int64     { return dg.finishedAt }
func (dg *DuelGame) IsFinished() bool      { return dg.status.IsTerminal() }

// Events returns collected domain events and clears them
func (dg *DuelGame) Events() []Event {
	events := dg.events
	dg.events = make([]Event, 0)
	return events
}

// ReconstructDuelGame reconstructs a DuelGame from persistence
func ReconstructDuelGame(
	id GameID,
	player1 DuelPlayer,
	player2 DuelPlayer,
	questionIDs []QuestionID,
	currentRound int,
	status GameStatus,
	roundAnswers map[int][]RoundAnswer,
	startedAt int64,
	finishedAt int64,
) *DuelGame {
	return &DuelGame{
		id:           id,
		player1:      player1,
		player2:      player2,
		questionIDs:  questionIDs,
		currentRound: currentRound,
		status:       status,
		roundAnswers: roundAnswers,
		startedAt:    startedAt,
		finishedAt:   finishedAt,
		events:       make([]Event, 0), // Don't replay events from DB
	}
}
