package quick_duel

// ========================================
// Common DTOs
// ========================================

// DuelPlayerDTO represents a player in a duel
type DuelPlayerDTO struct {
	ID         string  `json:"id"`
	Username   string  `json:"username"`
	Avatar     string  `json:"avatar,omitempty"`
	MMR        int     `json:"mmr"`
	League     string  `json:"league"`
	Division   int     `json:"division"`
	LeagueIcon string  `json:"leagueIcon"`
	Score      int     `json:"score"`
	Connected  bool    `json:"connected"`
}

// DuelMatchDTO represents a duel match
type DuelMatchDTO struct {
	ID            string          `json:"id"`
	Status        string          `json:"status"`
	Player1       DuelPlayerDTO   `json:"player1"`
	Player2       DuelPlayerDTO   `json:"player2"`
	CurrentRound  int             `json:"currentRound"`
	TotalRounds   int             `json:"totalRounds"`
	StartedAt     int64           `json:"startedAt"`
	FinishedAt    int64           `json:"finishedAt,omitempty"`
	WinnerID      *string         `json:"winnerId,omitempty"`
	IsFriendMatch bool            `json:"isFriendMatch"`
}

// DuelQuestionDTO represents a question in a duel (no IsCorrect field!)
type DuelQuestionDTO struct {
	ID           string            `json:"id"`
	QuestionNum  int               `json:"questionNumber"`
	Text         string            `json:"text"`
	Answers      []DuelAnswerDTO   `json:"answers"`
	TimeLimit    int               `json:"timeLimit"`
	ServerTime   int64             `json:"serverTime"`
}

// DuelAnswerDTO represents an answer option
type DuelAnswerDTO struct {
	ID   string `json:"id"`
	Text string `json:"text"`
}

// PlayerRatingDTO represents a player's competitive ranking
type PlayerRatingDTO struct {
	PlayerID     string  `json:"playerId"`
	MMR          int     `json:"mmr"`
	League       string  `json:"league"`
	Division     int     `json:"division"`
	LeagueLabel  string  `json:"leagueLabel"`
	LeagueIcon   string  `json:"leagueIcon"`
	PeakMMR      int     `json:"peakMmr"`
	PeakLeague   string  `json:"peakLeague"`
	SeasonWins   int     `json:"seasonWins"`
	SeasonLosses int     `json:"seasonLosses"`
	WinRate      float64 `json:"winRate"`
	GamesAtRank  int     `json:"gamesAtRank"`
	CanDemote    bool    `json:"canDemote"`
}

// LeaderboardEntryDTO represents a leaderboard entry
type LeaderboardEntryDTO struct {
	Rank       int     `json:"rank"`
	PlayerID   string  `json:"playerId"`
	Username   string  `json:"username"`
	Avatar     string  `json:"avatar,omitempty"`
	MMR        int     `json:"mmr"`
	League     string  `json:"league"`
	LeagueIcon string  `json:"leagueIcon"`
	Wins       int     `json:"wins"`
	Losses     int     `json:"losses"`
	WinRate    float64 `json:"winRate"`
}

// ChallengeDTO represents a duel challenge
type ChallengeDTO struct {
	ID            string  `json:"id"`
	ChallengerID  string  `json:"challengerId"`
	ChallengedID  *string `json:"challengedId,omitempty"`
	Type          string  `json:"type"`
	Status        string  `json:"status"`
	ChallengeLink string  `json:"challengeLink,omitempty"`
	ExpiresAt     int64   `json:"expiresAt"`
	ExpiresIn     int     `json:"expiresIn"`
	CreatedAt     int64   `json:"createdAt"`
}

// MatchHistoryEntryDTO represents a match in history
type MatchHistoryEntryDTO struct {
	MatchID       string `json:"matchId"`
	Opponent      string `json:"opponent"`
	OpponentMMR   int    `json:"opponentMmr"`
	Result        string `json:"result"`
	PlayerScore   int    `json:"playerScore"`
	OpponentScore int    `json:"opponentScore"`
	MMRChange     int    `json:"mmrChange"`
	IsFriendMatch bool   `json:"isFriendMatch"`
	CompletedAt   int64  `json:"completedAt"`
}

// ReferralDTO represents a referral
type ReferralDTO struct {
	ID                string   `json:"id"`
	InviteeID         string   `json:"inviteeId"`
	InviteeUsername   string   `json:"inviteeUsername"`
	MilestonesReached []string `json:"milestonesReached"`
	PendingRewards    []string `json:"pendingRewards"`
	CreatedAt         int64    `json:"createdAt"`
}

// ReferralRewardDTO represents a referral reward
type ReferralRewardDTO struct {
	Tickets int    `json:"tickets"`
	Coins   int    `json:"coins"`
	Badge   string `json:"badge,omitempty"`
	Avatar  string `json:"avatar,omitempty"`
	Title   string `json:"title,omitempty"`
}

// FriendDTO represents an online friend
type FriendDTO struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Avatar   string `json:"avatar,omitempty"`
	MMR      int    `json:"mmr"`
	League   string `json:"league"`
	IsOnline bool   `json:"isOnline"`
	InMatch  bool   `json:"inMatch"`
}

// ========================================
// GetDuelStatus Use Case
// ========================================

type GetDuelStatusInput struct {
	PlayerID string `json:"playerId"`
}

type GetDuelStatusOutput struct {
	HasActiveDuel     bool              `json:"hasActiveDuel"`
	ActiveMatchID     *string           `json:"activeMatchId,omitempty"`
	Player            PlayerRatingDTO   `json:"player"`
	Tickets           int               `json:"tickets"`
	FriendsOnline     []FriendDTO       `json:"friendsOnline"`
	PendingChallenges []ChallengeDTO    `json:"pendingChallenges"`
	SeasonID          string            `json:"seasonId"`
	SeasonEndsAt      int64             `json:"seasonEndsAt"`
}

// ========================================
// JoinQueue Use Case
// ========================================

type JoinQueueInput struct {
	PlayerID string `json:"playerId"`
}

type JoinQueueOutput struct {
	QueueID       string `json:"queueId"`
	Status        string `json:"status"`
	EstimatedWait int    `json:"estimatedWait"`
	MMRRange      string `json:"mmrRange"`
	Position      int    `json:"position"`
}

// ========================================
// LeaveQueue Use Case
// ========================================

type LeaveQueueInput struct {
	PlayerID string `json:"playerId"`
}

type LeaveQueueOutput struct {
	Success        bool `json:"success"`
	TicketRefunded bool `json:"ticketRefunded"`
	NewTicketCount int  `json:"newTicketCount"`
}

// ========================================
// SendChallenge Use Case
// ========================================

type SendChallengeInput struct {
	PlayerID string `json:"playerId"`
	FriendID string `json:"friendId"`
}

type SendChallengeOutput struct {
	ChallengeID    string `json:"challengeId"`
	Status         string `json:"status"`
	ExpiresIn      int    `json:"expiresIn"`
	TicketConsumed bool   `json:"ticketConsumed"`
}

// ========================================
// RespondChallenge Use Case
// ========================================

type RespondChallengeInput struct {
	PlayerID    string `json:"playerId"`
	ChallengeID string `json:"challengeId"`
	Action      string `json:"action"` // "accept" or "decline"
}

type RespondChallengeOutput struct {
	Success        bool    `json:"success"`
	MatchID        *string `json:"matchId,omitempty"`
	TicketConsumed bool    `json:"ticketConsumed"`
	StartsIn       *int    `json:"startsIn,omitempty"`
}

// ========================================
// CreateChallengeLink Use Case
// ========================================

type CreateChallengeLinkInput struct {
	PlayerID string `json:"playerId"`
}

type CreateChallengeLinkOutput struct {
	ChallengeLink string `json:"challengeLink"`
	ExpiresAt     int64  `json:"expiresAt"`
	ShareText     string `json:"shareText"`
}

// ========================================
// GetMatchResult Use Case
// ========================================

type GetMatchResultInput struct {
	PlayerID string `json:"playerId"`
	MatchID  string `json:"matchId"`
}

type MatchQuestionResultDTO struct {
	QuestionNum       int    `json:"questionNumber"`
	QuestionText      string `json:"questionText"`
	PlayerAnswerID    string `json:"playerAnswerId"`
	PlayerCorrect     bool   `json:"playerCorrect"`
	PlayerTime        int64  `json:"playerTime"`
	OpponentAnswerID  string `json:"opponentAnswerId"`
	OpponentCorrect   bool   `json:"opponentCorrect"`
	OpponentTime      int64  `json:"opponentTime"`
	CorrectAnswerID   string `json:"correctAnswerId"`
}

type GetMatchResultOutput struct {
	MatchID         string                   `json:"matchId"`
	Result          string                   `json:"result"` // "win", "loss", "draw"
	PlayerScore     int                      `json:"playerScore"`
	OpponentScore   int                      `json:"opponentScore"`
	MMRChange       int                      `json:"mmrChange"`
	NewMMR          int                      `json:"newMmr"`
	RankChange      *string                  `json:"rankChange,omitempty"` // "promoted", "demoted", null
	NewLeague       string                   `json:"newLeague"`
	NewDivision     int                      `json:"newDivision"`
	Opponent        DuelPlayerDTO            `json:"opponent"`
	Questions       []MatchQuestionResultDTO `json:"questions"`
	CanRematch      bool                     `json:"canRematch"`
	RematchExpiresIn *int                    `json:"rematchExpiresIn,omitempty"`
}

// ========================================
// RequestRematch Use Case
// ========================================

type RequestRematchInput struct {
	PlayerID string `json:"playerId"`
	MatchID  string `json:"matchId"`
}

type RequestRematchOutput struct {
	RematchID string `json:"rematchId"`
	ExpiresIn int    `json:"expiresIn"`
	Status    string `json:"status"`
}

// ========================================
// GetMatchHistory Use Case
// ========================================

type GetMatchHistoryInput struct {
	PlayerID string `json:"playerId"`
	Limit    int    `json:"limit"`
	Offset   int    `json:"offset"`
	Filter   string `json:"filter"` // "all", "friends", "wins", "losses"
}

type GetMatchHistoryOutput struct {
	Matches []MatchHistoryEntryDTO `json:"matches"`
	Total   int                    `json:"total"`
	HasMore bool                   `json:"hasMore"`
}

// ========================================
// GetLeaderboard Use Case
// ========================================

type GetLeaderboardInput struct {
	PlayerID string `json:"playerId"`
	Type     string `json:"type"`  // "seasonal", "friends", "referrals"
	Limit    int    `json:"limit"`
}

type GetLeaderboardOutput struct {
	Type       string                `json:"type"`
	SeasonID   string                `json:"seasonId,omitempty"`
	EndsAt     int64                 `json:"endsAt,omitempty"`
	Entries    []LeaderboardEntryDTO `json:"entries"`
	PlayerRank int                   `json:"playerRank"`
}

// ========================================
// GetReferrals Use Case
// ========================================

type GetReferralsInput struct {
	PlayerID string `json:"playerId"`
}

type GetReferralsOutput struct {
	ReferralLink          string        `json:"referralLink"`
	TotalReferrals        int           `json:"totalReferrals"`
	ActiveReferrals       int           `json:"activeReferrals"`
	PendingRewards        []string      `json:"pendingRewards"`
	ReferralLeaderboardRank int         `json:"referralLeaderboardRank"`
	Referrals             []ReferralDTO `json:"referrals"`
}

// ========================================
// ClaimReferralReward Use Case
// ========================================

type ClaimReferralRewardInput struct {
	PlayerID  string `json:"playerId"`
	FriendID  string `json:"friendId"`
	Milestone string `json:"milestone"`
}

type ClaimReferralRewardOutput struct {
	Success          bool              `json:"success"`
	Rewards          ReferralRewardDTO `json:"rewards"`
	NewTicketBalance int               `json:"newTicketBalance"`
	NewCoinBalance   int               `json:"newCoinBalance"`
}

// ========================================
// SubmitDuelAnswer Use Case
// ========================================

type SubmitDuelAnswerInput struct {
	PlayerID   string `json:"playerId"`
	MatchID    string `json:"matchId"`
	QuestionID string `json:"questionId"`
	AnswerID   string `json:"answerId"`
	ClientTime int64  `json:"clientTime"` // For latency compensation
}

type SubmitDuelAnswerOutput struct {
	Received        bool              `json:"received"`
	IsCorrect       bool              `json:"isCorrect"`
	PointsEarned    int               `json:"pointsEarned"`
	PlayerScore     int               `json:"playerScore"`
	OpponentScore   int               `json:"opponentScore"`
	OpponentAnswered bool             `json:"opponentAnswered"`
	RoundComplete   bool              `json:"roundComplete"`
	IsGameFinished  bool              `json:"isGameFinished"`
	GameResult      *GetMatchResultOutput `json:"gameResult,omitempty"`
}
