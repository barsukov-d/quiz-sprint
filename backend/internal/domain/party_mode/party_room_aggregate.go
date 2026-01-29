package party_mode

// PartyRoom is the aggregate root for party room (lobby before game starts)
type PartyRoom struct {
	id         RoomID
	code       RoomCode
	name       string
	hostID     UserID
	settings   RoomSettings
	players    []RoomPlayer
	status     RoomStatus
	createdAt  int64
	expiresAt  int64 // Auto-close after 1 hour

	// Domain events collected during operations
	events []Event
}

// NewPartyRoom creates a new party room
func NewPartyRoom(
	hostID UserID,
	hostUsername string,
	name string,
	settings RoomSettings,
	createdAt int64,
) (*PartyRoom, error) {
	// Validate
	if hostID.IsZero() {
		return nil, ErrInvalidRoomID
	}

	if err := settings.Validate(); err != nil {
		return nil, err
	}

	// Create
	roomID := NewRoomID()
	roomCode := GenerateRoomCode()
	expiresAt := createdAt + (60 * 60) // 1 hour

	// Host is the first player
	host := NewRoomPlayer(hostID, hostUsername, true, createdAt)
	players := []RoomPlayer{host}

	room := &PartyRoom{
		id:        roomID,
		code:      roomCode,
		name:      name,
		hostID:    hostID,
		settings:  settings,
		players:   players,
		status:    RoomStatusLobby,
		createdAt: createdAt,
		expiresAt: expiresAt,
		events:    make([]Event, 0),
	}

	// Publish RoomCreated event
	room.events = append(room.events, NewRoomCreatedEvent(
		roomID,
		roomCode,
		hostID,
		settings,
		createdAt,
	))

	return room, nil
}

// JoinPlayer adds a player to the room
func (pr *PartyRoom) JoinPlayer(playerID UserID, username string, joinedAt int64) error {
	// Validate
	if pr.status != RoomStatusLobby {
		return ErrRoomAlreadyStarted
	}

	// Check if already in room
	if pr.hasPlayer(playerID) {
		return ErrPlayerAlreadyInRoom
	}

	// Check if room is full
	if len(pr.players) >= pr.settings.MaxPlayers() {
		return ErrRoomFull
	}

	// Add player
	player := NewRoomPlayer(playerID, username, false, joinedAt)
	pr.players = append(pr.players, player)

	// Publish PlayerJoined event
	pr.events = append(pr.events, NewPlayerJoinedEvent(
		pr.id,
		playerID,
		username,
		joinedAt,
	))

	return nil
}

// RemovePlayer removes a player from the room
func (pr *PartyRoom) RemovePlayer(playerID UserID, leftAt int64) error {
	// Find and remove player
	playerIndex := pr.findPlayerIndex(playerID)
	if playerIndex == -1 {
		return ErrPlayerNotFound
	}

	player := pr.players[playerIndex]
	pr.players = append(pr.players[:playerIndex], pr.players[playerIndex+1:]...)

	// Publish PlayerLeft event
	pr.events = append(pr.events, NewPlayerLeftEvent(
		pr.id,
		playerID,
		leftAt,
	))

	// If player was host, transfer to another player
	if player.IsHost() && len(pr.players) > 0 {
		newHost := pr.players[0]
		oldHostID := pr.hostID
		pr.hostID = newHost.UserID()
		pr.players[0] = newHost.SetHost(true)

		pr.events = append(pr.events, NewHostChangedEvent(
			pr.id,
			oldHostID,
			pr.hostID,
			leftAt,
		))
	}

	// Close room if empty
	if len(pr.players) == 0 {
		// Validate state transition
		if !pr.status.CanTransitionTo(RoomStatusClosed) {
			return ErrInvalidRoomStatus
		}

		pr.status = RoomStatusClosed
	}

	return nil
}

// SetPlayerReady sets player ready status
func (pr *PartyRoom) SetPlayerReady(playerID UserID, ready bool, readyAt int64) error {
	if pr.status != RoomStatusLobby {
		return ErrRoomAlreadyStarted
	}

	playerIndex := pr.findPlayerIndex(playerID)
	if playerIndex == -1 {
		return ErrPlayerNotFound
	}

	pr.players[playerIndex] = pr.players[playerIndex].SetReady(ready)

	// Publish PlayerReady event
	pr.events = append(pr.events, NewPlayerReadyEvent(
		pr.id,
		playerID,
		ready,
		readyAt,
	))

	return nil
}

// CanStartGame checks if game can be started
func (pr *PartyRoom) CanStartGame(requesterID UserID) error {
	// Only host can start
	if !pr.hostID.Equals(requesterID) {
		return ErrOnlyHostCanStart
	}

	// Need at least 2 players
	if len(pr.players) < 2 {
		return ErrNotEnoughPlayers
	}

	// All players must be ready
	for _, player := range pr.players {
		if !player.IsReady() && !player.IsHost() {
			return ErrNotAllPlayersReady
		}
	}

	return nil
}

// StartGame transitions room to playing status
func (pr *PartyRoom) StartGame() error {
	// Validate state transition
	if !pr.status.CanTransitionTo(RoomStatusPlaying) {
		return ErrInvalidRoomStatus
	}

	pr.status = RoomStatusPlaying
	return nil
}

// Helper methods

func (pr *PartyRoom) hasPlayer(playerID UserID) bool {
	return pr.findPlayerIndex(playerID) != -1
}

func (pr *PartyRoom) findPlayerIndex(playerID UserID) int {
	for i, p := range pr.players {
		if p.UserID().Equals(playerID) {
			return i
		}
	}
	return -1
}

// Getters
func (pr *PartyRoom) ID() RoomID              { return pr.id }
func (pr *PartyRoom) Code() RoomCode          { return pr.code }
func (pr *PartyRoom) Name() string            { return pr.name }
func (pr *PartyRoom) HostID() UserID          { return pr.hostID }
func (pr *PartyRoom) Settings() RoomSettings  { return pr.settings }
func (pr *PartyRoom) Players() []RoomPlayer   {
	// Return copy
	copy := make([]RoomPlayer, len(pr.players))
	for i, p := range pr.players {
		copy[i] = p
	}
	return copy
}
func (pr *PartyRoom) PlayerCount() int        { return len(pr.players) }
func (pr *PartyRoom) Status() RoomStatus      { return pr.status }
func (pr *PartyRoom) CreatedAt() int64        { return pr.createdAt }
func (pr *PartyRoom) ExpiresAt() int64        { return pr.expiresAt }

// Events returns collected domain events and clears them
func (pr *PartyRoom) Events() []Event {
	events := pr.events
	pr.events = make([]Event, 0)
	return events
}

// ReconstructPartyRoom reconstructs a PartyRoom from persistence
func ReconstructPartyRoom(
	id RoomID,
	code RoomCode,
	name string,
	hostID UserID,
	settings RoomSettings,
	players []RoomPlayer,
	status RoomStatus,
	createdAt int64,
	expiresAt int64,
) *PartyRoom {
	return &PartyRoom{
		id:        id,
		code:      code,
		name:      name,
		hostID:    hostID,
		settings:  settings,
		players:   players,
		status:    status,
		createdAt: createdAt,
		expiresAt: expiresAt,
		events:    make([]Event, 0),
	}
}
