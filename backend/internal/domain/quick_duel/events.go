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

// PlayerPromotedEvent fired when player's rank increases
type PlayerPromotedEvent struct {
	playerID     UserID
	fromLeague   League
	fromDivision Division
	toLeague     League
	toDivision   Division
	newMMR       int
	occurredAt   int64
}

func NewPlayerPromotedEvent(
	playerID UserID,
	fromLeague League,
	fromDivision Division,
	toLeague League,
	toDivision Division,
	newMMR int,
	occurredAt int64,
) PlayerPromotedEvent {
	return PlayerPromotedEvent{
		playerID:     playerID,
		fromLeague:   fromLeague,
		fromDivision: fromDivision,
		toLeague:     toLeague,
		toDivision:   toDivision,
		newMMR:       newMMR,
		occurredAt:   occurredAt,
	}
}

func (e PlayerPromotedEvent) EventType() string     { return "player_promoted" }
func (e PlayerPromotedEvent) OccurredAt() int64     { return e.occurredAt }
func (e PlayerPromotedEvent) PlayerID() UserID      { return e.playerID }
func (e PlayerPromotedEvent) FromLeague() League    { return e.fromLeague }
func (e PlayerPromotedEvent) FromDivision() Division { return e.fromDivision }
func (e PlayerPromotedEvent) ToLeague() League      { return e.toLeague }
func (e PlayerPromotedEvent) ToDivision() Division  { return e.toDivision }
func (e PlayerPromotedEvent) NewMMR() int           { return e.newMMR }

// PlayerDemotedEvent fired when player's rank decreases
type PlayerDemotedEvent struct {
	playerID     UserID
	fromLeague   League
	fromDivision Division
	toLeague     League
	toDivision   Division
	newMMR       int
	occurredAt   int64
}

func NewPlayerDemotedEvent(
	playerID UserID,
	fromLeague League,
	fromDivision Division,
	toLeague League,
	toDivision Division,
	newMMR int,
	occurredAt int64,
) PlayerDemotedEvent {
	return PlayerDemotedEvent{
		playerID:     playerID,
		fromLeague:   fromLeague,
		fromDivision: fromDivision,
		toLeague:     toLeague,
		toDivision:   toDivision,
		newMMR:       newMMR,
		occurredAt:   occurredAt,
	}
}

func (e PlayerDemotedEvent) EventType() string     { return "player_demoted" }
func (e PlayerDemotedEvent) OccurredAt() int64     { return e.occurredAt }
func (e PlayerDemotedEvent) PlayerID() UserID      { return e.playerID }
func (e PlayerDemotedEvent) FromLeague() League    { return e.fromLeague }
func (e PlayerDemotedEvent) FromDivision() Division { return e.fromDivision }
func (e PlayerDemotedEvent) ToLeague() League      { return e.toLeague }
func (e PlayerDemotedEvent) ToDivision() Division  { return e.toDivision }
func (e PlayerDemotedEvent) NewMMR() int           { return e.newMMR }

// SeasonResetEvent fired when MMR is reset for new season
type SeasonResetEvent struct {
	playerID    UserID
	oldSeasonID string
	newSeasonID string
	newMMR      int
	newLeague   League
	newDivision Division
	occurredAt  int64
}

func NewSeasonResetEvent(
	playerID UserID,
	oldSeasonID string,
	newSeasonID string,
	newMMR int,
	newLeague League,
	newDivision Division,
	occurredAt int64,
) SeasonResetEvent {
	return SeasonResetEvent{
		playerID:    playerID,
		oldSeasonID: oldSeasonID,
		newSeasonID: newSeasonID,
		newMMR:      newMMR,
		newLeague:   newLeague,
		newDivision: newDivision,
		occurredAt:  occurredAt,
	}
}

func (e SeasonResetEvent) EventType() string   { return "season_reset" }
func (e SeasonResetEvent) OccurredAt() int64   { return e.occurredAt }
func (e SeasonResetEvent) PlayerID() UserID    { return e.playerID }
func (e SeasonResetEvent) OldSeasonID() string { return e.oldSeasonID }
func (e SeasonResetEvent) NewSeasonID() string { return e.newSeasonID }
func (e SeasonResetEvent) NewMMR() int         { return e.newMMR }
func (e SeasonResetEvent) NewLeague() League   { return e.newLeague }
func (e SeasonResetEvent) NewDivision() Division { return e.newDivision }

// Challenge Events

// ChallengeCreatedEvent fired when a challenge is created
type ChallengeCreatedEvent struct {
	challengeID   ChallengeID
	challengerID  UserID
	challengedID  *UserID
	challengeType ChallengeType
	expiresAt     int64
	occurredAt    int64
}

func NewChallengeCreatedEvent(
	challengeID ChallengeID,
	challengerID UserID,
	challengedID *UserID,
	challengeType ChallengeType,
	expiresAt int64,
	occurredAt int64,
) ChallengeCreatedEvent {
	return ChallengeCreatedEvent{
		challengeID:   challengeID,
		challengerID:  challengerID,
		challengedID:  challengedID,
		challengeType: challengeType,
		expiresAt:     expiresAt,
		occurredAt:    occurredAt,
	}
}

func (e ChallengeCreatedEvent) EventType() string           { return "challenge_created" }
func (e ChallengeCreatedEvent) OccurredAt() int64           { return e.occurredAt }
func (e ChallengeCreatedEvent) ChallengeID() ChallengeID    { return e.challengeID }
func (e ChallengeCreatedEvent) ChallengerID() UserID        { return e.challengerID }
func (e ChallengeCreatedEvent) ChallengedID() *UserID       { return e.challengedID }
func (e ChallengeCreatedEvent) ChallengeType() ChallengeType { return e.challengeType }
func (e ChallengeCreatedEvent) ExpiresAt() int64            { return e.expiresAt }

// ChallengeAcceptedEvent fired when a challenge is accepted
type ChallengeAcceptedEvent struct {
	challengeID  ChallengeID
	challengerID UserID
	accepterID   UserID
	occurredAt   int64
}

func NewChallengeAcceptedEvent(
	challengeID ChallengeID,
	challengerID UserID,
	accepterID UserID,
	occurredAt int64,
) ChallengeAcceptedEvent {
	return ChallengeAcceptedEvent{
		challengeID:  challengeID,
		challengerID: challengerID,
		accepterID:   accepterID,
		occurredAt:   occurredAt,
	}
}

func (e ChallengeAcceptedEvent) EventType() string        { return "challenge_accepted" }
func (e ChallengeAcceptedEvent) OccurredAt() int64        { return e.occurredAt }
func (e ChallengeAcceptedEvent) ChallengeID() ChallengeID { return e.challengeID }
func (e ChallengeAcceptedEvent) ChallengerID() UserID     { return e.challengerID }
func (e ChallengeAcceptedEvent) AccepterID() UserID       { return e.accepterID }

// ChallengeDeclinedEvent fired when a challenge is declined
type ChallengeDeclinedEvent struct {
	challengeID  ChallengeID
	challengerID UserID
	declinerID   UserID
	occurredAt   int64
}

func NewChallengeDeclinedEvent(
	challengeID ChallengeID,
	challengerID UserID,
	declinerID UserID,
	occurredAt int64,
) ChallengeDeclinedEvent {
	return ChallengeDeclinedEvent{
		challengeID:  challengeID,
		challengerID: challengerID,
		declinerID:   declinerID,
		occurredAt:   occurredAt,
	}
}

func (e ChallengeDeclinedEvent) EventType() string        { return "challenge_declined" }
func (e ChallengeDeclinedEvent) OccurredAt() int64        { return e.occurredAt }
func (e ChallengeDeclinedEvent) ChallengeID() ChallengeID { return e.challengeID }
func (e ChallengeDeclinedEvent) ChallengerID() UserID     { return e.challengerID }
func (e ChallengeDeclinedEvent) DeclinerID() UserID       { return e.declinerID }

// ChallengeExpiredEvent fired when a challenge expires
type ChallengeExpiredEvent struct {
	challengeID  ChallengeID
	challengerID UserID
	challengedID *UserID
	occurredAt   int64
}

func NewChallengeExpiredEvent(
	challengeID ChallengeID,
	challengerID UserID,
	challengedID *UserID,
	occurredAt int64,
) ChallengeExpiredEvent {
	return ChallengeExpiredEvent{
		challengeID:  challengeID,
		challengerID: challengerID,
		challengedID: challengedID,
		occurredAt:   occurredAt,
	}
}

func (e ChallengeExpiredEvent) EventType() string        { return "challenge_expired" }
func (e ChallengeExpiredEvent) OccurredAt() int64        { return e.occurredAt }
func (e ChallengeExpiredEvent) ChallengeID() ChallengeID { return e.challengeID }
func (e ChallengeExpiredEvent) ChallengerID() UserID     { return e.challengerID }
func (e ChallengeExpiredEvent) ChallengedID() *UserID    { return e.challengedID }

// Referral Events

// ReferralCreatedEvent fired when a referral is created
type ReferralCreatedEvent struct {
	referralID ChallengeID
	inviterID  UserID
	inviteeID  UserID
	occurredAt int64
}

func NewReferralCreatedEvent(
	referralID ChallengeID,
	inviterID UserID,
	inviteeID UserID,
	occurredAt int64,
) ReferralCreatedEvent {
	return ReferralCreatedEvent{
		referralID: referralID,
		inviterID:  inviterID,
		inviteeID:  inviteeID,
		occurredAt: occurredAt,
	}
}

func (e ReferralCreatedEvent) EventType() string        { return "referral_created" }
func (e ReferralCreatedEvent) OccurredAt() int64        { return e.occurredAt }
func (e ReferralCreatedEvent) ReferralID() ChallengeID  { return e.referralID }
func (e ReferralCreatedEvent) InviterID() UserID        { return e.inviterID }
func (e ReferralCreatedEvent) InviteeID() UserID        { return e.inviteeID }

// ReferralMilestoneEvent fired when a referral milestone is achieved
type ReferralMilestoneEvent struct {
	referralID ChallengeID
	inviterID  UserID
	inviteeID  UserID
	milestone  string
	occurredAt int64
}

func NewReferralMilestoneEvent(
	referralID ChallengeID,
	inviterID UserID,
	inviteeID UserID,
	milestone string,
	occurredAt int64,
) ReferralMilestoneEvent {
	return ReferralMilestoneEvent{
		referralID: referralID,
		inviterID:  inviterID,
		inviteeID:  inviteeID,
		milestone:  milestone,
		occurredAt: occurredAt,
	}
}

func (e ReferralMilestoneEvent) EventType() string       { return "referral_milestone" }
func (e ReferralMilestoneEvent) OccurredAt() int64       { return e.occurredAt }
func (e ReferralMilestoneEvent) ReferralID() ChallengeID { return e.referralID }
func (e ReferralMilestoneEvent) InviterID() UserID       { return e.inviterID }
func (e ReferralMilestoneEvent) InviteeID() UserID       { return e.inviteeID }
func (e ReferralMilestoneEvent) Milestone() string       { return e.milestone }
