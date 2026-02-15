package quick_duel

import "math"

// MMR calculation constants
const (
	InitialMMR        = 1000
	MinMMR            = 0
	MaxMMR            = 9999
	KFactor           = 32  // ELO K-factor
	MinMMRChange      = 10  // Minimum change per match
	DemotionProtection = 3  // Games protected from demotion at new rank
)

// PlayerRating represents a player's competitive ranking (aggregate root)
type PlayerRating struct {
	playerID      UserID
	mmr           int
	league        League
	division      Division
	peakMMR       int
	peakLeague    League
	peakDivision  Division
	gamesAtRank   int      // Games played at current rank (for demotion protection)
	seasonID      string
	seasonWins    int
	seasonLosses  int
	updatedAt     int64

	events []Event
}

// NewPlayerRating creates a new player rating with initial values
func NewPlayerRating(playerID UserID, seasonID string, createdAt int64) *PlayerRating {
	leagueInfo := GetLeagueFromMMR(InitialMMR)

	return &PlayerRating{
		playerID:     playerID,
		mmr:          InitialMMR,
		league:       leagueInfo.League(),
		division:     leagueInfo.Division(),
		peakMMR:      InitialMMR,
		peakLeague:   leagueInfo.League(),
		peakDivision: leagueInfo.Division(),
		gamesAtRank:  0,
		seasonID:     seasonID,
		seasonWins:   0,
		seasonLosses: 0,
		updatedAt:    createdAt,
		events:       make([]Event, 0),
	}
}

// ReconstructPlayerRating reconstructs from persistence
func ReconstructPlayerRating(
	playerID UserID,
	mmr int,
	league League,
	division Division,
	peakMMR int,
	peakLeague League,
	peakDivision Division,
	gamesAtRank int,
	seasonID string,
	seasonWins int,
	seasonLosses int,
	updatedAt int64,
) *PlayerRating {
	return &PlayerRating{
		playerID:     playerID,
		mmr:          mmr,
		league:       league,
		division:     division,
		peakMMR:      peakMMR,
		peakLeague:   peakLeague,
		peakDivision: peakDivision,
		gamesAtRank:  gamesAtRank,
		seasonID:     seasonID,
		seasonWins:   seasonWins,
		seasonLosses: seasonLosses,
		updatedAt:    updatedAt,
		events:       make([]Event, 0),
	}
}

// GameResult represents the result of a duel game for rating calculation
type GameResult struct {
	Won         bool
	OpponentMMR int
	GameTime    int64
}

// ApplyGameResult updates rating based on game result
func (pr *PlayerRating) ApplyGameResult(result GameResult) {
	oldLeague := pr.league
	oldDivision := pr.division

	// Calculate expected score (probability of winning)
	expectedScore := 1.0 / (1.0 + math.Pow(10, float64(result.OpponentMMR-pr.mmr)/400.0))

	// Actual score
	var actualScore float64
	if result.Won {
		actualScore = 1.0
		pr.seasonWins++
	} else {
		actualScore = 0.0
		pr.seasonLosses++
	}

	// Calculate rating change
	delta := float64(KFactor) * (actualScore - expectedScore)

	// Enforce minimum change
	if delta > 0 && delta < float64(MinMMRChange) {
		delta = float64(MinMMRChange)
	} else if delta < 0 && delta > -float64(MinMMRChange) {
		delta = -float64(MinMMRChange)
	}

	// Apply change
	pr.mmr += int(math.Round(delta))

	// Enforce bounds
	if pr.mmr < MinMMR {
		pr.mmr = MinMMR
	}
	if pr.mmr > MaxMMR {
		pr.mmr = MaxMMR
	}

	// Update league/division
	newLeagueInfo := GetLeagueFromMMR(pr.mmr)
	newLeague := newLeagueInfo.League()
	newDivision := newLeagueInfo.Division()

	// Check demotion protection
	if pr.shouldPreventDemotion(oldLeague, oldDivision, newLeague, newDivision) {
		// Keep at current rank floor
		pr.mmr = pr.getRankFloorMMR(oldLeague, oldDivision)
		newLeagueInfo = GetLeagueFromMMR(pr.mmr)
		newLeague = newLeagueInfo.League()
		newDivision = newLeagueInfo.Division()
	}

	// Check if rank changed
	rankChanged := oldLeague != newLeague || oldDivision != newDivision

	if rankChanged {
		pr.gamesAtRank = 0 // Reset games counter at new rank

		// Publish promotion/demotion event
		if newLeague > oldLeague || (newLeague == oldLeague && newDivision < oldDivision) {
			// Promoted (lower division number = higher rank)
			pr.events = append(pr.events, NewPlayerPromotedEvent(
				pr.playerID,
				oldLeague, oldDivision,
				newLeague, newDivision,
				pr.mmr,
				result.GameTime,
			))
		} else {
			// Demoted
			pr.events = append(pr.events, NewPlayerDemotedEvent(
				pr.playerID,
				oldLeague, oldDivision,
				newLeague, newDivision,
				pr.mmr,
				result.GameTime,
			))
		}
	} else {
		pr.gamesAtRank++
	}

	pr.league = newLeague
	pr.division = newDivision
	pr.updatedAt = result.GameTime

	// Update peak if new high
	if pr.mmr > pr.peakMMR {
		pr.peakMMR = pr.mmr
		pr.peakLeague = pr.league
		pr.peakDivision = pr.division
	}
}

// shouldPreventDemotion checks if demotion protection applies
func (pr *PlayerRating) shouldPreventDemotion(
	oldLeague League, oldDivision Division,
	newLeague League, newDivision Division,
) bool {
	// No demotion happening
	if newLeague > oldLeague || (newLeague == oldLeague && newDivision <= oldDivision) {
		return false
	}

	// Protection only applies if player is at rank floor and played < 3 games
	if pr.gamesAtRank < DemotionProtection {
		return true
	}

	return false
}

// getRankFloorMMR returns minimum MMR for a rank
func (pr *PlayerRating) getRankFloorMMR(league League, division Division) int {
	baseMMR := league.MinMMR()

	// Calculate division offset (IV=0, III=125, II=250, I=375)
	divisionOffset := (int(DivisionIV) - int(division)) * DivisionSpan

	return baseMMR + divisionOffset
}

// SeasonReset applies seasonal MMR reset
func (pr *PlayerRating) SeasonReset(newSeasonID string, resetTime int64) {
	// Soft reset formula: newMMR = 1000 + (currentMMR - 1000) * 0.5
	// Minimum: 500 MMR
	newMMR := 1000 + int(float64(pr.mmr-1000)*0.5)
	if newMMR < 500 {
		newMMR = 500
	}

	oldSeasonID := pr.seasonID

	pr.mmr = newMMR
	pr.seasonID = newSeasonID
	pr.seasonWins = 0
	pr.seasonLosses = 0
	pr.gamesAtRank = 0
	pr.updatedAt = resetTime

	// Update league/division
	leagueInfo := GetLeagueFromMMR(pr.mmr)
	pr.league = leagueInfo.League()
	pr.division = leagueInfo.Division()

	// Publish season reset event
	pr.events = append(pr.events, NewSeasonResetEvent(
		pr.playerID,
		oldSeasonID,
		newSeasonID,
		pr.mmr,
		pr.league,
		pr.division,
		resetTime,
	))
}

// GetLeagueLabel returns formatted league label
func (pr *PlayerRating) GetLeagueLabel() string {
	return LeagueInfo{league: pr.league, division: pr.division}.FullLabel()
}

// WinRate returns win rate as percentage
func (pr *PlayerRating) WinRate() float64 {
	total := pr.seasonWins + pr.seasonLosses
	if total == 0 {
		return 0
	}
	return float64(pr.seasonWins) / float64(total) * 100
}

// CanDemote returns true if player can be demoted (no protection)
func (pr *PlayerRating) CanDemote() bool {
	return pr.gamesAtRank >= DemotionProtection
}

// Getters
func (pr *PlayerRating) PlayerID() UserID       { return pr.playerID }
func (pr *PlayerRating) MMR() int               { return pr.mmr }
func (pr *PlayerRating) League() League         { return pr.league }
func (pr *PlayerRating) Division() Division     { return pr.division }
func (pr *PlayerRating) PeakMMR() int           { return pr.peakMMR }
func (pr *PlayerRating) PeakLeague() League     { return pr.peakLeague }
func (pr *PlayerRating) PeakDivision() Division { return pr.peakDivision }
func (pr *PlayerRating) GamesAtRank() int       { return pr.gamesAtRank }
func (pr *PlayerRating) SeasonID() string       { return pr.seasonID }
func (pr *PlayerRating) SeasonWins() int        { return pr.seasonWins }
func (pr *PlayerRating) SeasonLosses() int      { return pr.seasonLosses }
func (pr *PlayerRating) UpdatedAt() int64       { return pr.updatedAt }

// Events returns collected domain events and clears them
func (pr *PlayerRating) Events() []Event {
	events := pr.events
	pr.events = make([]Event, 0)
	return events
}
