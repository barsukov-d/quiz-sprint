package quick_duel

import "database/sql"

// DuelGameRepository defines the interface for duel game persistence
type DuelGameRepository interface {
	// Save persists a duel game
	Save(game *DuelGame) error

	// FindByID retrieves a duel game by ID
	FindByID(id GameID) (*DuelGame, error)

	// FindActiveByPlayer retrieves the active duel game for a player
	// Returns nil if no active game found
	FindActiveByPlayer(playerID UserID) (*DuelGame, error)

	// FindByPlayerPaginated retrieves match history for a player
	FindByPlayerPaginated(playerID UserID, limit int, offset int, filter string) ([]*DuelGame, int, error)

	// Delete removes a duel game
	Delete(id GameID) error

	// AbandonStaleGames marks games stuck in waiting_start/in_progress as abandoned
	// if their started_at is before cutoffTime (unix timestamp).
	AbandonStaleGames(cutoffTime int64) (int, error)

	// FindRecentOpponents returns unique opponents from completed games, most recent first
	FindRecentOpponents(playerID UserID, limit int) ([]RecentOpponentEntry, error)
}

// RecentOpponentEntry represents a recent opponent with game count
type RecentOpponentEntry struct {
	OpponentID   UserID
	GamesCount   int
	LastPlayedAt int64
}

// MatchmakingQueue defines the interface for matchmaking queue operations
// (Usually implemented with Redis sorted sets)
type MatchmakingQueue interface {
	// AddToQueue adds a player to matchmaking queue
	// Priority is typically based on MMR rating
	AddToQueue(playerID UserID, mmr int, joinedAt int64) error

	// RemoveFromQueue removes a player from matchmaking queue
	RemoveFromQueue(playerID UserID) error

	// FindMatch finds a suitable opponent for a player
	// Returns opponent's UserID and MMR, or nil if no match found
	FindMatch(playerID UserID, mmr int, searchSeconds int) (*UserID, *int, error)

	// GetQueueLength returns number of players in queue
	GetQueueLength() (int, error)

	// IsPlayerInQueue checks if player is already in queue
	IsPlayerInQueue(playerID UserID) (bool, error)

	// GetPlayerQueueInfo returns player's queue info (joinedAt, mmr)
	GetPlayerQueueInfo(playerID UserID) (int64, int, error)

	// GetStaleQueueEntries returns IDs of players who joined before cutoffTime (unix sec)
	GetStaleQueueEntries(cutoffTime int64) ([]UserID, error)
}

// PlayerRatingRepository defines the interface for player rating persistence
type PlayerRatingRepository interface {
	// Save persists or updates a player rating
	Save(rating *PlayerRating) error

	// FindByPlayerID retrieves player rating
	FindByPlayerID(playerID UserID) (*PlayerRating, error)

	// FindOrCreate retrieves or creates player rating with initial values
	FindOrCreate(playerID UserID, seasonID string, createdAt int64) (*PlayerRating, error)

	// GetLeaderboard retrieves top players by MMR for a season
	GetLeaderboard(seasonID string, limit int, offset int) ([]*PlayerRating, error)

	// GetFriendsLeaderboard retrieves friends sorted by MMR
	GetFriendsLeaderboard(playerID UserID, friendIDs []UserID, limit int) ([]*PlayerRating, error)

	// GetPlayerRank retrieves player's rank position
	GetPlayerRank(playerID UserID, seasonID string) (int, error)

	// GetTotalPlayers returns total players in season
	GetTotalPlayers(seasonID string) (int, error)

	// FindAllBySeasonID retrieves all player ratings for a given season (used for reward distribution)
	FindAllBySeasonID(seasonID string) ([]*PlayerRating, error)

	// ResetAllRatingsForSeason applies soft reset to all players and moves them to newSeasonID.
	// resetFn computes the new MMR from the current MMR.
	ResetAllRatingsForSeason(newSeasonID string, resetFn func(currentMMR int) int) error
}

// ChallengeRepository defines the interface for duel challenge persistence
type ChallengeRepository interface {
	// Save persists a challenge
	Save(challenge *DuelChallenge) error

	// FindByID retrieves a challenge by ID
	FindByID(id ChallengeID) (*DuelChallenge, error)

	// FindByLink retrieves a challenge by full link
	FindByLink(link string) (*DuelChallenge, error)

	// FindByLinkCode retrieves a challenge by link code (e.g., "duel_abc12345")
	FindByLinkCode(code string) (*DuelChallenge, error)

	// FindPendingForPlayer retrieves pending challenges for a player
	FindPendingForPlayer(playerID UserID) ([]*DuelChallenge, error)

	// FindPendingByChallenger retrieves pending challenges sent by a player
	FindPendingByChallenger(playerID UserID) ([]*DuelChallenge, error)

	// FindAcceptedWaitingForPlayer retrieves challenges where the player is the invitee
	// and the challenge status is accepted_waiting_inviter (game not yet started).
	FindAcceptedWaitingForPlayer(playerID UserID) ([]*DuelChallenge, error)

	// Delete removes a challenge
	Delete(id ChallengeID) error

	// FindPendingExpired returns pending challenges whose expires_at <= currentTime
	FindPendingExpired(currentTime int64) ([]*DuelChallenge, error)

	// FindWaitingExpired returns accepted_waiting_inviter challenges
	// whose responded_at + AcceptedWaitingExpirySeconds <= currentTime
	FindWaitingExpired(currentTime int64) ([]*DuelChallenge, error)

	// DeleteExpired removes all expired pending challenges
	DeleteExpired(currentTime int64) error

	// FindPendingExpiredWithMessageID returns pending challenges that expired and have a telegram message to edit
	FindPendingExpiredWithMessageID(currentTime int64) ([]*DuelChallenge, error)

	// DeleteHardExpired deletes expired/declined challenges older than the given unix timestamp
	DeleteHardExpired(olderThan int64) error

	// FindExpiredForPlayer returns expired challenges visible to this player (as inviter or invitee)
	// within the last 24 hours (still shown in UI before auto-deletion)
	FindExpiredForPlayer(playerID UserID) ([]*DuelChallenge, error)

	// FindByIDForUpdate retrieves a challenge by ID with a row-level lock (SELECT FOR UPDATE).
	FindByIDForUpdate(tx *sql.Tx, id ChallengeID) (*DuelChallenge, error)

	// FindByLinkCodeForUpdate retrieves a pending challenge by link code with a row-level lock.
	FindByLinkCodeForUpdate(tx *sql.Tx, code string) (*DuelChallenge, error)

	// SaveInTx persists a challenge within an existing transaction.
	SaveInTx(tx *sql.Tx, challenge *DuelChallenge) error
}

// ReferralRepository defines the interface for referral persistence
type ReferralRepository interface {
	// Save persists a referral
	Save(referral *Referral) error

	// FindByID retrieves a referral by ID
	FindByID(id ReferralID) (*Referral, error)

	// FindByInviterAndInvitee retrieves a referral by inviter and invitee
	FindByInviterAndInvitee(inviterID UserID, inviteeID UserID) (*Referral, error)

	// FindByInvitee retrieves the referral for an invitee (they can only have one)
	FindByInvitee(inviteeID UserID) (*Referral, error)

	// FindByInviter retrieves all referrals made by an inviter
	FindByInviter(inviterID UserID) ([]*Referral, error)

	// CountByInviter returns total referrals by an inviter
	CountByInviter(inviterID UserID) (int, error)

	// CountActiveByInviter returns referrals where invitee has played at least 5 duels
	CountActiveByInviter(inviterID UserID) (int, error)

	// GetReferralLeaderboard returns top inviters by referral count
	GetReferralLeaderboard(limit int) ([]ReferralLeaderboardEntry, error)

	// GetPlayerReferralRank returns player's referral leaderboard position
	GetPlayerReferralRank(inviterID UserID) (int, error)
}

// ReferralLeaderboardEntry represents a row in referral leaderboard
type ReferralLeaderboardEntry struct {
	PlayerID       UserID
	Username       string
	TotalReferrals int
	ActiveReferrals int
}

// SeasonRepository defines the interface for season management
type SeasonRepository interface {
	// GetCurrentSeason returns the current active season ID
	GetCurrentSeason() (string, error)

	// CreateSeason creates a new season
	CreateSeason(seasonID string, startsAt int64, endsAt int64) error

	// GetSeasonInfo returns season start/end times
	GetSeasonInfo(seasonID string) (int64, int64, error)
}
