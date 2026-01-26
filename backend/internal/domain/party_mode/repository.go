package party_mode

// RoomRepository defines the interface for party room persistence
type RoomRepository interface {
	// Save persists a party room
	Save(room *PartyRoom) error

	// FindByID retrieves a party room by ID
	FindByID(id RoomID) (*PartyRoom, error)

	// FindByCode retrieves a party room by code
	FindByCode(code RoomCode) (*PartyRoom, error)

	// FindActiveRooms retrieves all active (lobby status) rooms
	FindActiveRooms() ([]*PartyRoom, error)

	// Delete removes a party room
	Delete(id RoomID) error
}

// GameRepository defines the interface for party game persistence
type GameRepository interface {
	// Save persists a party game
	Save(game *PartyGame) error

	// FindByID retrieves a party game by ID
	FindByID(id GameID) (*PartyGame, error)

	// FindByRoomID retrieves a party game by room ID
	FindByRoomID(roomID RoomID) (*PartyGame, error)

	// Delete removes a party game
	Delete(id GameID) error
}
