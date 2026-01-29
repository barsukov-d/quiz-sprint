package party_mode

import "github.com/barsukov/quiz-sprint/backend/internal/domain/quiz"

const (
	BasePointsCorrect = 100 // Base points for correct answer
)

// QuestionAnswer tracks answers for a question
type QuestionAnswer struct {
	playerID  UserID
	answerID  AnswerID
	timeTaken int64
	isCorrect bool
	points    int
	position  int // 1st, 2nd, 3rd to answer correctly
}

// PartyGame is the aggregate root for active party game
type PartyGame struct {
	id              GameID
	roomID          RoomID
	questionIDs     []QuestionID
	players         []PartyPlayer
	currentQuestion int                       // Current question index (0-based)
	questionAnswers map[int][]QuestionAnswer  // Question index -> answers
	status          GameStatus
	startedAt       int64
	finishedAt      int64

	// Domain events collected during operations
	events []Event
}

// NewPartyGame creates a new party game
func NewPartyGame(
	roomID RoomID,
	questionIDs []QuestionID,
	roomPlayers []RoomPlayer,
	startedAt int64,
) (*PartyGame, error) {
	// Validate
	if roomID.IsZero() {
		return nil, ErrInvalidRoomID
	}

	if len(questionIDs) == 0 {
		return nil, ErrInvalidGameID
	}

	if len(roomPlayers) < 2 {
		return nil, ErrNotEnoughPlayers
	}

	// Convert RoomPlayers to PartyPlayers
	players := make([]PartyPlayer, len(roomPlayers))
	playerIDs := make([]UserID, len(roomPlayers))
	for i, rp := range roomPlayers {
		players[i] = NewPartyPlayer(rp.UserID(), rp.Username())
		playerIDs[i] = rp.UserID()
	}

	// Create
	gameID := NewGameID()
	game := &PartyGame{
		id:              gameID,
		roomID:          roomID,
		questionIDs:     questionIDs,
		players:         players,
		currentQuestion: 0,
		questionAnswers: make(map[int][]QuestionAnswer),
		status:          GameStatusInProgress,
		startedAt:       startedAt,
		finishedAt:      0,
		events:          make([]Event, 0),
	}

	// Publish GameStarted event
	game.events = append(game.events, NewGameStartedEvent(
		gameID,
		roomID,
		playerIDs,
		questionIDs,
		startedAt,
	))

	// Publish first QuestionStarted event
	game.events = append(game.events, NewQuestionStartedEvent(
		gameID,
		questionIDs[0],
		1, // Question number (1-based)
		startedAt,
	))

	return game, nil
}

// SubmitAnswerResult holds result of answering
type SubmitAnswerResult struct {
	IsCorrect      bool
	PointsEarned   int
	Position       int // Position in correct answers (1st, 2nd, 3rd)
	PlayerScore    int
	QuestionNumber int
	AllAnswered    bool // True if all players answered
	IsGameFinished bool
	WinnerID       *UserID
}

// SubmitAnswer processes a player's answer
func (pg *PartyGame) SubmitAnswer(
	playerID UserID,
	answerID AnswerID,
	timeTaken int64,
	question *quiz.Question,
	answeredAt int64,
) (*SubmitAnswerResult, error) {
	// 1. Validate
	if pg.status != GameStatusInProgress {
		return nil, ErrGameNotActive
	}

	if !pg.hasPlayer(playerID) {
		return nil, ErrPlayerNotFound
	}

	if pg.hasPlayerAnsweredCurrentQuestion(playerID) {
		return nil, ErrPlayerAlreadyAnswered
	}

	// 2. Validate answer
	answer, err := question.GetAnswer(answerID)
	if err != nil {
		return nil, err
	}

	isCorrect := answer.IsCorrect()

	// 3. Calculate points
	points := 0
	position := 0

	if isCorrect {
		// Base points
		points = BasePointsCorrect

		// Speed bonus
		speedBonus := CalculateSpeedBonus(timeTaken)
		points += speedBonus

		// Position bonus (for being Nth to answer correctly)
		position = pg.countCorrectAnswersForCurrentQuestion() + 1
		positionBonus := CalculatePositionBonus(position)
		points += positionBonus
	}

	// 4. Record answer
	questionAnswer := QuestionAnswer{
		playerID:  playerID,
		answerID:  answerID,
		timeTaken: timeTaken,
		isCorrect: isCorrect,
		points:    points,
		position:  position,
	}
	pg.questionAnswers[pg.currentQuestion] = append(pg.questionAnswers[pg.currentQuestion], questionAnswer)

	// 5. Update player score
	playerIndex := pg.findPlayerIndex(playerID)
	pg.players[playerIndex] = pg.players[playerIndex].AddScore(points)

	// 6. Publish PartyPlayerAnswered event
	pg.events = append(pg.events, NewPartyPlayerAnsweredEvent(
		pg.id,
		playerID,
		question.ID(),
		answerID,
		timeTaken,
		isCorrect,
		points,
		position,
		answeredAt,
	))

	// 7. Check if all players answered
	allAnswered := len(pg.questionAnswers[pg.currentQuestion]) == len(pg.players)

	result := &SubmitAnswerResult{
		IsCorrect:      isCorrect,
		PointsEarned:   points,
		Position:       position,
		PlayerScore:    pg.players[playerIndex].Score(),
		QuestionNumber: pg.currentQuestion + 1,
		AllAnswered:    allAnswered,
		IsGameFinished: false,
		WinnerID:       nil,
	}

	// 8. If all answered, move to next question or finish
	if allAnswered {
		if err := pg.completeCurrentQuestion(answeredAt); err != nil {
			return nil, err
		}

		if pg.status == GameStatusFinished {
			result.IsGameFinished = true
			winnerID := pg.determineWinner()
			result.WinnerID = &winnerID
		}
	}

	return result, nil
}

// completeCurrentQuestion completes current question and moves to next (or finishes)
func (pg *PartyGame) completeCurrentQuestion(completedAt int64) error {
	// 1. Publish QuestionCompleted event
	currentQuestionID := pg.questionIDs[pg.currentQuestion]
	pg.events = append(pg.events, NewQuestionCompletedEvent(
		pg.id,
		currentQuestionID,
		pg.currentQuestion+1,
		completedAt,
	))

	// 2. Check if all questions completed
	if pg.currentQuestion >= len(pg.questionIDs)-1 {
		return pg.finishGame(completedAt)
	}

	// 3. Move to next question
	pg.currentQuestion++
	nextQuestionID := pg.questionIDs[pg.currentQuestion]

	pg.events = append(pg.events, NewQuestionStartedEvent(
		pg.id,
		nextQuestionID,
		pg.currentQuestion+1,
		completedAt,
	))

	return nil
}

// finishGame finishes the game
func (pg *PartyGame) finishGame(finishedAt int64) error {
	// Validate state transition
	if !pg.status.CanTransitionTo(GameStatusFinished) {
		return ErrInvalidGameStatus
	}

	pg.status = GameStatusFinished
	pg.finishedAt = finishedAt

	winnerID := pg.determineWinner()

	// Publish GameFinished event
	pg.events = append(pg.events, NewGameFinishedEvent(
		pg.id,
		pg.roomID,
		winnerID,
		pg.players,
		finishedAt,
	))

	return nil
}

// Helper methods

func (pg *PartyGame) hasPlayer(playerID UserID) bool {
	return pg.findPlayerIndex(playerID) != -1
}

func (pg *PartyGame) findPlayerIndex(playerID UserID) int {
	for i, p := range pg.players {
		if p.UserID().Equals(playerID) {
			return i
		}
	}
	return -1
}

func (pg *PartyGame) hasPlayerAnsweredCurrentQuestion(playerID UserID) bool {
	answers, exists := pg.questionAnswers[pg.currentQuestion]
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

func (pg *PartyGame) countCorrectAnswersForCurrentQuestion() int {
	answers, exists := pg.questionAnswers[pg.currentQuestion]
	if !exists {
		return 0
	}

	count := 0
	for _, ans := range answers {
		if ans.isCorrect {
			count++
		}
	}
	return count
}

func (pg *PartyGame) determineWinner() UserID {
	if len(pg.players) == 0 {
		return UserID{}
	}

	winner := pg.players[0]
	for _, p := range pg.players {
		if p.Score() > winner.Score() {
			winner = p
		}
	}

	return winner.UserID()
}

// Getters
func (pg *PartyGame) ID() GameID                { return pg.id }
func (pg *PartyGame) RoomID() RoomID            { return pg.roomID }
func (pg *PartyGame) QuestionIDs() []QuestionID {
	copy := make([]QuestionID, len(pg.questionIDs))
	for i, qid := range pg.questionIDs {
		copy[i] = qid
	}
	return copy
}
func (pg *PartyGame) Players() []PartyPlayer {
	copy := make([]PartyPlayer, len(pg.players))
	for i, p := range pg.players {
		copy[i] = p
	}
	return copy
}
func (pg *PartyGame) CurrentQuestion() int      { return pg.currentQuestion }
func (pg *PartyGame) Status() GameStatus        { return pg.status }
func (pg *PartyGame) StartedAt() int64          { return pg.startedAt }
func (pg *PartyGame) FinishedAt() int64         { return pg.finishedAt }
func (pg *PartyGame) IsFinished() bool          { return pg.status.IsTerminal() }

// Events returns collected domain events and clears them
func (pg *PartyGame) Events() []Event {
	events := pg.events
	pg.events = make([]Event, 0)
	return events
}

// ReconstructPartyGame reconstructs a PartyGame from persistence
func ReconstructPartyGame(
	id GameID,
	roomID RoomID,
	questionIDs []QuestionID,
	players []PartyPlayer,
	currentQuestion int,
	questionAnswers map[int][]QuestionAnswer,
	status GameStatus,
	startedAt int64,
	finishedAt int64,
) *PartyGame {
	return &PartyGame{
		id:              id,
		roomID:          roomID,
		questionIDs:     questionIDs,
		players:         players,
		currentQuestion: currentQuestion,
		questionAnswers: questionAnswers,
		status:          status,
		startedAt:       startedAt,
		finishedAt:      finishedAt,
		events:          make([]Event, 0),
	}
}
