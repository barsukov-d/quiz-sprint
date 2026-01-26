package quick_duel

// Event is the base interface for all quick duel domain events
type Event interface {
	EventType() string
	OccurredAt() int64
}

// DuelGameCreatedEvent fired when a duel game is created (matchmaking successful)
type DuelGameCreatedEvent struct {
	gameID      GameID
	player1     DuelPlayer
	player2     DuelPlayer
	questionIDs []QuestionID
	occurredAt  int64
}

func NewDuelGameCreatedEvent(
	gameID GameID,
	player1 DuelPlayer,
	player2 DuelPlayer,
	questionIDs []QuestionID,
	occurredAt int64,
) DuelGameCreatedEvent {
	return DuelGameCreatedEvent{
		gameID:      gameID,
		player1:     player1,
		player2:     player2,
		questionIDs: questionIDs,
		occurredAt:  occurredAt,
	}
}

func (e DuelGameCreatedEvent) EventType() string         { return "duel_game_created" }
func (e DuelGameCreatedEvent) OccurredAt() int64         { return e.occurredAt }
func (e DuelGameCreatedEvent) GameID() GameID            { return e.gameID }
func (e DuelGameCreatedEvent) Player1() DuelPlayer       { return e.player1 }
func (e DuelGameCreatedEvent) Player2() DuelPlayer       { return e.player2 }
func (e DuelGameCreatedEvent) QuestionIDs() []QuestionID { return e.questionIDs }

// DuelGameStartedEvent fired when game actually starts (both players ready)
type DuelGameStartedEvent struct {
	gameID     GameID
	player1ID  UserID
	player2ID  UserID
	occurredAt int64
}

func NewDuelGameStartedEvent(
	gameID GameID,
	player1ID UserID,
	player2ID UserID,
	occurredAt int64,
) DuelGameStartedEvent {
	return DuelGameStartedEvent{
		gameID:     gameID,
		player1ID:  player1ID,
		player2ID:  player2ID,
		occurredAt: occurredAt,
	}
}

func (e DuelGameStartedEvent) EventType() string { return "duel_game_started" }
func (e DuelGameStartedEvent) OccurredAt() int64 { return e.occurredAt }
func (e DuelGameStartedEvent) GameID() GameID    { return e.gameID }
func (e DuelGameStartedEvent) Player1ID() UserID { return e.player1ID }
func (e DuelGameStartedEvent) Player2ID() UserID { return e.player2ID }

// RoundStartedEvent fired when a new round (question) starts
type RoundStartedEvent struct {
	gameID      GameID
	roundNumber int
	questionID  QuestionID
	occurredAt  int64
}

func NewRoundStartedEvent(
	gameID GameID,
	roundNumber int,
	questionID QuestionID,
	occurredAt int64,
) RoundStartedEvent {
	return RoundStartedEvent{
		gameID:      gameID,
		roundNumber: roundNumber,
		questionID:  questionID,
		occurredAt:  occurredAt,
	}
}

func (e RoundStartedEvent) EventType() string    { return "round_started" }
func (e RoundStartedEvent) OccurredAt() int64    { return e.occurredAt }
func (e RoundStartedEvent) GameID() GameID       { return e.gameID }
func (e RoundStartedEvent) RoundNumber() int     { return e.roundNumber }
func (e RoundStartedEvent) QuestionID() QuestionID { return e.questionID }

// PlayerAnsweredEvent fired when a player submits an answer
type PlayerAnsweredEvent struct {
	gameID      GameID
	playerID    UserID
	questionID  QuestionID
	answerID    AnswerID
	timeTaken   int64 // milliseconds
	isCorrect   bool
	pointsEarned int
	occurredAt  int64
}

func NewPlayerAnsweredEvent(
	gameID GameID,
	playerID UserID,
	questionID QuestionID,
	answerID AnswerID,
	timeTaken int64,
	isCorrect bool,
	pointsEarned int,
	occurredAt int64,
) PlayerAnsweredEvent {
	return PlayerAnsweredEvent{
		gameID:       gameID,
		playerID:     playerID,
		questionID:   questionID,
		answerID:     answerID,
		timeTaken:    timeTaken,
		isCorrect:    isCorrect,
		pointsEarned: pointsEarned,
		occurredAt:   occurredAt,
	}
}

func (e PlayerAnsweredEvent) EventType() string      { return "player_answered" }
func (e PlayerAnsweredEvent) OccurredAt() int64      { return e.occurredAt }
func (e PlayerAnsweredEvent) GameID() GameID         { return e.gameID }
func (e PlayerAnsweredEvent) PlayerID() UserID       { return e.playerID }
func (e PlayerAnsweredEvent) QuestionID() QuestionID { return e.questionID }
func (e PlayerAnsweredEvent) AnswerID() AnswerID     { return e.answerID }
func (e PlayerAnsweredEvent) TimeTaken() int64       { return e.timeTaken }
func (e PlayerAnsweredEvent) IsCorrect() bool        { return e.isCorrect }
func (e PlayerAnsweredEvent) PointsEarned() int      { return e.pointsEarned }

// RoundCompletedEvent fired when both players answered (or timeout)
type RoundCompletedEvent struct {
	gameID         GameID
	roundNumber    int
	player1Score   int
	player2Score   int
	player1Answered bool
	player2Answered bool
	occurredAt     int64
}

func NewRoundCompletedEvent(
	gameID GameID,
	roundNumber int,
	player1Score int,
	player2Score int,
	player1Answered bool,
	player2Answered bool,
	occurredAt int64,
) RoundCompletedEvent {
	return RoundCompletedEvent{
		gameID:          gameID,
		roundNumber:     roundNumber,
		player1Score:    player1Score,
		player2Score:    player2Score,
		player1Answered: player1Answered,
		player2Answered: player2Answered,
		occurredAt:      occurredAt,
	}
}

func (e RoundCompletedEvent) EventType() string      { return "round_completed" }
func (e RoundCompletedEvent) OccurredAt() int64      { return e.occurredAt }
func (e RoundCompletedEvent) GameID() GameID         { return e.gameID }
func (e RoundCompletedEvent) RoundNumber() int       { return e.roundNumber }
func (e RoundCompletedEvent) Player1Score() int      { return e.player1Score }
func (e RoundCompletedEvent) Player2Score() int      { return e.player2Score }
func (e RoundCompletedEvent) Player1Answered() bool  { return e.player1Answered }
func (e RoundCompletedEvent) Player2Answered() bool  { return e.player2Answered }

// DuelGameFinishedEvent fired when game ends (all rounds completed)
type DuelGameFinishedEvent struct {
	gameID        GameID
	winnerID      *UserID // nil if draw
	player1       DuelPlayer
	player2       DuelPlayer
	player1NewElo EloRating
	player2NewElo EloRating
	occurredAt    int64
}

func NewDuelGameFinishedEvent(
	gameID GameID,
	winnerID *UserID,
	player1 DuelPlayer,
	player2 DuelPlayer,
	player1NewElo EloRating,
	player2NewElo EloRating,
	occurredAt int64,
) DuelGameFinishedEvent {
	return DuelGameFinishedEvent{
		gameID:        gameID,
		winnerID:      winnerID,
		player1:       player1,
		player2:       player2,
		player1NewElo: player1NewElo,
		player2NewElo: player2NewElo,
		occurredAt:    occurredAt,
	}
}

func (e DuelGameFinishedEvent) EventType() string       { return "duel_game_finished" }
func (e DuelGameFinishedEvent) OccurredAt() int64       { return e.occurredAt }
func (e DuelGameFinishedEvent) GameID() GameID          { return e.gameID }
func (e DuelGameFinishedEvent) WinnerID() *UserID       { return e.winnerID }
func (e DuelGameFinishedEvent) Player1() DuelPlayer     { return e.player1 }
func (e DuelGameFinishedEvent) Player2() DuelPlayer     { return e.player2 }
func (e DuelGameFinishedEvent) Player1NewElo() EloRating { return e.player1NewElo }
func (e DuelGameFinishedEvent) Player2NewElo() EloRating { return e.player2NewElo }

// PlayerDisconnectedEvent fired when a player loses connection
type PlayerDisconnectedEvent struct {
	gameID     GameID
	playerID   UserID
	occurredAt int64
}

func NewPlayerDisconnectedEvent(
	gameID GameID,
	playerID UserID,
	occurredAt int64,
) PlayerDisconnectedEvent {
	return PlayerDisconnectedEvent{
		gameID:     gameID,
		playerID:   playerID,
		occurredAt: occurredAt,
	}
}

func (e PlayerDisconnectedEvent) EventType() string { return "player_disconnected" }
func (e PlayerDisconnectedEvent) OccurredAt() int64 { return e.occurredAt }
func (e PlayerDisconnectedEvent) GameID() GameID    { return e.gameID }
func (e PlayerDisconnectedEvent) PlayerID() UserID  { return e.playerID }

// PlayerReconnectedEvent fired when a player reconnects
type PlayerReconnectedEvent struct {
	gameID     GameID
	playerID   UserID
	occurredAt int64
}

func NewPlayerReconnectedEvent(
	gameID GameID,
	playerID UserID,
	occurredAt int64,
) PlayerReconnectedEvent {
	return PlayerReconnectedEvent{
		gameID:     gameID,
		playerID:   playerID,
		occurredAt: occurredAt,
	}
}

func (e PlayerReconnectedEvent) EventType() string { return "player_reconnected" }
func (e PlayerReconnectedEvent) OccurredAt() int64 { return e.occurredAt }
func (e PlayerReconnectedEvent) GameID() GameID    { return e.gameID }
func (e PlayerReconnectedEvent) PlayerID() UserID  { return e.playerID }
