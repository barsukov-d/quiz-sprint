package party_mode

// Event is the base interface for all party mode domain events
type Event interface {
	EventType() string
	OccurredAt() int64
}

// === PartyRoom Events ===

// RoomCreatedEvent fired when a party room is created
type RoomCreatedEvent struct {
	roomID     RoomID
	roomCode   RoomCode
	hostID     UserID
	settings   RoomSettings
	occurredAt int64
}

func NewRoomCreatedEvent(roomID RoomID, roomCode RoomCode, hostID UserID, settings RoomSettings, occurredAt int64) RoomCreatedEvent {
	return RoomCreatedEvent{roomID: roomID, roomCode: roomCode, hostID: hostID, settings: settings, occurredAt: occurredAt}
}

func (e RoomCreatedEvent) EventType() string       { return "room_created" }
func (e RoomCreatedEvent) OccurredAt() int64       { return e.occurredAt }
func (e RoomCreatedEvent) RoomID() RoomID          { return e.roomID }
func (e RoomCreatedEvent) RoomCode() RoomCode      { return e.roomCode }
func (e RoomCreatedEvent) HostID() UserID          { return e.hostID }
func (e RoomCreatedEvent) Settings() RoomSettings  { return e.settings }

// PlayerJoinedEvent fired when a player joins the room
type PlayerJoinedEvent struct {
	roomID     RoomID
	playerID   UserID
	username   string
	occurredAt int64
}

func NewPlayerJoinedEvent(roomID RoomID, playerID UserID, username string, occurredAt int64) PlayerJoinedEvent {
	return PlayerJoinedEvent{roomID: roomID, playerID: playerID, username: username, occurredAt: occurredAt}
}

func (e PlayerJoinedEvent) EventType() string { return "player_joined" }
func (e PlayerJoinedEvent) OccurredAt() int64 { return e.occurredAt }
func (e PlayerJoinedEvent) RoomID() RoomID    { return e.roomID }
func (e PlayerJoinedEvent) PlayerID() UserID  { return e.playerID }
func (e PlayerJoinedEvent) Username() string  { return e.username }

// PlayerLeftEvent fired when a player leaves the room
type PlayerLeftEvent struct {
	roomID     RoomID
	playerID   UserID
	occurredAt int64
}

func NewPlayerLeftEvent(roomID RoomID, playerID UserID, occurredAt int64) PlayerLeftEvent {
	return PlayerLeftEvent{roomID: roomID, playerID: playerID, occurredAt: occurredAt}
}

func (e PlayerLeftEvent) EventType() string { return "player_left" }
func (e PlayerLeftEvent) OccurredAt() int64 { return e.occurredAt }
func (e PlayerLeftEvent) RoomID() RoomID    { return e.roomID }
func (e PlayerLeftEvent) PlayerID() UserID  { return e.playerID }

// PlayerReadyEvent fired when a player marks ready/not ready
type PlayerReadyEvent struct {
	roomID     RoomID
	playerID   UserID
	isReady    bool
	occurredAt int64
}

func NewPlayerReadyEvent(roomID RoomID, playerID UserID, isReady bool, occurredAt int64) PlayerReadyEvent {
	return PlayerReadyEvent{roomID: roomID, playerID: playerID, isReady: isReady, occurredAt: occurredAt}
}

func (e PlayerReadyEvent) EventType() string { return "player_ready" }
func (e PlayerReadyEvent) OccurredAt() int64 { return e.occurredAt }
func (e PlayerReadyEvent) RoomID() RoomID    { return e.roomID }
func (e PlayerReadyEvent) PlayerID() UserID  { return e.playerID }
func (e PlayerReadyEvent) IsReady() bool     { return e.isReady }

// HostChangedEvent fired when host is transferred
type HostChangedEvent struct {
	roomID     RoomID
	oldHostID  UserID
	newHostID  UserID
	occurredAt int64
}

func NewHostChangedEvent(roomID RoomID, oldHostID UserID, newHostID UserID, occurredAt int64) HostChangedEvent {
	return HostChangedEvent{roomID: roomID, oldHostID: oldHostID, newHostID: newHostID, occurredAt: occurredAt}
}

func (e HostChangedEvent) EventType() string { return "host_changed" }
func (e HostChangedEvent) OccurredAt() int64 { return e.occurredAt }
func (e HostChangedEvent) RoomID() RoomID    { return e.roomID }
func (e HostChangedEvent) OldHostID() UserID { return e.oldHostID }
func (e HostChangedEvent) NewHostID() UserID { return e.newHostID }

// === PartyGame Events ===

// GameStartedEvent fired when party game starts
type GameStartedEvent struct {
	gameID      GameID
	roomID      RoomID
	playerIDs   []UserID
	questionIDs []QuestionID
	occurredAt  int64
}

func NewGameStartedEvent(gameID GameID, roomID RoomID, playerIDs []UserID, questionIDs []QuestionID, occurredAt int64) GameStartedEvent {
	return GameStartedEvent{gameID: gameID, roomID: roomID, playerIDs: playerIDs, questionIDs: questionIDs, occurredAt: occurredAt}
}

func (e GameStartedEvent) EventType() string         { return "game_started" }
func (e GameStartedEvent) OccurredAt() int64         { return e.occurredAt }
func (e GameStartedEvent) GameID() GameID            { return e.gameID }
func (e GameStartedEvent) RoomID() RoomID            { return e.roomID }
func (e GameStartedEvent) PlayerIDs() []UserID       { return e.playerIDs }
func (e GameStartedEvent) QuestionIDs() []QuestionID { return e.questionIDs }

// QuestionStartedEvent fired when a new question starts
type QuestionStartedEvent struct {
	gameID       GameID
	questionID   QuestionID
	questionNumber int
	occurredAt   int64
}

func NewQuestionStartedEvent(gameID GameID, questionID QuestionID, questionNumber int, occurredAt int64) QuestionStartedEvent {
	return QuestionStartedEvent{gameID: gameID, questionID: questionID, questionNumber: questionNumber, occurredAt: occurredAt}
}

func (e QuestionStartedEvent) EventType() string    { return "question_started" }
func (e QuestionStartedEvent) OccurredAt() int64    { return e.occurredAt }
func (e QuestionStartedEvent) GameID() GameID       { return e.gameID }
func (e QuestionStartedEvent) QuestionID() QuestionID { return e.questionID }
func (e QuestionStartedEvent) QuestionNumber() int  { return e.questionNumber }

// PartyPlayerAnsweredEvent fired when a player submits answer
type PartyPlayerAnsweredEvent struct {
	gameID       GameID
	playerID     UserID
	questionID   QuestionID
	answerID     AnswerID
	timeTaken    int64
	isCorrect    bool
	pointsEarned int
	position     int // Position in answering (1st, 2nd, etc.)
	occurredAt   int64
}

func NewPartyPlayerAnsweredEvent(gameID GameID, playerID UserID, questionID QuestionID, answerID AnswerID, timeTaken int64, isCorrect bool, pointsEarned int, position int, occurredAt int64) PartyPlayerAnsweredEvent {
	return PartyPlayerAnsweredEvent{gameID: gameID, playerID: playerID, questionID: questionID, answerID: answerID, timeTaken: timeTaken, isCorrect: isCorrect, pointsEarned: pointsEarned, position: position, occurredAt: occurredAt}
}

func (e PartyPlayerAnsweredEvent) EventType() string      { return "party_player_answered" }
func (e PartyPlayerAnsweredEvent) OccurredAt() int64      { return e.occurredAt }
func (e PartyPlayerAnsweredEvent) GameID() GameID         { return e.gameID }
func (e PartyPlayerAnsweredEvent) PlayerID() UserID       { return e.playerID }
func (e PartyPlayerAnsweredEvent) QuestionID() QuestionID { return e.questionID }
func (e PartyPlayerAnsweredEvent) AnswerID() AnswerID     { return e.answerID }
func (e PartyPlayerAnsweredEvent) TimeTaken() int64       { return e.timeTaken }
func (e PartyPlayerAnsweredEvent) IsCorrect() bool        { return e.isCorrect }
func (e PartyPlayerAnsweredEvent) PointsEarned() int      { return e.pointsEarned }
func (e PartyPlayerAnsweredEvent) Position() int          { return e.position }

// QuestionCompletedEvent fired when all players answered or timeout
type QuestionCompletedEvent struct {
	gameID         GameID
	questionID     QuestionID
	questionNumber int
	occurredAt     int64
}

func NewQuestionCompletedEvent(gameID GameID, questionID QuestionID, questionNumber int, occurredAt int64) QuestionCompletedEvent {
	return QuestionCompletedEvent{gameID: gameID, questionID: questionID, questionNumber: questionNumber, occurredAt: occurredAt}
}

func (e QuestionCompletedEvent) EventType() string    { return "question_completed" }
func (e QuestionCompletedEvent) OccurredAt() int64    { return e.occurredAt }
func (e QuestionCompletedEvent) GameID() GameID       { return e.gameID }
func (e QuestionCompletedEvent) QuestionID() QuestionID { return e.questionID }
func (e QuestionCompletedEvent) QuestionNumber() int  { return e.questionNumber }

// GameFinishedEvent fired when party game ends
type GameFinishedEvent struct {
	gameID     GameID
	roomID     RoomID
	winnerID   UserID
	players    []PartyPlayer
	occurredAt int64
}

func NewGameFinishedEvent(gameID GameID, roomID RoomID, winnerID UserID, players []PartyPlayer, occurredAt int64) GameFinishedEvent {
	return GameFinishedEvent{gameID: gameID, roomID: roomID, winnerID: winnerID, players: players, occurredAt: occurredAt}
}

func (e GameFinishedEvent) EventType() string        { return "game_finished" }
func (e GameFinishedEvent) OccurredAt() int64        { return e.occurredAt }
func (e GameFinishedEvent) GameID() GameID           { return e.gameID }
func (e GameFinishedEvent) RoomID() RoomID           { return e.roomID }
func (e GameFinishedEvent) WinnerID() UserID         { return e.winnerID }
func (e GameFinishedEvent) Players() []PartyPlayer   { return e.players }
