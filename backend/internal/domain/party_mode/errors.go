package party_mode

import "errors"

// Domain errors for party mode
var (
	// Room errors
	ErrInvalidRoomID       = errors.New("invalid room ID")
	ErrRoomNotFound        = errors.New("room not found")
	ErrRoomFull            = errors.New("room is full")
	ErrRoomAlreadyStarted  = errors.New("room already started")
	ErrRoomClosed          = errors.New("room is closed")
	ErrInvalidRoomCode     = errors.New("invalid room code")
	ErrInvalidRoomSettings = errors.New("invalid room settings")
	ErrInvalidRoomStatus   = errors.New("invalid room status transition")

	// Player errors
	ErrPlayerNotFound      = errors.New("player not in room")
	ErrPlayerAlreadyInRoom = errors.New("player already in room")
	ErrNotEnoughPlayers    = errors.New("not enough players to start (minimum 2)")
	ErrNotAllPlayersReady  = errors.New("not all players ready")
	ErrOnlyHostCanStart    = errors.New("only host can start game")
	ErrCannotKickHost      = errors.New("cannot kick host")

	// Game errors
	ErrInvalidGameID        = errors.New("invalid game ID")
	ErrGameNotFound         = errors.New("party game not found")
	ErrGameAlreadyFinished  = errors.New("party game already finished")
	ErrGameNotActive        = errors.New("party game is not active")
	ErrPlayerAlreadyAnswered = errors.New("player already answered this question")
	ErrAllQuestionsAnswered = errors.New("all questions already answered")
	ErrInvalidGameStatus    = errors.New("invalid game status transition")
)
