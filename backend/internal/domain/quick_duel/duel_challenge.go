package quick_duel

import (
	"github.com/google/uuid"
)

// Challenge expiry constants
const (
	DirectChallengeExpirySeconds = 60       // 60 seconds for online friend
	PushChallengeExpirySeconds   = 300      // 5 minutes if push notification sent
	LinkChallengeExpirySeconds   = 86400    // 24 hours for link-based challenges
)

// ChallengeStatus represents the current state of a challenge
type ChallengeStatus string

const (
	ChallengeStatusPending  ChallengeStatus = "pending"
	ChallengeStatusAccepted ChallengeStatus = "accepted"
	ChallengeStatusDeclined ChallengeStatus = "declined"
	ChallengeStatusExpired  ChallengeStatus = "expired"
)

// ChallengeType represents the type of challenge
type ChallengeType string

const (
	ChallengeTypeDirect ChallengeType = "direct" // Direct challenge to specific friend
	ChallengeTypeLink   ChallengeType = "link"   // Shareable link challenge
)

// ChallengeID uniquely identifies a challenge
type ChallengeID struct {
	value string
}

func NewChallengeID() ChallengeID {
	return ChallengeID{value: uuid.New().String()}
}

func NewChallengeIDFromString(value string) ChallengeID {
	return ChallengeID{value: value}
}

func (id ChallengeID) String() string { return id.value }
func (id ChallengeID) IsZero() bool   { return id.value == "" }

// DuelChallenge represents a friend challenge request (aggregate root)
type DuelChallenge struct {
	id            ChallengeID
	challengerID  UserID          // Who sent the challenge
	challengedID  *UserID         // Who is challenged (nil for link-based)
	challengeType ChallengeType
	status        ChallengeStatus
	challengeLink string          // For link-based challenges
	expiresAt     int64           // Unix timestamp
	createdAt     int64
	respondedAt   int64           // When accepted/declined (0 if pending)
	matchID       *GameID         // Set when match starts

	events []Event
}

// NewDirectChallenge creates a direct challenge to a specific friend
func NewDirectChallenge(
	challengerID UserID,
	challengedID UserID,
	createdAt int64,
) (*DuelChallenge, error) {
	if challengerID.IsZero() || challengedID.IsZero() {
		return nil, ErrInvalidChallengeID
	}

	if challengerID.Equals(challengedID) {
		return nil, ErrCannotChallengeSelf
	}

	challenge := &DuelChallenge{
		id:            NewChallengeID(),
		challengerID:  challengerID,
		challengedID:  &challengedID,
		challengeType: ChallengeTypeDirect,
		status:        ChallengeStatusPending,
		challengeLink: "",
		expiresAt:     createdAt + DirectChallengeExpirySeconds,
		createdAt:     createdAt,
		respondedAt:   0,
		matchID:       nil,
		events:        make([]Event, 0),
	}

	challenge.events = append(challenge.events, NewChallengeCreatedEvent(
		challenge.id,
		challengerID,
		&challengedID,
		ChallengeTypeDirect,
		challenge.expiresAt,
		createdAt,
	))

	return challenge, nil
}

// NewLinkChallenge creates a shareable link challenge
func NewLinkChallenge(
	challengerID UserID,
	createdAt int64,
) (*DuelChallenge, error) {
	if challengerID.IsZero() {
		return nil, ErrInvalidChallengeID
	}

	challengeID := NewChallengeID()
	link := "t.me/quizsprintbot?start=duel_" + challengeID.String()[:8]

	challenge := &DuelChallenge{
		id:            challengeID,
		challengerID:  challengerID,
		challengedID:  nil,
		challengeType: ChallengeTypeLink,
		status:        ChallengeStatusPending,
		challengeLink: link,
		expiresAt:     createdAt + LinkChallengeExpirySeconds,
		createdAt:     createdAt,
		respondedAt:   0,
		matchID:       nil,
		events:        make([]Event, 0),
	}

	challenge.events = append(challenge.events, NewChallengeCreatedEvent(
		challenge.id,
		challengerID,
		nil,
		ChallengeTypeLink,
		challenge.expiresAt,
		createdAt,
	))

	return challenge, nil
}

// ReconstructDuelChallenge reconstructs from persistence
func ReconstructDuelChallenge(
	id ChallengeID,
	challengerID UserID,
	challengedID *UserID,
	challengeType ChallengeType,
	status ChallengeStatus,
	challengeLink string,
	expiresAt int64,
	createdAt int64,
	respondedAt int64,
	matchID *GameID,
) *DuelChallenge {
	return &DuelChallenge{
		id:            id,
		challengerID:  challengerID,
		challengedID:  challengedID,
		challengeType: challengeType,
		status:        status,
		challengeLink: challengeLink,
		expiresAt:     expiresAt,
		createdAt:     createdAt,
		respondedAt:   respondedAt,
		matchID:       matchID,
		events:        make([]Event, 0),
	}
}

// Accept accepts the challenge
func (dc *DuelChallenge) Accept(accepterID UserID, acceptedAt int64) error {
	if dc.status != ChallengeStatusPending {
		return ErrChallengeNotPending
	}

	if dc.IsExpired(acceptedAt) {
		dc.status = ChallengeStatusExpired
		return ErrChallengeExpired
	}

	// For direct challenges, verify accepter is the challenged player
	if dc.challengeType == ChallengeTypeDirect {
		if dc.challengedID == nil || !dc.challengedID.Equals(accepterID) {
			return ErrNotChallengedPlayer
		}
	} else {
		// For link challenges, set the challenged player
		if dc.challengerID.Equals(accepterID) {
			return ErrCannotChallengeSelf
		}
		dc.challengedID = &accepterID
	}

	dc.status = ChallengeStatusAccepted
	dc.respondedAt = acceptedAt

	dc.events = append(dc.events, NewChallengeAcceptedEvent(
		dc.id,
		dc.challengerID,
		accepterID,
		acceptedAt,
	))

	return nil
}

// Decline declines the challenge
func (dc *DuelChallenge) Decline(declinerID UserID, declinedAt int64) error {
	if dc.status != ChallengeStatusPending {
		return ErrChallengeNotPending
	}

	// For direct challenges, verify decliner is the challenged player
	if dc.challengeType == ChallengeTypeDirect {
		if dc.challengedID == nil || !dc.challengedID.Equals(declinerID) {
			return ErrNotChallengedPlayer
		}
	}

	dc.status = ChallengeStatusDeclined
	dc.respondedAt = declinedAt

	dc.events = append(dc.events, NewChallengeDeclinedEvent(
		dc.id,
		dc.challengerID,
		declinerID,
		declinedAt,
	))

	return nil
}

// SetMatchID sets the match ID when match starts
func (dc *DuelChallenge) SetMatchID(matchID GameID) {
	dc.matchID = &matchID
}

// Expire marks the challenge as expired
func (dc *DuelChallenge) Expire(expiredAt int64) error {
	if dc.status != ChallengeStatusPending {
		return ErrChallengeNotPending
	}

	dc.status = ChallengeStatusExpired
	dc.respondedAt = expiredAt

	dc.events = append(dc.events, NewChallengeExpiredEvent(
		dc.id,
		dc.challengerID,
		dc.challengedID,
		expiredAt,
	))

	return nil
}

// IsExpired checks if challenge has expired
func (dc *DuelChallenge) IsExpired(currentTime int64) bool {
	return currentTime >= dc.expiresAt
}

// IsPending checks if challenge is pending
func (dc *DuelChallenge) IsPending() bool {
	return dc.status == ChallengeStatusPending
}

// Getters
func (dc *DuelChallenge) ID() ChallengeID            { return dc.id }
func (dc *DuelChallenge) ChallengerID() UserID       { return dc.challengerID }
func (dc *DuelChallenge) ChallengedID() *UserID      { return dc.challengedID }
func (dc *DuelChallenge) Type() ChallengeType        { return dc.challengeType }
func (dc *DuelChallenge) Status() ChallengeStatus    { return dc.status }
func (dc *DuelChallenge) ChallengeLink() string      { return dc.challengeLink }
func (dc *DuelChallenge) ExpiresAt() int64           { return dc.expiresAt }
func (dc *DuelChallenge) CreatedAt() int64           { return dc.createdAt }
func (dc *DuelChallenge) RespondedAt() int64         { return dc.respondedAt }
func (dc *DuelChallenge) MatchID() *GameID           { return dc.matchID }

// Events returns collected domain events and clears them
func (dc *DuelChallenge) Events() []Event {
	events := dc.events
	dc.events = make([]Event, 0)
	return events
}
