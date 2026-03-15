package quick_duel

import (
	"github.com/barsukov/quiz-sprint/backend/internal/domain/quiz"
)

const (
	QuestionsPerDuel = 7  // Fixed: 7 questions per duel
	TimePerQuestionSec = 10 // 10 seconds per question
	BasePointsCorrect = 100 // Base points for correct answer
	MinAnswerTimeMs = 200 // Anti-cheat: minimum 0.2 sec
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

	// 3. Anti-cheat: validate and clamp answer time
	if timeTaken < 0 {
		return nil, ErrInvalidAnswerTime
	}
	const maxTimeTakenMs = int64(TimePerQuestionSec * 1000)       // 10000ms
	const networkToleranceMs = int64(500)                          // 500ms tolerance
	if timeTaken > maxTimeTakenMs+networkToleranceMs {
		timeTaken = maxTimeTakenMs // clamp to max, no speed bonus
	}
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
		dg.player1 = dg.player1.AddScore(points, timeTaken)
	} else {
		dg.player2 = dg.player2.AddScore(points, timeTaken)
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

	// Calculate new ELO ratings
	var player1NewElo, player2NewElo EloRating
	if dg.player1.Score() > dg.player2.Score() {
		player1NewElo = dg.player1.Elo().CalculateNewRating(true, dg.player2.Elo().Rating())
		player2NewElo = dg.player2.Elo().CalculateNewRating(false, dg.player1.Elo().Rating())
	} else if dg.player1.Score() < dg.player2.Score() {
		player1NewElo = dg.player1.Elo().CalculateNewRating(false, dg.player2.Elo().Rating())
		player2NewElo = dg.player2.Elo().CalculateNewRating(true, dg.player1.Elo().Rating())
	} else {
		// Draw - both get symmetric draw ELO adjustment
		player1NewElo = dg.player1.Elo().CalculateDrawRating(dg.player2.Elo().Rating())
		player2NewElo = dg.player2.Elo().CalculateDrawRating(dg.player1.Elo().Rating())
	}

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

// RecordTimeoutAnswer records a timeout (no answer) for a player in the current round.
// Used by the WebSocket hub when the round timer expires without a player answering.
// Returns ErrPlayerAlreadyAnswered if the player already submitted an answer (idempotent).
func (dg *DuelGame) RecordTimeoutAnswer(playerID UserID, timedOutAt int64) (*SubmitAnswerResult, error) {
	if dg.status != GameStatusInProgress {
		return nil, ErrGameNotActive
	}
	if !dg.isPlayerInGame(playerID) {
		return nil, ErrPlayerNotInGame
	}
	if dg.hasPlayerAnsweredRound(playerID, dg.currentRound) {
		return nil, ErrPlayerAlreadyAnswered
	}

	// Record as timed out: wrong, max time, 0 points, zero AnswerID
	const timeoutMs = TimePerQuestionSec * 1000
	roundAnswer := RoundAnswer{
		playerID:  playerID,
		timeTaken: timeoutMs,
		isCorrect: false,
		points:    0,
	}
	dg.roundAnswers[dg.currentRound] = append(dg.roundAnswers[dg.currentRound], roundAnswer)

	// Update player total time for tiebreaker tracking
	if dg.player1.UserID().Equals(playerID) {
		dg.player1 = dg.player1.AddTime(timeoutMs)
	} else {
		dg.player2 = dg.player2.AddTime(timeoutMs)
	}

	bothAnswered := len(dg.roundAnswers[dg.currentRound]) == 2
	result := &SubmitAnswerResult{
		IsCorrect:      false,
		PointsEarned:   0,
		PlayerScore:    dg.getPlayerScore(playerID),
		OpponentScore:  dg.getOpponentScore(playerID),
		RoundNumber:    dg.currentRound,
		BothAnswered:   bothAnswered,
		IsGameFinished: false,
	}

	if bothAnswered {
		if err := dg.completeRound(timedOutAt); err != nil {
			return nil, err
		}
		if dg.status == GameStatusFinished {
			result.IsGameFinished = true
			result.WinnerID = dg.determineWinner()
		}
	}

	return result, nil
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

// SurrenderResult holds the result of a surrender operation
type SurrenderResult struct {
	WinnerID UserID
}

// Surrender allows a player to forfeit the game mid-match.
// The surrendering player loses; the opponent wins with full ELO gain.
// Surrender is only allowed after the player has answered at least 3 questions.
func (dg *DuelGame) Surrender(playerID UserID, surrenderedAt int64) (*SurrenderResult, error) {
	if dg.status != GameStatusInProgress {
		return nil, ErrGameNotActive
	}
	if !dg.isPlayerInGame(playerID) {
		return nil, ErrPlayerNotInGame
	}

	// Enforce minimum: player must have answered at least 3 questions
	answeredCount := dg.countPlayerAnswers(playerID)
	if answeredCount < 3 {
		return nil, ErrTooEarlyToSurrender
	}

	if !dg.status.CanTransitionTo(GameStatusFinished) {
		return nil, ErrInvalidGameStatus
	}

	// Determine opponent
	var opponentID UserID
	if dg.player1.UserID().Equals(playerID) {
		opponentID = dg.player2.UserID()
	} else {
		opponentID = dg.player1.UserID()
	}

	// Publish surrender event before finishing
	dg.events = append(dg.events, NewPlayerSurrenderedEvent(
		dg.id,
		playerID,
		opponentID,
		surrenderedAt,
	))

	// Force scores: surrendering player gets 0, opponent gets 1 to guarantee win for ELO
	if dg.player1.UserID().Equals(playerID) {
		dg.player1 = dg.player1.WithScore(0)
		dg.player2 = dg.player2.WithScore(dg.player2.Score() + 1)
	} else {
		dg.player2 = dg.player2.WithScore(0)
		dg.player1 = dg.player1.WithScore(dg.player1.Score() + 1)
	}

	if err := dg.finishGame(surrenderedAt); err != nil {
		return nil, err
	}

	return &SurrenderResult{WinnerID: opponentID}, nil
}

// ReplayRoundAnswer restores a previously submitted answer into the in-memory roundAnswers map.
// Used to re-hydrate state from an external cache (Redis) since roundAnswers are not persisted to DB.
// Does NOT trigger game logic, events, or score changes — purely state restoration.
func (dg *DuelGame) ReplayRoundAnswer(round int, playerID UserID, answerID AnswerID, timeTaken int64, isCorrect bool, points int) {
	if dg.roundAnswers == nil {
		dg.roundAnswers = make(map[int][]RoundAnswer)
	}
	// Skip if already present
	for _, a := range dg.roundAnswers[round] {
		if a.playerID.Equals(playerID) {
			return
		}
	}
	dg.roundAnswers[round] = append(dg.roundAnswers[round], RoundAnswer{
		playerID:  playerID,
		answerID:  answerID,
		timeTaken: timeTaken,
		isCorrect: isCorrect,
		points:    points,
	})
}

// Helper methods

func (dg *DuelGame) countPlayerAnswers(playerID UserID) int {
	count := 0
	for _, answers := range dg.roundAnswers {
		for _, ans := range answers {
			if ans.playerID.Equals(playerID) {
				count++
				break
			}
		}
	}
	return count
}

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

	// Scores are equal: tiebreaker by total answer time (less is better)
	if dg.player1.TotalTimeMs() < dg.player2.TotalTimeMs() {
		id := dg.player1.UserID()
		return &id
	} else if dg.player2.TotalTimeMs() < dg.player1.TotalTimeMs() {
		id := dg.player2.UserID()
		return &id
	}

	// Times also equal: deterministic tiebreaker by smaller playerID string
	if dg.player1.UserID().String() < dg.player2.UserID().String() {
		id := dg.player1.UserID()
		return &id
	}
	id := dg.player2.UserID()
	return &id
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
